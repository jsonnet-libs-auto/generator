package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-clix/cli"
	"github.com/jsonnet-libs-auto/generator/pkg/verify"
)

type verifyResult struct {
	Lib string `json:"lib"`
	verify.Report
}

func verifyCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "verify <lib>...",
		Short: "Check generated libs in gen/ before publishing (JSON); non-zero on failure",
	}
	libsDir := libsDirFlag(cmd)
	genDir := cmd.Flags().String("gen", "gen", "generator output directory")
	noBaseline := cmd.Flags().Bool("no-baseline", false, "don't compare against the published repo")
	jpath := cmd.Flags().StringSliceP("jpath", "J", nil, "jsonnet search path resolving doc-util/ (default: docsonnet Go module dir)")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("name at least one lib")
		}
		jp := *jpath
		if len(jp) == 0 {
			dir, err := moduleDir("github.com/jsonnet-libs/docsonnet")
			if err != nil {
				return err
			}
			jp = []string{dir}
		}

		var results []verifyResult
		failed := false
		for _, lib := range args {
			res := verifyResult{Lib: lib}
			cfg, err := verify.ReadConfig(filepath.Join(*libsDir, lib))
			if err != nil {
				res.Errors = []string{err.Error()}
				results, failed = append(results, res), true
				continue
			}

			baseline := ""
			if !*noBaseline {
				b, cleanup, err := clonePublished(cfg.Repository)
				if err != nil {
					return err
				}
				defer cleanup()
				baseline = b
			}

			res.Report = verify.Verify(cfg, filepath.Join(*genDir, cfg.Repository), baseline, jp)
			failed = failed || len(res.Errors) > 0
			results = append(results, res)
		}
		if err := printJSON(results); err != nil {
			return err
		}
		if failed {
			return fmt.Errorf("verification failed")
		}
		return nil
	}
	return cmd
}

func moduleDir(mod string) (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", mod).Output()
	if err != nil {
		return "", fmt.Errorf("locating %s: %w", mod, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// clonePublished returns "" when the repo doesn't exist or is empty yet.
func clonePublished(repository string) (string, func(), error) {
	tmp, err := os.MkdirTemp("", "baseline-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmp) }
	c := exec.Command("git", "clone", "--quiet", "--depth", "1", "https://"+repository, tmp)
	c.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if err := c.Run(); err != nil {
		return "", cleanup, nil
	}
	return tmp, cleanup, nil
}
