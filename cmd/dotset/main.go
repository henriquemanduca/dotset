// dotset — manage dotfiles with symlinks, inspired by GNU Stow.
//
// dotset adopts a dotfile from your dotfiles repository into your home directory.
// It backs up the original and replaces it with a symlink. The undo subcommand
// reverses that operation, restoring the backup when one exists.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/henrique/dotset/internal/adapter"
	"github.com/henrique/dotset/internal/usecase"
)

// Version is injected at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	cli := adapter.NewCLI(os.Stdout, os.Stderr)

	flagSet := flag.NewFlagSet("dotset", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	flagSet.Usage = cli.PrintUsage

	var showVersion bool
	flagSet.BoolVar(&showVersion, "version", false, "print version")
	flagSet.BoolVar(&showVersion, "v", false, "print version")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(2)
	}

	if showVersion {
		cli.PrintVersion(Version)
		return
	}

	// Dispatch: `dotset undo <folder>` or `dotset <folder>`.
	args := flagSet.Args()
	undoing := len(args) > 0 && args[0] == "undo"
	if undoing && len(args) != 2 {
		cli.PrintUsage()
		os.Exit(1)
	}
	if !undoing && len(args) != 1 {
		cli.PrintUsage()
		os.Exit(1)
	}

	folder := args[len(args)-1]

	// Resolve the current directory: the source folder lives inside it.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dotset: failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Resolve home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dotset: failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	// Backup directory: ~/.local/state/dotset/
	backupDir := filepath.Join(homeDir, ".local", "state", "dotset")

	// Wire dependencies (manual DI).
	fs := adapter.NewOSFileSystem()

	if undoing {
		undoUseCase := usecase.NewUndoDotfileUseCase(fs, homeDir, backupDir)

		result, err := undoUseCase.Execute(folder, cwd)
		if err != nil {
			cli.PrintError(err)
		}

		cli.PrintUndoSuccess(result)
		return
	}

	adoptUseCase := usecase.NewAdoptDotfileUseCase(fs, homeDir, backupDir)

	// Execute.
	result, err := adoptUseCase.Execute(folder, cwd)
	if err != nil {
		cli.PrintError(err)
	}

	cli.PrintSuccess(result)
}
