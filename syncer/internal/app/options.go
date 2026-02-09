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
		fmt.Println("(상태 정보 없음)")
		return
	}

	fmt.Println("상태 요약:")
	fmt.Printf("  최신 상태    : %d개\n", report.Summary.UpToDate)
	fmt.Printf("  백업 필요    : %d개\n", report.Summary.NeedsBackup)
	fmt.Printf("  무시됨        : %d개 (시스템에만 존재, SyncData에 미정의)\n", report.Summary.Ignored)

	if len(report.Entries) == 0 {
		fmt.Println("\n모든 파일이 최신 상태입니다.")
		return
	}

	fmt.Println("\n상세 내역:")
	for _, entry := range report.Entries {
		var statusText string
		switch entry.Status {
		case "system_modified":
			statusText = "수정됨"
		case "system_deleted":
			statusText = "삭제됨"
		case "repo_only":
			statusText = "백업 필요 (SyncData에 정의됨)"
		default:
			statusText = string(entry.Status)
		}
		fmt.Printf("  [%s] %s\n", statusText, entry.Path)
	}
}
