package domain

// FileInfo holds metadata about a file.
type FileInfo struct {
	Path      string
	IsDir     bool
	IsSymlink bool
}

// FileSystem is the port (interface) that abstracts filesystem operations.
// Adapters (e.g., OSFileSystem) implement this interface, allowing the use case
// layer to remain decoupled from the real operating system — making it testable.
type FileSystem interface {
	// Stat returns file info without following symlinks.
	Stat(path string) (FileInfo, error)
	// ReadFile reads the entire contents of a file.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes data to a file with the given permissions.
	WriteFile(path string, data []byte, perm uint32) error
	// MkdirAll creates a directory and all parents.
	MkdirAll(path string, perm uint32) error
	// Symlink creates newname as a symbolic link to oldname.
	Symlink(oldname, newname string) error
	// Remove removes the named file or (empty) directory.
	Remove(path string) error
	// Exists returns true if the path exists (any type).
	Exists(path string) (bool, error)
	// Readlink returns the destination of a symbolic link.
	Readlink(path string) (string, error)
}
