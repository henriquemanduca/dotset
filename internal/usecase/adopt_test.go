package usecase

import (
	"errors"
	"os"
	"sort"
	"strings"
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

func (m *mockFS) ReadDir(path string) ([]domain.FileInfo, error) {
	info, ok := m.infos[path]
	if !ok || !info.IsDir {
		return nil, os.ErrNotExist
	}

	prefix := path + "/"
	var entries []domain.FileInfo
	for p, fi := range m.infos {
		rest, found := strings.CutPrefix(p, prefix)
		if !found || rest == "" || strings.Contains(rest, "/") {
			continue
		}
		entries = append(entries, fi)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
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
	sourceDir := workingDir + "/bash"

	t.Run("successful adoption with backup", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourceDir+"/.bashrc", []byte("source content"))
		fs.addFile(homeDir+"/.bashrc", []byte("original content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute("bash", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.BackedUp {
			t.Error("expected backup to be created")
		}
		if result.Dotfile.SourcePath != sourceDir+"/.bashrc" {
			t.Errorf("source = %q, want %q", result.Dotfile.SourcePath, sourceDir+"/.bashrc")
		}
		if result.Dotfile.TargetPath != homeDir+"/.bashrc" {
			t.Errorf("target = %q, want %q", result.Dotfile.TargetPath, homeDir+"/.bashrc")
		}

		backupData, err := fs.ReadFile(homeDir + "/.local/state/dotset/.bashrc.bak")
		if err != nil {
			t.Fatalf("backup not found: %v", err)
		}
		if string(backupData) != "original content" {
			t.Errorf("backup = %q, want %q", string(backupData), "original content")
		}

		linkDest, err := fs.Readlink(homeDir + "/.bashrc")
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}
		if linkDest != sourceDir+"/.bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, sourceDir+"/.bashrc")
		}
	})

	t.Run("adoption without existing target (no backup)", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourceDir+"/.bashrc", []byte("source content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute("bash", workingDir)

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
		if linkDest != sourceDir+"/.bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, sourceDir+"/.bashrc")
		}
	})

	t.Run("filename without leading dot", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourceDir+"/bashrc", []byte("source content"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute("bash", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		linkDest, err := fs.Readlink(homeDir + "/.bashrc")
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}
		if linkDest != sourceDir+"/bashrc" {
			t.Errorf("symlink dest = %q, want %q", linkDest, sourceDir+"/bashrc")
		}
		if result.Dotfile.Name != "bashrc" {
			t.Errorf("name = %q, want %q", result.Dotfile.Name, "bashrc")
		}
	})

	t.Run("target is a directory", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourceDir+"/config", []byte("source content"))
		fs.addDir(homeDir + "/.config")

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domain.ErrTargetIsDir) {
			t.Errorf("expected ErrTargetIsDir, got %v", err)
		}
	})

	t.Run("already linked to same source", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourceDir+"/.bashrc", []byte("source content"))
		fs.Symlink(sourceDir+"/.bashrc", homeDir+"/.bashrc")

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if err == nil {
			t.Fatal("expected error for already-linked file")
		}
	})

	t.Run("empty source folder", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceFolderEmpty) {
			t.Errorf("expected ErrSourceFolderEmpty, got %v", err)
		}
	})

	t.Run("source folder with multiple files", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourceDir+"/.bashrc", []byte("one"))
		fs.addFile(sourceDir+"/.bash_profile", []byte("two"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceFolderMultiple) {
			t.Errorf("expected ErrSourceFolderMultiple, got %v", err)
		}
	})

	t.Run("source folder contains a subdirectory", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addDir(sourceDir + "/conf")

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceNotRegularFile) {
			t.Errorf("expected ErrSourceNotRegularFile, got %v", err)
		}
	})

	t.Run("source folder contains a symlink", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.Symlink("/tmp/elsewhere", sourceDir+"/bashrc")

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceNotRegularFile) {
			t.Errorf("expected ErrSourceNotRegularFile, got %v", err)
		}
	})

	t.Run("source folder does not exist", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceDirNotFound) {
			t.Errorf("expected ErrSourceDirNotFound, got %v", err)
		}
	})

	t.Run("folder argument is a file", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addFile(workingDir+"/bash", []byte("data"))

		uc := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceDirNotFound) {
			t.Errorf("expected ErrSourceDirNotFound, got %v", err)
		}
	})
}
