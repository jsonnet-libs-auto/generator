// libctl manages libs/<name>/upstream.json: listing managed libs, discovering
// upstream releases and bumping tracked versions.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-clix/cli"
	"github.com/jsonnet-libs-auto/generator/pkg/upstream"
)

func main() {
	log.SetFlags(0)
	root := &cli.Command{
		Use:   "libctl",
		Short: "Manage tracked upstream versions of generated libs",
	}
	root.AddCommand(listCmd(), discoverCmd(), bumpCmd(), verifyCmd())
	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func libsDirFlag(cmd *cli.Command) *string {
	return cmd.Flags().String("libs", "libs", "directory containing the libs")
}

func listCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "Print managed libs (those with an upstream.json), one per line",
	}
	libsDir := libsDirFlag(cmd)
	all := cmd.Flags().Bool("all", false, "include disabled libs")
	cmd.Run = func(cmd *cli.Command, args []string) error {
		names, err := upstream.Libs(*libsDir)
		if err != nil {
			return err
		}
		for _, n := range names {
			u, err := upstream.Read(filepath.Join(*libsDir, n))
			if err != nil {
				return err
			}
			if u.Enabled || *all {
				fmt.Println(n)
			}
		}
		return nil
	}
	return cmd
}

// Result is one lib's discovery outcome, printed as JSON.
type Result struct {
	Lib     string             `json:"lib"`
	Diff    upstream.Diff      `json:"diff"`
	Desired []upstream.Version `json:"desired,omitempty"`
	Error   string             `json:"error,omitempty"`
	current *upstream.Upstream
}

func discover(libsDir string, only []string) ([]Result, error) {
	names := only
	if len(names) == 0 {
		var err error
		if names, err = upstream.Libs(libsDir); err != nil {
			return nil, err
		}
	}

	remote := upstream.GitRemote{}
	var results []Result
	for _, n := range names {
		r := Result{Lib: n}
		u, err := upstream.Read(filepath.Join(libsDir, n))
		if err != nil {
			return nil, err
		}
		if !u.Enabled {
			continue
		}
		r.current = u

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		cands, err := upstream.Candidates(ctx, remote, u.Source, u.Policy)
		cancel()
		if err != nil {
			r.Error = err.Error()
			results = append(results, r)
			continue
		}
		r.Desired = upstream.Select(cands, u.Policy)
		if len(r.Desired) == 0 {
			r.Error = "no upstream release matches source.tagPattern"
		}
		r.Diff = upstream.Compare(u.Versions, r.Desired)
		results = append(results, r)
	}
	return results, nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func failed(results []Result) error {
	for _, r := range results {
		if r.Error != "" {
			return fmt.Errorf("discovery failed for at least one lib")
		}
	}
	return nil
}

func discoverCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "discover [lib...]",
		Short: "Report how each lib's tracked versions differ from upstream (JSON)",
	}
	libsDir := libsDirFlag(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		results, err := discover(*libsDir, args)
		if err != nil {
			return err
		}
		if err := printJSON(results); err != nil {
			return err
		}
		return failed(results)
	}
	return cmd
}

func bumpCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "bump [lib...]",
		Short: "Rewrite upstream.json versions to match upstream; prints what changed (JSON)",
	}
	libsDir := libsDirFlag(cmd)
	cmd.Run = func(cmd *cli.Command, args []string) error {
		results, err := discover(*libsDir, args)
		if err != nil {
			return err
		}
		var changed []Result
		for _, r := range results {
			if r.Error != "" || r.Diff.Empty() {
				continue
			}
			r.current.Versions = r.Desired
			if err := upstream.Write(filepath.Join(*libsDir, r.Lib), r.current); err != nil {
				return err
			}
			changed = append(changed, r)
		}
		for _, r := range results {
			if r.Error != "" {
				changed = append(changed, r)
			}
		}
		if err := printJSON(changed); err != nil {
			return err
		}
		return failed(results)
	}
	return cmd
}
