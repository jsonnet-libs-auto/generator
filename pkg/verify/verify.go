// Package verify checks a generated lib before it is published.
package verify

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/google/go-jsonnet"
	"gopkg.in/yaml.v3"
)

// Config is the subset of a rendered libs/<name>/config.yml that verify needs.
type Config struct {
	Repository string `yaml:"repository"`
	Specs      []struct {
		Output string `yaml:"output"`
	} `yaml:"specs"`
}

func ReadConfig(libDir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(libDir, "config.yml"))
	if err != nil {
		return nil, fmt.Errorf("%w (generate the lib first)", err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

type Output struct {
	Output string `json:"output"`
	// Kinds counts generated _gen files, a proxy for API surface.
	Kinds int `json:"kinds"`
	// Fields counts values reached when walking main.libsonnet.
	Fields   int      `json:"fields"`
	Baseline int      `json:"baseline,omitempty"`
	Removed  []string `json:"removed,omitempty"`
}

type Report struct {
	Outputs []Output `json:"outputs"`
	Errors  []string `json:"errors,omitempty"`
}

func (r *Report) errorf(format string, args ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

// walkDepth reaches group.version.kind.field(s); deeper is docs and mixins.
const walkDepth = 6

// Verify checks dir (a generated lib repo) against cfg. baseline is the
// currently published repo, or "" if there is none yet. jpath must resolve
// doc-util/main.libsonnet.
func Verify(cfg *Config, dir, baseline string, jpath []string) Report {
	var r Report

	want := map[string]bool{}
	for _, s := range cfg.Specs {
		want[s.Output] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		r.errorf("reading %s: %v", dir, err)
		return r
	}
	for _, e := range entries {
		if e.IsDir() && !want[e.Name()] && e.Name() != "docs" && e.Name()[0] != '.' {
			r.errorf("unexpected output dir %q", e.Name())
		}
	}

	for _, s := range cfg.Specs {
		o := Output{Output: s.Output}
		out := filepath.Join(dir, s.Output)

		kinds, err := genFiles(out)
		if err != nil {
			r.errorf("%s: %v", s.Output, err)
			r.Outputs = append(r.Outputs, o)
			continue
		}
		o.Kinds = len(kinds)
		if o.Kinds == 0 {
			r.errorf("%s: no generated kinds", s.Output)
		}

		if _, err := os.Stat(filepath.Join(dir, "docs", s.Output, "README.md")); err != nil {
			r.errorf("%s: missing docs/%s/README.md", s.Output, s.Output)
		}

		fields, err := evaluate(out, jpath)
		if err != nil {
			r.errorf("%s: main.libsonnet does not evaluate: %v", s.Output, err)
		}
		o.Fields = fields

		if baseline != "" {
			if old, err := genFiles(filepath.Join(baseline, s.Output)); err == nil {
				o.Baseline = len(old)
				o.Removed = missing(old, kinds)
				if len(o.Removed) > 0 {
					r.errorf("%s: %d generated files removed within an already published minor, e.g. %s", s.Output, len(o.Removed), o.Removed[0])
				}
			}
		}
		r.Outputs = append(r.Outputs, o)
	}
	return r
}

func genFiles(out string) ([]string, error) {
	root := filepath.Join(out, "_gen")
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(p) == ".libsonnet" {
			rel, _ := filepath.Rel(root, p)
			files = append(files, rel)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func missing(old, cur []string) []string {
	have := map[string]bool{}
	for _, c := range cur {
		have[c] = true
	}
	var out []string
	for _, o := range old {
		if !have[o] {
			out = append(out, o)
		}
	}
	return out
}

// evaluate forces every import reachable from main.libsonnet down to
// walkDepth and returns how many values it visited.
func evaluate(out string, jpath []string) (int, error) {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{JPaths: append([]string{out}, jpath...)})
	snippet := `
local walk(x, d) =
  if d > ` + strconv.Itoa(walkDepth) + ` || !std.isObject(x) then 1
  else std.foldl(function(acc, k) acc + walk(x[k], d + 1), std.objectFieldsAll(x), 1);
walk(import 'main.libsonnet', 0)
`
	res, err := vm.EvaluateAnonymousSnippet(filepath.Join(out, "verify.jsonnet"), snippet)
	if err != nil {
		return 0, err
	}
	var n int
	if _, err := fmt.Sscan(res, &n); err != nil {
		return 0, fmt.Errorf("unexpected result %q", res)
	}
	return n, nil
}
