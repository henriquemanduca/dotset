package usecase

import (
	"errors"
	"testing"

	"github.com/henrique/dotset/internal/domain"
)

func TestUndoDotfileUseCase_Execute(t *testing.T) {
	homeDir := "/home/user"
	backupDir := "/home/user/.local/state/dotset"
	workingDir := "/home/user/dotfiles"
	sourceDir := workingDir + "/bash"
	sourcePath := sourceDir + "/.bashrc"
	targetPath := homeDir + "/.bashrc"
	backupPath := backupDir + "/.bashrc.bak"

	// linkedRepo builds a mock filesystem where bash/.bashrc is adopted.
	linkedRepo := func() *mockFS {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourcePath, []byte("source content"))
		fs.Symlink(sourcePath, targetPath)
		return fs
	}

	t.Run("successful undo restores the backup and consumes it", func(t *testing.T) {
		fs := linkedRepo()
		fs.addFile(backupPath, []byte("original content"))

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute("bash", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Restored {
			t.Error("expected backup to be restored")
		}
		if result.BackupMissing {
			t.Error("did not expect BackupMissing")
		}

		data, err := fs.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("target not restored: %v", err)
		}
		if string(data) != "original content" {
			t.Errorf("target = %q, want %q", string(data), "original content")
		}

		info, err := fs.Stat(targetPath)
		if err != nil {
			t.Fatalf("stat target: %v", err)
		}
		if info.IsSymlink {
			t.Error("target is still a symlink")
		}

		if exists, _ := fs.Exists(backupPath); exists {
			t.Error("backup should be removed after being restored")
		}
	})

	t.Run("undo without a backup removes the symlink and reports it", func(t *testing.T) {
		fs := linkedRepo()

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		result, err := uc.Execute("bash", workingDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Restored {
			t.Error("did not expect a restore")
		}
		if !result.BackupMissing {
			t.Error("expected BackupMissing")
		}

		if exists, _ := fs.Exists(targetPath); exists {
			t.Error("symlink should be removed")
		}
	})

	t.Run("target does not exist", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourcePath, []byte("source content"))

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrTargetNotFound) {
			t.Errorf("expected ErrTargetNotFound, got %v", err)
		}
	})

	t.Run("target is a regular file", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)
		fs.addFile(sourcePath, []byte("source content"))
		fs.addFile(targetPath, []byte("user data"))
		fs.addFile(backupPath, []byte("original content"))

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrTargetNotSymlink) {
			t.Errorf("expected ErrTargetNotSymlink, got %v", err)
		}

		data, err := fs.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("target disappeared: %v", err)
		}
		if string(data) != "user data" {
			t.Errorf("target = %q, want untouched %q", string(data), "user data")
		}
	})

	t.Run("target symlink points to a different source", func(t *testing.T) {
		fs := linkedRepo()
		fs.symlinks[targetPath] = "/somewhere/else"

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrTargetNotLinked) {
			t.Errorf("expected ErrTargetNotLinked, got %v", err)
		}

		dest, err := fs.Readlink(targetPath)
		if err != nil {
			t.Fatalf("symlink disappeared: %v", err)
		}
		if dest != "/somewhere/else" {
			t.Errorf("symlink dest = %q, want untouched %q", dest, "/somewhere/else")
		}
	})

	t.Run("source folder does not exist", func(t *testing.T) {
		fs := linkedRepo()

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("nope", workingDir)

		if !errors.Is(err, domain.ErrSourceDirNotFound) {
			t.Errorf("expected ErrSourceDirNotFound, got %v", err)
		}
	})

	t.Run("empty source folder", func(t *testing.T) {
		fs := newMockFS()
		fs.addDir(workingDir)
		fs.addDir(sourceDir)

		uc := NewUndoDotfileUseCase(fs, homeDir, backupDir)
		_, err := uc.Execute("bash", workingDir)

		if !errors.Is(err, domain.ErrSourceFolderEmpty) {
			t.Errorf("expected ErrSourceFolderEmpty, got %v", err)
		}
	})
}

func TestAdoptThenUndo_RoundTrip(t *testing.T) {
	homeDir := "/home/user"
	backupDir := "/home/user/.local/state/dotset"
	workingDir := "/home/user/dotfiles"
	sourceDir := workingDir + "/bash"
	sourcePath := sourceDir + "/.bashrc"
	targetPath := homeDir + "/.bashrc"

	fs := newMockFS()
	fs.addDir(workingDir)
	fs.addDir(sourceDir)
	fs.addFile(sourcePath, []byte("source content"))
	fs.addFile(targetPath, []byte("original content"))

	adopt := NewAdoptDotfileUseCase(fs, homeDir, backupDir)
	if _, err := adopt.Execute("bash", workingDir); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	undo := NewUndoDotfileUseCase(fs, homeDir, backupDir)
	if _, err := undo.Execute("bash", workingDir); err != nil {
		t.Fatalf("undo: %v", err)
	}

	data, err := fs.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("target not restored: %v", err)
	}
	if string(data) != "original content" {
		t.Errorf("target = %q, want %q", string(data), "original content")
	}

	info, err := fs.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if info.IsSymlink {
		t.Error("target should be a regular file after undo")
	}

	if exists, _ := fs.Exists(backupDir + "/.bashrc.bak"); exists {
		t.Error("backup should be consumed by undo")
	}
}
