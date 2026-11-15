# dotset

A small CLI to manage dotfiles with symlinks — inspired by [GNU Stow](https://www.gnu.org/software/stow/).

`dotset` adopts a dotfile from your dotfiles repository into your home directory: it backs up the original file and replaces it with a symlink pointing back into the repo. Linux only.

## Install

```bash
make install                    # installs to ~/.local/bin
make install ARGS="--system"    # installs to /usr/local/bin (sudo when needed)
./install.sh --prefix /opt/dot  # installs to /opt/dot/bin
make build                      # builds ./bin/dotset
```

`make install` installs into `$(HOME)/.local/bin` by default; see `./install.sh --help`
for `--user`, `--system`, `--prefix`, and `--uninstall`.

Requires Go 1.27+ (the canonical version is declared in `go.mod`).

## Usage

Run it **from inside your dotfiles repository**, passing the name of a folder:

```bash
cd ~/dotfiles
dotset bash
```

The folder must contain exactly one regular file. For example:

```
~/dotfiles/
└── bash/
    └── .bashrc
```

```
✓ Adopted .bashrc
  Source : /home/user/dotfiles/bash/.bashrc
  Target : /home/user/.bashrc → /home/user/dotfiles/bash/.bashrc
  Backup : /home/user/.local/state/dotset/.bashrc.bak
```

The source filename determines the target. The leading dot is optional —
`bash/bashrc` also creates `~/.bashrc`.

To reverse an adoption, run the same folder name with `undo`:

```bash
dotset undo bash
```

It removes `~/.bashrc` only if it is a symlink to `bash/.bashrc`, then restores
the original from `~/.local/state/dotset/.bashrc.bak` when that backup exists.
If there is no backup, the symlink is removed anyway and a notice is printed.

```bash
dotset --version   # print version
```

## How it works

1. Looks up `./<folder>` in the current directory and requires it to contain
   exactly one regular file.
2. Uses that file's name to resolve the target (`~/.<file>`).
3. If the target exists, copies it to `~/.local/state/dotset/<file>.bak`.
4. Replaces the target with a symlink to `<folder>/<file>`.

`dotset undo <folder>` runs the same discovery, then:

1. Verifies `~/.<file>` is a symlink pointing to `<folder>/<file>`.
2. Removes the symlink.
3. Restores `<file>.bak` to `~/.<file>` and deletes the backup, when it exists.
   Without a backup, it stops after removing the symlink and says so.

## Behavior notes

- **Folders contain exactly one regular file.** Empty folders, folders with
  more than one entry, and subdirectories or symlinks inside the folder are
  rejected.
- **The folder must exist and be a directory** — validated before any change.
- **`undo` is a reserved first argument.** To adopt a folder literally named
  `undo`, pass it as a path: `dotset ./undo`.
- **No backup needed** when the target doesn't exist — the symlink is created directly.
- **Idempotency guard:** if the target is already a symlink to the same source, `dotset` exits with an error instead of touching anything.
- **Existing symlinks** (pointing elsewhere) are backed up by reading their content before replacement.
- **Undo only touches symlinks.** If `~/.<file>` is a regular file or a symlink
  pointing somewhere else, `dotset undo` refuses and leaves it untouched.
- **Undo consumes the backup.** After restoring `<file>.bak` it is deleted;
  re-adopting creates a fresh backup from the restored file.
- **Undo without a backup** still removes the symlink, then prints a notice
  that the original file was not restored.
- **Undo needs the source folder.** The target name comes from the single file
  in `<folder>`, so the folder must still exist with that file in place.

## Project layout

```
cmd/dotset/        entry point and manual DI wiring
internal/domain/   entities and the FileSystem port (zero dependencies)
internal/usecase/  business logic (AdoptDotfileUseCase)
internal/adapter/  os-backed filesystem and CLI output
```

Clean Architecture with a strict `domain → usecase → adapter → cmd` dependency rule. Standard library only — no third-party modules.

## Development

```bash
make test          # go test -race -count=1 ./...
make test-cover    # coverage report → coverage.html
make lint          # golangci-lint (config in .golangci.yml)
make fmt           # go fmt + go vet
make check         # fmt + test
make build-all     # Linux/macOS/Windows binaries in ./bin
make checksums     # SHA-256 of release binaries
make audit         # govulncheck
```

Exit codes are `0` on success, `1` on runtime errors, and `2` on usage errors.
Errors are printed by the CLI adapter; process exit is owned by `cmd/dotset/main.go`.
Build metadata (`version`, `commit`, `date`) lives in `internal/buildinfo`.

See [AGENTS.md](AGENTS.md) for agent-facing notes.

## Limitations / roadmap

- `undo` needs the source folder to still contain its single file; there is no
  `dotset undo <file>` form.
- Files only — directory adoption (e.g. `~/.config/nvim`) is not supported.
- No support for multiple sources per invocation or stow-style package trees.
