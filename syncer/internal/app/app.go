package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nir414/pc-setup/syncer/internal/config"
	"github.com/nir414/pc-setup/syncer/internal/engine"
	"github.com/nir414/pc-setup/syncer/internal/state"
)

const (
	defaultConfigName = "sync.toml"
	stateDirName      = ".syncer"
	stateFileName     = "state.json"
)

// App coordinates command execution.
type App struct{}

// New creates a new App instance.
func New() *App {
	return &App{}
}

// Run executes the application using the provided arguments.
func (a *App) Run(ctx context.Context, args []string) error {
	opts, rest, err := parseGlobalOptions(args)
	if err != nil {
		return err
	}

	if len(rest) == 0 {
		return errors.New("no command provided; expected one of: backup, status, sync")
	}

	command := rest[0]
	commandArgs := rest[1:]

	root, configPath, err := resolvePaths(opts)
	if err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	snapshotPath := filepath.Join(root, stateDirName, stateFileName)
	store := state.NewFileStore(snapshotPath)

	eng := engine.New(engine.Options{
		Root:          root,
		Config:        cfg,
		SnapshotStore: store,
		Logger:        opts.Logger,
	})

	switch strings.ToLower(command) {
	case "backup":
		return a.runBackup(ctx, eng, commandArgs)
	case "status":
		return a.runStatus(ctx, eng, commandArgs, opts)
	case "sync":
		return a.runSync(ctx, eng, commandArgs, opts)
	case "help", "-h", "--help":
		fmt.Print(helpText)
		return nil
	default:
		return fmt.Errorf("unknown command %q; expected one of: backup, status, sync", command)
	}
}

func (a *App) runBackup(ctx context.Context, eng *engine.Engine, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("backup command does not accept additional arguments: %v", args)
	}

	result, err := eng.Backup(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Backup completed: %d files copied, %d skipped, %.2f MiB moved\n",
		result.CopiedFiles,
		result.SkippedFiles,
		float64(result.CopiedBytes)/1024/1024,
	)

	return nil
}

func (a *App) runStatus(ctx context.Context, eng *engine.Engine, args []string, opts globalOptions) error {
	if len(args) != 0 {
		return fmt.Errorf("status command does not accept additional arguments: %v", args)
	}

	report, err := eng.Status(ctx)
	if err != nil {
		return err
	}

	printStatusReport(report)
	return nil
}

func (a *App) runSync(ctx context.Context, eng *engine.Engine, args []string, opts globalOptions) error {
	if len(args) != 0 {
		return fmt.Errorf("sync command does not accept additional arguments: %v", args)
	}

	result, err := eng.Sync(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Sync completed: %d files updated, %d skipped, %d removals, %.2f MiB moved\n",
		result.UpdatedFiles,
		result.SkippedFiles,
		result.RemovedFiles,
		float64(result.UpdatedBytes)/1024/1024,
	)

	return nil
}

func resolvePaths(opts globalOptions) (string, string, error) {
	cfgPath := opts.ConfigPath
	if cfgPath == "" {
		cfgPath = defaultConfigName
	}

	absCfg, err := filepath.Abs(cfgPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve config path: %w", err)
	}

	root := opts.RootPath
	if root == "" {
		root = filepath.Dir(absCfg)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve root path: %w", err)
	}

	return absRoot, absCfg, nil
}

const helpText = `syncer - Windows 설정 백업/동기화 도우미

사용법:
  syncer [전역 옵션] <command>

전역 옵션:
  --config <path>   사용할 TOML 설정 파일 경로 (기본: sync.toml)
  --root <path>     SyncData가 위치한 프로젝트 루트 (기본: 설정 파일 위치)

명령:
  backup            시스템 -> 저장소로 백업 실행
  status            현재 차이점 요약 출력
  sync              저장소 -> 시스템 동기화 실행
  help              이 도움말 출력
`
