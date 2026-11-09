// Package upstream reads and writes libs/<name>/upstream.json and decides
// which upstream versions a lib should track.
package upstream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const FileName = "upstream.json"

// Source kinds.
const (
	// KindGitTags tracks semver tags of Source.Repo.
	KindGitTags = "git-tags"
	// KindHelmChart tracks chart tags of Source.Repo; the tracked version is
	// the chart's appVersion, so outputs follow the app, not the chart.
	KindHelmChart = "helm-chart"
)

type Upstream struct {
	Enabled  bool      `json:"enabled"`
	Source   Source    `json:"source"`
	Policy   Policy    `json:"policy"`
	Versions []Version `json:"versions"`
}

type Source struct {
	Kind string `json:"kind"`
	// Repo is a git URL, e.g. https://github.com/cert-manager/cert-manager.
	Repo string `json:"repo"`
	// TagPattern must capture major, minor and patch as groups 1-3.
	TagPattern string `json:"tagPattern"`
	// ChartPath is the Chart.yaml directory inside Repo (helm-chart only).
	ChartPath string `json:"chartPath,omitempty"`
}

type Policy struct {
	// Minors is how many of the newest minor versions to track.
	Minors int `json:"minors"`
	// Keep lists extra outputs ("X.Y") that stay tracked when they fall out
	// of Minors, for consumers pinned to an older minor.
	Keep []string `json:"keep,omitempty"`
}

type Version struct {
	// Output is the directory in the lib repo, always "major.minor".
	Output string `json:"output"`
	// Tag is the git tag (chart tag for helm-chart) to generate from.
	Tag string `json:"tag"`
	// Version is the bare semver of the app, e.g. 1.21.0.
	Version string `json:"version"`
}

func Read(libDir string) (*Upstream, error) {
	data, err := os.ReadFile(filepath.Join(libDir, FileName))
	if err != nil {
		return nil, err
	}
	var u Upstream
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&u); err != nil {
		return nil, fmt.Errorf("%s/%s: %w", libDir, FileName, err)
	}
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("%s/%s: %w", libDir, FileName, err)
	}
	return &u, nil
}

func Write(libDir string, u *Upstream) error {
	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(libDir, FileName), append(data, '\n'), 0o644)
}

func (u *Upstream) Validate() error {
	switch u.Source.Kind {
	case KindGitTags:
	case KindHelmChart:
		if u.Source.ChartPath == "" {
			return fmt.Errorf("source.chartPath is required for %s", KindHelmChart)
		}
	default:
		return fmt.Errorf("unknown source.kind %q", u.Source.Kind)
	}
	if u.Source.Repo == "" {
		return fmt.Errorf("source.repo is required")
	}
	if _, err := compileTagPattern(u.Source.TagPattern); err != nil {
		return err
	}
	if u.Policy.Minors < 1 {
		return fmt.Errorf("policy.minors must be >= 1")
	}
	return nil
}

// Libs returns the names of lib dirs under libsDir that have an upstream.json,
// sorted. Libs without one are unmanaged and never generated.
func Libs(libsDir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(libsDir, "*", FileName))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, m := range matches {
		names = append(names, filepath.Base(filepath.Dir(m)))
	}
	sort.Strings(names)
	return names, nil
}
