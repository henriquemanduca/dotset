// Package domain defines the core entities and business rules of dotset.
// It is the innermost layer of Clean Architecture and has zero external dependencies.
package domain

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Common domain errors.
var (
	ErrSourceNotFound    = errors.New("source file not found in the source folder")
	ErrSourceDirNotFound = errors.New("source folder not found or not a directory")
	ErrTargetNotFound    = errors.New("target dotfile not found in home directory")
	ErrTargetIsDir       = errors.New("target is a directory, not a file")
	ErrTargetIsSymlink   = errors.New("target is already a symlink")
)

// Dotfile represents a dotfile being managed by dotset.
// It links a source file (from the dotfiles repository) to a target location
// (typically in the user's home directory), keeping a backup of the original.
type Dotfile struct {
	// Name is the filename (e.g., ".bashrc").
	Name string
	// SourcePath is the absolute path to the file in the dotfiles repository
	// (the current working directory where dotset is invoked).
	SourcePath string
	// TargetPath is the absolute path in the home directory (e.g., "/home/user/.bashrc").
	TargetPath string
	// BackupPath is the absolute path where the original file is backed up.
	BackupPath string
}

// NewDotfile creates a Dotfile from a filename, the current working directory,
// and the user's home directory.
func NewDotfile(name, workingDir, homeDir, backupDir string) (*Dotfile, error) {
	if name == "" {
		return nil, errors.New("filename must not be empty")
	}

	sourcePath := filepath.Join(workingDir, name)
	// Ensure the filename has a leading dot for home directory dotfiles.
	dotName := name
	if dotName[0] != '.' {
		dotName = "." + dotName
	}
	targetPath := filepath.Join(homeDir, dotName)
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.bak", dotName))

	return &Dotfile{
		Name:       name,
		SourcePath: sourcePath,
		TargetPath: targetPath,
		BackupPath: backupPath,
	}, nil
}
