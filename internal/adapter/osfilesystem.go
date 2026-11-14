// Package adapter provides implementations of domain ports that interact
// with the real operating system and command-line interface.
package adapter

import (
	"os"

	"github.com/henrique/dotset/internal/domain"
)

// Compile-time check: OSFileSystem implements domain.FileSystem.
var _ domain.FileSystem = (*OSFileSystem)(nil)

// OSFileSystem implements domain.FileSystem using the os standard library.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS-backed filesystem adapter.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (fs *OSFileSystem) Stat(path string) (domain.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return domain.FileInfo{}, err
	}
	return domain.FileInfo{
		Path:      path,
		IsDir:     info.IsDir(),
		IsSymlink: info.Mode()&os.ModeSymlink != 0,
	}, nil
}

func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *OSFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}

func (fs *OSFileSystem) MkdirAll(path string, perm uint32) error {
	return os.MkdirAll(path, os.FileMode(perm))
}

func (fs *OSFileSystem) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (fs *OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (fs *OSFileSystem) Exists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (fs *OSFileSystem) Readlink(path string) (string, error) {
	return os.Readlink(path)
}
