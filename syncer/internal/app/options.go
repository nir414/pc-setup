package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/nir414/pc-setup/syncer/internal/engine"
)

type globalOptions struct {
	ConfigPath string
	RootPath   string
	Verbose    bool
	Logger     *log.Logger
}

func parseGlobalOptions(args []string) (globalOptions, []string, error) {
	opts := globalOptions{}
	idx := 0
	for idx < len(args) {
		token := args[idx]
		if token == "--" {
			idx++
			break
		}
		if !strings.HasPrefix(token, "-") {
			break
		}

		var value string
		switch {
		case token == "--config" || token == "-c":
			if idx+1 >= len(args) {
				return opts, nil, fmt.Errorf("option %s requires a value", token)
			}
			value = args[idx+1]
			idx += 2
			opts.ConfigPath = value
			continue
		case strings.HasPrefix(token, "--config="):
			value = strings.TrimPrefix(token, "--config=")
			opts.ConfigPath = value
			idx++
			continue
		case token == "--root":
			if idx+1 >= len(args) {
				return opts, nil, fmt.Errorf("option %s requires a value", token)
			}
			value = args[idx+1]
			idx += 2
			opts.RootPath = value
			continue
		case strings.HasPrefix(token, "--root="):
			value = strings.TrimPrefix(token, "--root=")
			opts.RootPath = value
			idx++
			continue
		case token == "--verbose" || token == "-v":
			opts.Verbose = true
			idx++
			continue
		default:
			return opts, nil, fmt.Errorf("unknown option %s", token)
		}
	}

	if opts.Logger == nil {
		writer := io.Discard
		if opts.Verbose {
			writer = os.Stdout
		}
		opts.Logger = log.New(writer, "[syncer] ", log.LstdFlags)
	}

	if opts.ConfigPath == "" {
		if env := os.Getenv("SYNCER_CONFIG"); env != "" {
			opts.ConfigPath = env
		}
	}
	if opts.RootPath == "" {
		if env := os.Getenv("SYNCER_ROOT"); env != "" {
			opts.RootPath = env
		}
	}

	return opts, args[idx:], nil
}

func printStatusReport(report *engine.StatusReport) {
	if report == nil {
		fmt.Println("(no status information)")
		return
	}

	fmt.Println("Status Summary:")
	fmt.Printf("  Up-to-date    : %d\n", report.Summary.UpToDate)
	fmt.Printf("  Needs backup  : %d\n", report.Summary.NeedsBackup)
	fmt.Printf("  Needs sync    : %d\n", report.Summary.NeedsSync)
	fmt.Printf("  Conflicts     : %d\n", report.Summary.Conflicts)

	if len(report.Entries) == 0 {
		fmt.Println("\nEverything is up-to-date.")
		return
	}

	fmt.Println("\nDetails:")
	for _, entry := range report.Entries {
		fmt.Printf("  [%s] %s\n", entry.Status, entry.Path)
		if entry.System != nil && entry.Repo != nil && entry.System.Hash != entry.Repo.Hash {
			fmt.Printf("    system hash: %s\n", entry.System.Hash)
			fmt.Printf("    repo   hash: %s\n", entry.Repo.Hash)
		}
		if entry.System != nil && entry.Repo == nil {
			fmt.Printf("    system hash: %s\n", entry.System.Hash)
		}
		if entry.System == nil && entry.Repo != nil {
			fmt.Printf("    repo   hash: %s\n", entry.Repo.Hash)
		}
	}
}
