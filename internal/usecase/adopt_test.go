package usecase

import (
	"errors"
	"os"
	"testing"

	"github.com/henrique/dotset/internal/domain"
)

// mockFS is a fake filesystem for testing use cases.
type mockFS struct {
	files    map[string][]byte
	symlinks map[string]string
	infos    map[string]domain.FileInfo
}

func newMockFS() *mockFS {
	return &mockFS{
		files:    make(map[string][]byte),
		symlinks: make(map[string]string),
		infos:    make(map[string]domain.FileInfo),
	}
}

func (m *mockFS) Stat(path string) (domain.FileInfo, error) {
	info, ok := m.infos[path]
	if !ok {
		return domain.FileInfo{}, os.ErrNotExist
	}
	return info, nil
}

func (m *mockFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (m *mockFS) WriteFile(path string, data []byte, _ uint32) error {
	m.files[path] = data
	m.infos[path] = domain.FileInfo{Path: path}
	return nil
}

func (m *mockFS) MkdirAll(_ string, _ uint32) error {
	return nil
}

func (m *mockFS) Symlink(oldname, newname string) error {
	m.symlinks[newname] = oldname
	m.infos[newname] = domain.FileInfo{Path: newname, IsSymlink: true}
	return nil
}

func (m *mockFS) Remove(path string) error {
	delete(m.files, path)
	delete(m.symlinks, path)
	delete(m.infos, path)
	return nil
}

func (m *mockFS) Exists(path string) (bool, error) {
	_, hasFile := m.files[path]
	_, hasLink := m.symlinks[path]
	_, hasInfo := m.infos[path]
	return hasFile || hasLink || hasInfo, nil
}

func (m *mockFS) Readlink(path string) (string, error) {
	dest, ok := m.symlinks[path]
	if !ok {
		return "", os.ErrNotExist
	}
	return dest, nil
}

func (m *mockFS) addFile(path string, content []byte) {
	m.files[path] = content
	m.infos[path] = domain.FileInfo{Path: path}
}

func (m *mockFS) addDir(path string) {
	m.infos[path] = domain.FileInfo{Path: path, IsDir: true}
}

func TestAdoptDotfileUseCase_Execute(t *testing.T) {
	homeDir := "/home/user"
	backupDir := "/home/user/.local/state/dotset"
	workingDir := "/home/user/dotfiles"

	t.Run("successful adoption with backup", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		// Source file exists in dotfiles repo.
		fs.addFile(workingDir+"/.bashrc", []byte("source content"))
		// Target file exists in home.
		fs.addFile(homeDir+"/.bashrc", []byte("original content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute(".bashrc", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.BackedUp {
			t.Error("expected backup to be created")
		}

		// Check backup was written.
		backupData, err := fs.ReadFile(homeDir + "/.local/state/dotset/.bashrc.bak")
		if err != nil {
			t.Fatalf("backup not found: %v", err)
		}
		if string(backupData) != "original content" {
			t.Errorf("backup = %q, want %q", string(backupData), "original content")
		}

		// Check symlink was created.
		linkDest, err := fs.Readlink(homeDir + "/.bashrc")
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}
		if linkDest != workingDir+"/.bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, workingDir+"/.bashrc")
		}
	})

	t.Run("adoption without existing target (no backup)", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addFile(workingDir+"/.bashrc", []byte("source content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute(".bashrc", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.BackedUp {
			t.Error("expected no backup when target does not exist")
		}

		linkDest, err := fs.Readlink(homeDir + "/.bashrc")
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}
		if linkDest != workingDir+"/.bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, workingDir+"/.bashrc")
		}
	})

	t.Run("source not found", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute(".bashrc", workingDir)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrSourceNotFound) {
			t.Errorf("expected ErrSourceNotFound, got %v", err)
		}
	})

	t.Run("target is a directory", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addFile(workingDir+"/.config", []byte("source content"))
		fs.addDir(homeDir + "/.config")

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute(".config", workingDir)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrTargetIsDir) {
			t.Errorf("expected ErrTargetIsDir, got %v", err)
		}
	})

	t.Run("filename without leading dot", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addFile(workingDir+"/bashrc", []byte("source content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute("bashrc", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		linkDest, err := fs.Readlink(homeDir + "/.bashrc")
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}
		if linkDest != workingDir+"/bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, workingDir+"/bashrc")
		}
		// Source name stays "bashrc" (no dot added in repo).
		if result.Dotfile.Name != "bashrc" {
			t.Errorf("name = %q, want %q", result.Dotfile.Name, "bashrc")
		}
	})

	t.Run("already linked to same source", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addFile(workingDir+"/.bashrc", []byte("source content"))
		fs.Symlink(workingDir+"/.bashrc", homeDir+"/.bashrc")

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute(".bashrc", workingDir)

		if err == nil {
			t.Fatal("expected error for already-linked file")
		}
	})

	t.Run("custom source folder", func(t *testing.T) {
		customDir := "/home/user/dotfiles/work"
		fs := newMockFS()
		fs.addDir(customDir)
		fs.addFile(customDir+"/.bashrc", []byte("work content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute(".bashrc", customDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Dotfile.SourcePath != customDir+"/.bashrc" {
			t.Errorf("source = %q, want %q", result.Dotfile.SourcePath, customDir+"/.bashrc")
		}

		linkDest, err := fs.Readlink(homeDir + "/.bashrc")
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}
		if linkDest != customDir+"/.bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, customDir+"/.bashrc")
		}
	})

	t.Run("source folder does not exist", func(t *testing.T) {
		fs := newMockFS()
		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute(".bashrc", "/home/user/missing")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrSourceDirNotFound) {
			t.Errorf("expected ErrSourceDirNotFound, got %v", err)
		}
	})

	t.Run("source folder is a file", func(t *testing.T) {
		fs := newMockFS()
		fs.addFile("/home/user/not-a-dir", []byte("data"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute(".bashrc", "/home/user/not-a-dir")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrSourceDirNotFound) {
			t.Errorf("expected ErrSourceDirNotFound, got %v", err)
		}
	})
}
