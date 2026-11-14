package adapter

import (
	"fmt"
	"io"
	"os"

	"github.com/henrique/dotset/internal/usecase"
)

// CLI handles command-line output and argument parsing.
type CLI struct {
	stdout io.Writer
	stderr io.Writer
}

// NewCLI creates a new CLI adapter writing to the given output streams.
func NewCLI(stdout, stderr io.Writer) *CLI {
	return &CLI{stdout: stdout, stderr: stderr}
}

// PrintSuccess prints the adoption result to stdout.
func (c *CLI) PrintSuccess(result *usecase.AdoptResult) {
	d := result.Dotfile
	fmt.Fprintf(c.stdout, "✓ Adopted %s\n", d.Name)
	fmt.Fprintf(c.stdout, "  Source : %s\n", d.SourcePath)
	fmt.Fprintf(c.stdout, "  Target : %s → %s\n", d.TargetPath, d.SourcePath)
	if result.BackedUp {
		fmt.Fprintf(c.stdout, "  Backup : %s\n", d.BackupPath)
	}
}

// PrintUndoSuccess prints the undo result to stdout.
func (c *CLI) PrintUndoSuccess(result *usecase.UndoResult) {
	d := result.Dotfile
	fmt.Fprintf(c.stdout, "✓ Unlinked %s\n", d.Name)
	fmt.Fprintf(c.stdout, "  Target  : %s\n", d.TargetPath)
	if result.Restored {
		fmt.Fprintf(c.stdout, "  Restored: %s\n", d.BackupPath)
		return
	}
	fmt.Fprintf(c.stdout, "  Backup  : none at %s — original not restored\n", d.BackupPath)
}

// PrintError prints an error to stderr and exits with code 1.
func (c *CLI) PrintError(err error) {
	fmt.Fprintf(c.stderr, "dotset: %v\n", err)
	os.Exit(1)
}

// PrintVersion prints the application version.
func (c *CLI) PrintVersion(version string) {
	fmt.Fprintf(c.stdout, "dotset %s\n", version)
}

// PrintUsage prints the usage information.
func (c *CLI) PrintUsage() {
	fmt.Fprintf(c.stderr, `dotset — manage dotfiles with symlinks

Usage:
  dotset [flags] <folder>
  dotset [flags] undo <folder>

Flags:
  -v, --version        Print version

Description:
  Adopts the file inside <folder> into your home directory via a symlink.

  <folder> is resolved relative to the current directory and must contain
  exactly one regular file. That file's name determines the target.

  Steps:
    1. Looks up ./<folder> in the current directory.
    2. Reads the single file inside it.
    3. Backs up ~/.<file> to ~/.local/state/dotset/<file>.bak if it exists.
    4. Creates a symlink: ~/.<file> -> ./<folder>/<file>

  undo reverses the adoption: it removes the symlink and restores the
  original file from ~/.local/state/dotset/<file>.bak when a backup exists.
  If there is no backup, the symlink is still removed and a notice is shown.

Examples:
  dotset bash                     # Adopt the file inside ./bash
  dotset nvim                     # Adopt the file inside ./nvim
  dotset undo bash                # Unlink it and restore the original
  dotset --version                # Print version
`)
}
