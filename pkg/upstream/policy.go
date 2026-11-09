package upstream

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

type Semver struct {
	Major, Minor, Patch int
}

func (s Semver) String() string      { return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch) }
func (s Semver) MinorString() string { return fmt.Sprintf("%d.%d", s.Major, s.Minor) }

func (s Semver) Less(o Semver) bool {
	if s.Major != o.Major {
		return s.Major < o.Major
	}
	if s.Minor != o.Minor {
		return s.Minor < o.Minor
	}
	return s.Patch < o.Patch
}

// Candidate is one upstream release a lib could track.
type Candidate struct {
	Tag     string
	Version Semver
	// Rank breaks ties between candidates with the same Version (e.g. several
	// chart releases shipping one appVersion); higher wins.
	Rank Semver
}

var semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

// parseMinor parses "X.Y" as X.Y.0.
func parseMinor(s string) (Semver, bool) {
	return ParseSemver(s + ".0")
}

// ParseSemver parses a plain "1.2.3" or "v1.2.3"; pre-releases are rejected.
func ParseSemver(s string) (Semver, bool) {
	return parseWith(semverRe, s)
}

func compileTagPattern(p string) (*regexp.Regexp, error) {
	re, err := regexp.Compile(p)
	if err != nil {
		return nil, fmt.Errorf("source.tagPattern: %w", err)
	}
	if re.NumSubexp() < 3 {
		return nil, fmt.Errorf("source.tagPattern must capture major, minor and patch")
	}
	return re, nil
}

func parseWith(re *regexp.Regexp, s string) (Semver, bool) {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return Semver{}, false
	}
	var v [3]int
	for i := range v {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return Semver{}, false
		}
		v[i] = n
	}
	return Semver{v[0], v[1], v[2]}, true
}

// Select applies policy to candidates: the latest patch of the newest
// policy.Minors minors, plus any policy.Keep minors. Newest first.
func Select(cands []Candidate, policy Policy) []Version {
	since, hasSince := parseMinor(policy.Since)

	best := map[string]Candidate{}
	for _, c := range cands {
		if hasSince && c.Version.Less(since) {
			continue
		}
		key := c.Version.MinorString()
		cur, ok := best[key]
		if !ok || cur.Version.Less(c.Version) || (cur.Version == c.Version && cur.Rank.Less(c.Rank)) {
			best[key] = c
		}
	}

	minors := make([]Candidate, 0, len(best))
	for _, c := range best {
		minors = append(minors, c)
	}
	sort.Slice(minors, func(i, j int) bool { return minors[j].Version.Less(minors[i].Version) })

	keep := map[string]bool{}
	for _, k := range policy.Keep {
		keep[k] = true
	}

	var out []Version
	for i, c := range minors {
		if i < policy.Minors || keep[c.Version.MinorString()] {
			out = append(out, Version{
				Output:  c.Version.MinorString(),
				Tag:     c.Tag,
				Version: c.Version.String(),
			})
		}
	}
	return out
}

// Change describes how a lib's tracked versions move.
type Change struct {
	Output string `json:"output"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
}

type Diff struct {
	Added   []Change `json:"added,omitempty"`
	Updated []Change `json:"updated,omitempty"`
	Removed []Change `json:"removed,omitempty"`
}

func (d Diff) Empty() bool {
	return len(d.Added) == 0 && len(d.Updated) == 0 && len(d.Removed) == 0
}

func Compare(current, desired []Version) Diff {
	cur := map[string]Version{}
	for _, v := range current {
		cur[v.Output] = v
	}
	des := map[string]bool{}

	var d Diff
	for _, v := range desired {
		des[v.Output] = true
		old, ok := cur[v.Output]
		switch {
		case !ok:
			d.Added = append(d.Added, Change{Output: v.Output, To: v.Version})
		case old.Tag != v.Tag || old.Version != v.Version:
			d.Updated = append(d.Updated, Change{Output: v.Output, From: old.Version, To: v.Version})
		}
	}
	for _, v := range current {
		if !des[v.Output] {
			d.Removed = append(d.Removed, Change{Output: v.Output, From: v.Version})
		}
	}
	return d
}
