// Package usecase implements the application's business logic (use cases).
// It depends only on the domain layer, following Clean Architecture's dependency rule.
package usecase

import (
	"errors"
	"fmt"

	"github.com/henrique/dotset/internal/domain"
)

// AdoptDotfileUseCase orchestrates the adoption of a dotfile:
//  1. Verifies the source folder exists in the current directory (dotfiles repo)
//     and contains exactly one regular file.
//  2. Checks the target in the home directory.
//  3. Backs up the original target file.
//  4. Creates a symlink from the home directory target to the source.
type AdoptDotfileUseCase struct {
	fs        domain.FileSystem
	homeDir   string
	backupDir string
}

// NewAdoptDotfileUseCase creates a new use case with the given filesystem adapter
// and configuration.
func NewAdoptDotfileUseCase(fs domain.FileSystem, homeDir, backupDir string) *AdoptDotfileUseCase {
	return &AdoptDotfileUseCase{
		fs:        fs,
		homeDir:   homeDir,
		backupDir: backupDir,
	}
}

// AdoptResult holds the result of an adopt operation.
type AdoptResult struct {
	Dotfile    *domain.Dotfile
	BackedUp   bool
	WasSymlink bool
}

// Execute runs the adopt use case for the given folder from workingDir.
// The folder lives inside workingDir and must contain exactly one regular file;
// that file's name determines the dotfile linked into the home directory.
func (uc *AdoptDotfileUseCase) Execute(folder, workingDir string) (*AdoptResult, error) {
	// 0. Locate the single regular file inside the source folder.
	d, err := resolveDotfile(uc.fs, folder, workingDir, uc.homeDir, uc.backupDir)
	if err != nil {
		return nil, err
	}

	// 1. Verify source exists in the source folder.
	exists, err := uc.fs.Exists(d.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("checking source: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("%w: %s", domain.ErrSourceNotFound, d.SourcePath)
	}

	srcInfo, err := uc.fs.Stat(d.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat source: %w", err)
	}
	if srcInfo.IsDir {
		return nil, fmt.Errorf("source %s is a directory — only files are supported", d.Name)
	}

	// 2. Check the target in home directory.
	result := &AdoptResult{Dotfile: d}

	targetExists, err := uc.fs.Exists(d.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("checking target: %w", err)
	}

	if targetExists {
		targetInfo, err := uc.fs.Stat(d.TargetPath)
		if err != nil {
			return nil, fmt.Errorf("stat target: %w", err)
		}

		if targetInfo.IsSymlink {
			// Check if it's already pointing to our source.
			linkDest, err := uc.fs.Readlink(d.TargetPath)
			if err != nil {
				return nil, fmt.Errorf("reading symlink: %w", err)
			}
			if linkDest == d.SourcePath {
				return result, errors.New("target is already linked to the same source — nothing to do")
			}
			result.WasSymlink = true
		}

		if targetInfo.IsDir {
			return nil, fmt.Errorf("%w: %s", domain.ErrTargetIsDir, d.TargetPath)
		}

		// 3. Backup the original file.
		if err := uc.backup(d); err != nil {
			return nil, fmt.Errorf("backup failed: %w", err)
		}
		result.BackedUp = true
	}

	// 4. Create symlink: target -> source.
	if targetExists {
		if err := uc.fs.Remove(d.TargetPath); err != nil {
			return nil, fmt.Errorf("removing target before symlink: %w", err)
		}
	}

	if err := uc.fs.Symlink(d.SourcePath, d.TargetPath); err != nil {
		return nil, fmt.Errorf("creating symlink: %w", err)
	}

	return result, nil
}

// backup copies the original target file to the backup directory.
func (uc *AdoptDotfileUseCase) backup(d *domain.Dotfile) error {
	if err := uc.fs.MkdirAll(uc.backupDir, 0o700); err != nil {
		return fmt.Errorf("creating backup dir: %w", err)
	}

	data, err := uc.fs.ReadFile(d.TargetPath)
	if err != nil {
		return fmt.Errorf("reading target for backup: %w", err)
	}

	if err := uc.fs.WriteFile(d.BackupPath, data, 0o600); err != nil {
		return fmt.Errorf("writing backup: %w", err)
	}

	return nil
}
