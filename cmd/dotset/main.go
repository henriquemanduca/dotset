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
	"github.com/henrique/dotset/internal/buildinfo"
	"github.com/henrique/dotset/internal/usecase"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run executes the CLI and returns the process exit code:
// 0 on success, 1 on runtime errors, 2 on usage errors.
func run(args []string) int {
	cli := adapter.NewCLI(os.Stdout, os.Stderr)

	flagSet := flag.NewFlagSet("dotset", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	flagSet.Usage = cli.PrintUsage

	var showVersion bool
	flagSet.BoolVar(&showVersion, "version", false, "print version")
	flagSet.BoolVar(&showVersion, "v", false, "print version")

	if err := flagSet.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if showVersion {
		cli.PrintVersion(buildinfo.Get().String())
		return 0
	}

	// Dispatch: `dotset undo <folder>` or `dotset <folder>`.
	rest := flagSet.Args()
	undoing := len(rest) > 0 && rest[0] == "undo"

	if undoing {
		if len(rest) != 2 {
			cli.PrintUsage()
			return 2
		}
	} else if len(rest) != 1 {
		cli.PrintUsage()
		return 2
	}

	folder := rest[len(rest)-1]

	// Resolve the current directory: the source folder lives inside it.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dotset: failed to get current directory: %v\n", err)
		return 1
	}

	// Resolve home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dotset: failed to get home directory: %v\n", err)
		return 1
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
			return 1
		}

		cli.PrintUndoSuccess(result)
		return 0
	}

	adoptUseCase := usecase.NewAdoptDotfileUseCase(fs, homeDir, backupDir)

	result, err := adoptUseCase.Execute(folder, cwd)
	if err != nil {
		cli.PrintError(err)
		return 1
	}

	cli.PrintSuccess(result)
	return 0
}
