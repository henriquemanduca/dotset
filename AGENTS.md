# AGENTS.md — dotset

## Architecture

Clean Architecture with strict dependency rule: **domain → usecase → adapter → cmd**. Inner layers never import outer layers.

```
cmd/dotset/main.go            # Entry point, manual DI wiring
internal/domain/              # Entities + FileSystem port (zero deps)
internal/usecase/             # AdoptDotfileUseCase (business logic)
internal/adapter/             # OSFileSystem (os pkg) + CLI output
```

## Commands

```bash
make build        # go build → bin/dotset
make test         # go test -race -count=1 ./...
make test-cover   # tests + coverage.html
make lint         # golangci-lint run ./...
make lint-fix     # golangci-lint run --fix ./...
make fmt          # gofmt -s -w . && go vet ./...
make audit        # govulncheck ./...
```

Go is expected on `$PATH`. If it lives elsewhere, override with `make GO=/path/to/go build`.

Linter config is `.golangci.yml` at the project root. Run `make fmt` before committing.

## Testing

- **All use case tests use `mockFS`** (defined in `internal/usecase/adopt_test.go`) — never real I/O.
- Tests always run with `-race` and `-count=1` (no test caching).
- The `FileSystem` interface (`internal/domain/ports.go`) is the test seam; new features must pass a mock, not the OS adapter.

## Domain behavior (non-obvious)

- **Folder argument**: the single positional arg is a folder inside the cwd (`dotset bash` → `<cwd>/bash`). The folder must contain exactly one regular file; that file's name determines the target (`~/.<file>`). Relative resolution happens in the use case via `filepath.Join(cwd, folder)`; the use case validates the folder (`domain.ErrSourceDirNotFound`) and the entry count/type (`ErrSourceFolderEmpty`, `ErrSourceFolderMultiple`, `ErrSourceNotRegularFile`).
- **Leading dot is auto-prepended** for the target path only. `bash/bashrc` creates `~/.bashrc → <cwd>/bash/bashrc`. The source name stays as given.
- **Backup dir**: `~/.local/state/dotset/` — XDG state directory, not config.
- **Already-linked detection**: if `~/.file` is a symlink already pointing to the same source, the use case returns an error ("already linked to same source — nothing to do").
- **Directories are rejected**: only regular files. Returns `domain.ErrTargetIsDir`.
- **No-op when target doesn't exist**: symlink is created directly, no backup.

## Dependencies

Zero third-party modules. Standard library only (`go.mod` has no `require` block). Do not add external dependencies.