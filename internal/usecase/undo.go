package usecase

import (
	"fmt"

	"github.com/henrique/dotset/internal/domain"
)

// UndoDotfileUseCase reverses an adoption:
//  1. Resolves the dotfile from the source folder, exactly like adoption.
//  2. Verifies the target is a symlink pointing to that source.
//  3. Removes the symlink.
//  4. Restores the backup of the original file, when one exists.
type UndoDotfileUseCase struct {
	fs        domain.FileSystem
	homeDir   string
	backupDir string
}

// NewUndoDotfileUseCase creates a new undo use case with the given filesystem
// adapter and configuration.
func NewUndoDotfileUseCase(fs domain.FileSystem, homeDir, backupDir string) *UndoDotfileUseCase {
	return &UndoDotfileUseCase{
		fs:        fs,
		homeDir:   homeDir,
		backupDir: backupDir,
	}
}

// UndoResult holds the result of an undo operation.
type UndoResult struct {
	Dotfile *domain.Dotfile
	// Restored is true when a backup existed and was copied back to the target.
	Restored bool
	// BackupMissing is true when the symlink was removed but no backup existed,
	// so the original file could not be restored.
	BackupMissing bool
}

// Execute runs the undo use case for the given folder from workingDir.
// The folder lives inside workingDir and must contain exactly one regular file;
// that file's name determines the dotfile to unlink from the home directory.
func (uc *UndoDotfileUseCase) Execute(folder, workingDir string) (*UndoResult, error) {
	// 0. Locate the single regular file inside the source folder.
	d, err := resolveDotfile(uc.fs, folder, workingDir, uc.homeDir, uc.backupDir)
	if err != nil {
		return nil, err
	}

	// 1. The target must exist — otherwise there is nothing to undo.
	targetExists, err := uc.fs.Exists(d.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("checking target: %w", err)
	}
	if !targetExists {
		return nil, fmt.Errorf("%w: %s — nothing to undo", domain.ErrTargetNotFound, d.TargetPath)
	}

	// 2. Refuse to remove anything that is not a symlink.
	targetInfo, err := uc.fs.Stat(d.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("stat target: %w", err)
	}
	if !targetInfo.IsSymlink {
		return nil, fmt.Errorf("%w: %s", domain.ErrTargetNotSymlink, d.TargetPath)
	}

	// 3. Refuse to unlink a symlink that dotset does not own.
	linkDest, err := uc.fs.Readlink(d.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("reading symlink: %w", err)
	}
	if linkDest != d.SourcePath {
		return nil, fmt.Errorf("%w: %s → %s (expected %s)", domain.ErrTargetNotLinked, d.TargetPath, linkDest, d.SourcePath)
	}

	// 4. Read the backup before mutating anything, so a failure leaves the
	// symlink and the backup untouched.
	backupExists, err := uc.fs.Exists(d.BackupPath)
	if err != nil {
		return nil, fmt.Errorf("checking backup: %w", err)
	}

	var backupData []byte
	if backupExists {
		backupData, err = uc.fs.ReadFile(d.BackupPath)
		if err != nil {
			return nil, fmt.Errorf("reading backup: %w", err)
		}
	}

	// 5. Remove the symlink.
	if err := uc.fs.Remove(d.TargetPath); err != nil {
		return nil, fmt.Errorf("removing symlink: %w", err)
	}

	result := &UndoResult{Dotfile: d, BackupMissing: !backupExists}

	// 6. Restore the original file and consume the backup.
	if backupExists {
		if err := uc.fs.WriteFile(d.TargetPath, backupData, 0o600); err != nil {
			return nil, fmt.Errorf("restoring backup (kept at %s): %w", d.BackupPath, err)
		}
		if err := uc.fs.Remove(d.BackupPath); err != nil {
			return nil, fmt.Errorf("removing consumed backup: %w", err)
		}
		result.Restored = true
	}

	return result, nil
}
