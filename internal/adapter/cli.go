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
  dotset [flags] <file>

Flags:
  -f, --folder <dir>   Folder where the dotfile is looked up (default: current directory)
  -v, --version        Print version

Description:
  Adopts a dotfile from a dotfiles folder into your home directory via a symlink.

  Steps:
    1. Searches for <file> in the chosen folder (default: current directory).
    2. Looks up ~/.<file> in your home directory.
    3. Backs up the original to ~/.local/state/dotset/<file>.bak.
    4. Creates a symlink: ~/.<file> -> <folder>/<file>

Examples:
  dotset .bashrc                  # Adopt .bashrc from the current directory
  dotset bashrc                   # Same as above (leading dot is added automatically)
  dotset -f ~/dotfiles/work .bashrc
                                  # Look for .bashrc in ~/dotfiles/work
  dotset --folder=./server .vimrc # Look for .vimrc in ./server
  dotset --version                # Print version
`)
}
