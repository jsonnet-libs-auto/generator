package upstream

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Remote is the network side of discovery; tests replace it.
type Remote interface {
	// Tags returns the tag names of a git repo.
	Tags(ctx context.Context, repo string) ([]string, error)
	// File returns a file from a GitHub repo at ref.
	File(ctx context.Context, repo, ref, path string) ([]byte, error)
}

// Candidates lists upstream releases matching the lib's source; for
// helm-chart sources only as many as policy needs.
func Candidates(ctx context.Context, r Remote, src Source, policy Policy) ([]Candidate, error) {
	re, err := compileTagPattern(src.TagPattern)
	if err != nil {
		return nil, err
	}
	tags, err := r.Tags(ctx, src.Repo)
	if err != nil {
		return nil, err
	}

	var matched []Candidate
	for _, t := range tags {
		if v, ok := parseWith(re, t); ok {
			matched = append(matched, Candidate{Tag: t, Version: v, Rank: v})
		}
	}

	switch src.Kind {
	case KindGitTags:
		return matched, nil
	case KindHelmChart:
		return chartCandidates(ctx, r, src, policy, matched)
	}
	return nil, fmt.Errorf("unknown source.kind %q", src.Kind)
}

// chartCandidates maps chart tags to their appVersion, newest chart minor
// first, fetching only the newest patch of each chart minor and stopping once
// policy is satisfied.
func chartCandidates(ctx context.Context, r Remote, src Source, policy Policy, charts []Candidate) ([]Candidate, error) {
	newest := map[string]Candidate{}
	for _, c := range charts {
		k := c.Version.MinorString()
		if cur, ok := newest[k]; !ok || cur.Version.Less(c.Version) {
			newest[k] = c
		}
	}
	ordered := make([]Candidate, 0, len(newest))
	for _, c := range newest {
		ordered = append(ordered, c)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[j].Version.Less(ordered[i].Version) })

	wantKeep := map[string]bool{}
	for _, k := range policy.Keep {
		wantKeep[k] = true
	}
	appMinors := map[string]bool{}

	var out []Candidate
	for _, c := range ordered {
		if len(appMinors) >= policy.Minors+1 && len(wantKeep) == 0 {
			// One extra app minor seen means every newer minor is complete.
			break
		}
		data, err := r.File(ctx, src.Repo, c.Tag, strings.TrimSuffix(src.ChartPath, "/")+"/Chart.yaml")
		if err != nil {
			return nil, err
		}
		var chart struct {
			AppVersion string `yaml:"appVersion"`
		}
		if err := yaml.Unmarshal(data, &chart); err != nil {
			return nil, fmt.Errorf("%s Chart.yaml at %s: %w", src.Repo, c.Tag, err)
		}
		app, ok := ParseSemver(chart.AppVersion)
		if !ok {
			continue
		}
		appMinors[app.MinorString()] = true
		delete(wantKeep, app.MinorString())
		out = append(out, Candidate{Tag: c.Tag, Version: app, Rank: c.Version})
	}
	return out, nil
}

// GitRemote talks to real remotes: git for tags, raw.githubusercontent.com
// for files. Neither needs a token or counts against the GitHub API limit.
type GitRemote struct {
	HTTP *http.Client
}

func (g GitRemote) Tags(ctx context.Context, repo string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", repo)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-remote %s: %w: %s", repo, err, strings.TrimSpace(stderr.String()))
	}
	var tags []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 {
			tags = append(tags, strings.TrimPrefix(fields[1], "refs/tags/"))
		}
	}
	return tags, sc.Err()
}

func (g GitRemote) File(ctx context.Context, repo, ref, path string) ([]byte, error) {
	slug, ok := strings.CutPrefix(strings.TrimSuffix(repo, ".git"), "https://github.com/")
	if !ok {
		return nil, fmt.Errorf("fetching files is only supported for GitHub repos, got %s", repo)
	}
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", slug, ref, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := g.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
