package upstream

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRemote struct {
	tags    []string
	files   map[string]string // "<ref>:<path>" -> content
	fetched []string
}

func (f *fakeRemote) Tags(context.Context, string) ([]string, error) { return f.tags, nil }

func (f *fakeRemote) File(_ context.Context, _, ref, path string) ([]byte, error) {
	key := ref + ":" + path
	f.fetched = append(f.fetched, ref)
	c, ok := f.files[key]
	if !ok {
		return nil, fmt.Errorf("not found: %s", key)
	}
	return []byte(c), nil
}

var vTags = Source{Kind: KindGitTags, Repo: "https://github.com/x/y", TagPattern: `^v(\d+)\.(\d+)\.(\d+)$`}

func TestSelectNewestMinorsAtLatestPatch(t *testing.T) {
	r := &fakeRemote{tags: []string{
		"v1.19.0", "v1.19.1", "v1.20.0", "v1.20.3", "v1.20.10",
		"v1.21.0", "v1.21.0-rc.1", "v1.22.0-alpha.0", "helm-v9.9.9",
	}}
	cands, err := Candidates(context.Background(), r, vTags, Policy{Minors: 2})
	require.NoError(t, err)

	got := Select(cands, Policy{Minors: 2})
	assert.Equal(t, []Version{
		{Output: "1.21", Tag: "v1.21.0", Version: "1.21.0"},
		{Output: "1.20", Tag: "v1.20.10", Version: "1.20.10"},
	}, got)
}

func TestSelectKeepsPinnedMinors(t *testing.T) {
	r := &fakeRemote{tags: []string{"v0.13.6", "v0.14.0", "v0.15.0", "v0.16.1"}}
	cands, err := Candidates(context.Background(), r, vTags, Policy{Minors: 1})
	require.NoError(t, err)

	got := Select(cands, Policy{Minors: 1, Keep: []string{"0.13"}})
	assert.Equal(t, []string{"0.16", "0.13"}, outputs(got))
}

func TestSelectAcrossMajors(t *testing.T) {
	r := &fakeRemote{tags: []string{"v2.11.2", "v3.0.0", "v3.1.4", "v10.0.0"}}
	cands, err := Candidates(context.Background(), r, vTags, Policy{Minors: 3})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0", "3.1", "3.0"}, outputs(Select(cands, Policy{Minors: 3})))
}

func TestPrefixedTagPattern(t *testing.T) {
	src := Source{Kind: KindGitTags, Repo: "r", TagPattern: `^vertical-pod-autoscaler-(\d+)\.(\d+)\.(\d+)$`}
	r := &fakeRemote{tags: []string{"vertical-pod-autoscaler-1.5.1", "cluster-autoscaler-1.34.0", "vertical-pod-autoscaler-1.4.2"}}
	cands, err := Candidates(context.Background(), r, src, Policy{Minors: 5})
	require.NoError(t, err)
	got := Select(cands, Policy{Minors: 5})
	assert.Equal(t, []Version{
		{Output: "1.5", Tag: "vertical-pod-autoscaler-1.5.1", Version: "1.5.1"},
		{Output: "1.4", Tag: "vertical-pod-autoscaler-1.4.2", Version: "1.4.2"},
	}, got)
}

func TestHelmChartTracksAppVersion(t *testing.T) {
	src := Source{Kind: KindHelmChart, Repo: "r", TagPattern: `^v(\d+)\.(\d+)\.(\d+)$`, ChartPath: "traefik"}
	app := func(v string) string { return "appVersion: " + v + "\n" }
	r := &fakeRemote{
		tags: []string{"v37.0.0", "v38.0.0", "v38.0.1", "v39.0.4", "v39.0.5", "v39.1.0"},
		files: map[string]string{
			"v39.1.0:traefik/Chart.yaml": app("v3.6.10"),
			"v39.0.5:traefik/Chart.yaml": app("v3.6.9"),
			"v38.0.1:traefik/Chart.yaml": app("v3.5.2"),
			"v37.0.0:traefik/Chart.yaml": app("v3.4.0"),
		},
	}
	policy := Policy{Minors: 1}
	cands, err := Candidates(context.Background(), r, src, policy)
	require.NoError(t, err)

	got := Select(cands, policy)
	assert.Equal(t, []Version{{Output: "3.6", Tag: "v39.1.0", Version: "3.6.10"}}, got)
	// Stops after one app minor beyond what policy needs.
	assert.Equal(t, []string{"v39.1.0", "v39.0.5", "v38.0.1"}, r.fetched)
}

func TestHelmChartSameAppVersionPrefersNewestChart(t *testing.T) {
	src := Source{Kind: KindHelmChart, Repo: "r", TagPattern: `^v(\d+)\.(\d+)\.(\d+)$`, ChartPath: "c"}
	r := &fakeRemote{
		tags: []string{"v1.0.0", "v1.1.0"},
		files: map[string]string{
			"v1.0.0:c/Chart.yaml": "appVersion: 2.0.0\n",
			"v1.1.0:c/Chart.yaml": "appVersion: 2.0.0\n",
		},
	}
	cands, err := Candidates(context.Background(), r, src, Policy{Minors: 1})
	require.NoError(t, err)
	assert.Equal(t, "v1.1.0", Select(cands, Policy{Minors: 1})[0].Tag)
}

func TestCompare(t *testing.T) {
	current := []Version{
		{Output: "1.21", Tag: "v1.21.0", Version: "1.21.0"},
		{Output: "1.20", Tag: "v1.20.3", Version: "1.20.3"},
		{Output: "1.19", Tag: "v1.19.1", Version: "1.19.1"},
	}
	desired := []Version{
		{Output: "1.22", Tag: "v1.22.0", Version: "1.22.0"},
		{Output: "1.21", Tag: "v1.21.0", Version: "1.21.0"},
		{Output: "1.20", Tag: "v1.20.4", Version: "1.20.4"},
	}
	d := Compare(current, desired)
	assert.Equal(t, []Change{{Output: "1.22", To: "1.22.0"}}, d.Added)
	assert.Equal(t, []Change{{Output: "1.20", From: "1.20.3", To: "1.20.4"}}, d.Updated)
	assert.Equal(t, []Change{{Output: "1.19", From: "1.19.1"}}, d.Removed)
	assert.True(t, Compare(current, current).Empty())
}

func TestReadWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	u := &Upstream{
		Enabled:  true,
		Source:   vTags,
		Policy:   Policy{Minors: 3},
		Versions: []Version{{Output: "1.0", Tag: "v1.0.0", Version: "1.0.0"}},
	}
	require.NoError(t, Write(dir, u))
	got, err := Read(dir)
	require.NoError(t, err)
	assert.Equal(t, u, got)
}

func TestReadRejectsUnknownFieldsAndBadPolicy(t *testing.T) {
	for name, content := range map[string]string{
		"unknown field": `{"enabled":true,"source":{"kind":"git-tags","repo":"r","tagPattern":"^v(\\d+)\\.(\\d+)\\.(\\d+)$"},"policy":{"minors":1},"versions":[],"extra":1}`,
		"no minors":     `{"enabled":true,"source":{"kind":"git-tags","repo":"r","tagPattern":"^v(\\d+)\\.(\\d+)\\.(\\d+)$"},"policy":{"minors":0},"versions":[]}`,
		"few groups":    `{"enabled":true,"source":{"kind":"git-tags","repo":"r","tagPattern":"^v(\\d+)$"},"policy":{"minors":1},"versions":[]}`,
		"helm no path":  `{"enabled":true,"source":{"kind":"helm-chart","repo":"r","tagPattern":"^v(\\d+)\\.(\\d+)\\.(\\d+)$"},"policy":{"minors":1},"versions":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0o644))
			_, err := Read(dir)
			assert.Error(t, err)
		})
	}
}

func outputs(vs []Version) []string {
	var out []string
	for _, v := range vs {
		out = append(out, v.Output)
	}
	return out
}
