package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeLib builds a minimal generated lib: one output with the given kinds.
func writeLib(t *testing.T, dir, output string, kinds ...string) {
	t.Helper()
	files := map[string]string{
		output + "/main.libsonnet":        "(import 'gen.libsonnet')",
		output + "/_gen/g/main.libsonnet": "{}",
		"docs/" + output + "/README.md":   "# docs",
	}
	var fields []string
	for _, k := range kinds {
		files[output+"/_gen/g/v1/"+k+".libsonnet"] = "{ new(name):: { name: name } }"
		fields = append(fields, k+":: import '_gen/g/v1/"+k+".libsonnet'")
	}
	files[output+"/gen.libsonnet"] = "{ g:: { v1:: {" + strings.Join(fields, ",") + "} } }"
	for p, c := range files {
		full := filepath.Join(dir, p)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(c), 0o644))
	}
}

func config(outputs ...string) *Config {
	c := &Config{Repository: "github.com/o/x-libsonnet"}
	for _, o := range outputs {
		c.Specs = append(c.Specs, struct {
			Output string `yaml:"output"`
		}{o})
	}
	return c
}

func TestVerifyPasses(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "1.0", "foo", "bar")
	r := Verify(config("1.0"), dir, "", nil)
	assert.Empty(t, r.Errors)
	require.Len(t, r.Outputs, 1)
	assert.Equal(t, 3, r.Outputs[0].Kinds)
	assert.Greater(t, r.Outputs[0].Fields, 3)
}

func TestVerifyMissingAndUnexpectedOutputs(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "1.0", "foo")
	writeLib(t, dir, "0.9", "foo")
	r := Verify(config("1.0", "1.1"), dir, "", nil)
	assert.Len(t, r.Errors, 2)
	assert.Contains(t, strings.Join(r.Errors, "\n"), `unexpected output dir "0.9"`)
	assert.Contains(t, strings.Join(r.Errors, "\n"), "1.1:")
}

func TestVerifyMissingDocs(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "1.0", "foo")
	require.NoError(t, os.RemoveAll(filepath.Join(dir, "docs")))
	r := Verify(config("1.0"), dir, "", nil)
	assert.Contains(t, strings.Join(r.Errors, "\n"), "missing docs/1.0/README.md")
}

func TestVerifyEvaluationError(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "1.0", "foo")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "1.0/_gen/g/v1/foo.libsonnet"), []byte("{ broken: error 'boom' }"), 0o644))
	r := Verify(config("1.0"), dir, "", nil)
	assert.Contains(t, strings.Join(r.Errors, "\n"), "boom")
}

func TestVerifyRemovedKindsAgainstBaseline(t *testing.T) {
	baseline, dir := t.TempDir(), t.TempDir()
	writeLib(t, baseline, "1.0", "foo", "bar")
	writeLib(t, dir, "1.0", "foo")
	writeLib(t, dir, "1.1", "foo") // new minor: no baseline, removals allowed

	r := Verify(config("1.0", "1.1"), dir, baseline, nil)
	require.Len(t, r.Errors, 1)
	assert.Contains(t, r.Errors[0], "1.0: 1 generated files removed")
	assert.Equal(t, []string{"g/v1/bar.libsonnet"}, r.Outputs[0].Removed)
	assert.Equal(t, 0, r.Outputs[1].Baseline)
}
