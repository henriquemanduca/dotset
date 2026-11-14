# dotset

A small CLI to manage dotfiles with symlinks — inspired by [GNU Stow](https://www.gnu.org/software/stow/).

`dotset` adopts a dotfile from your dotfiles repository into your home directory: it backs up the original file and replaces it with a symlink pointing back into the repo. Linux only.

## Install

```bash
make install       # installs to $GOPATH/bin
# or
make build         # builds ./bin/dotset
```

Requires Go 1.27+.

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

```bash
dotset --version   # print version
```

## How it works

1. Looks up `./<folder>` in the current directory and requires it to contain
   exactly one regular file.
2. Uses that file's name to resolve the target (`~/.<file>`).
3. If the target exists, copies it to `~/.local/state/dotset/<file>.bak`.
4. Replaces the target with a symlink to `<folder>/<file>`.

## Behavior notes

- **Folders contain exactly one regular file.** Empty folders, folders with
  more than one entry, and subdirectories or symlinks inside the folder are
  rejected.
- **The folder must exist and be a directory** — validated before any change.
- **No backup needed** when the target doesn't exist — the symlink is created directly.
- **Idempotency guard:** if the target is already a symlink to the same source, `dotset` exits with an error instead of touching anything.
- **Existing symlinks** (pointing elsewhere) are backed up by reading their content before replacement.

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
make lint          # golangci-lint
make fmt           # gofmt + go vet
```

See [AGENTS.md](AGENTS.md) for agent-facing notes.

## Limitations / roadmap

- No `unlink` / restore command yet.
- Files only — directory adoption (e.g. `~/.config/nvim`) is not supported.
- No support for multiple sources per invocation or stow-style package trees.
