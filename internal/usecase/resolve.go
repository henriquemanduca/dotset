package usecase

import (
	"fmt"
	"path/filepath"

	"github.com/henrique/dotset/internal/domain"
)

// resolveDotfile locates the single regular file inside folder (relative to
// workingDir) and builds the corresponding Dotfile. Both adopt and undo use it
// so the folder rules stay identical between the two operations.
func resolveDotfile(fs domain.FileSystem, folder, workingDir, homeDir, backupDir string) (*domain.Dotfile, error) {
	sourceDir := filepath.Join(workingDir, folder)

	// Verify the source folder exists and is a directory.
	dirInfo, err := fs.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("%w: %q not found in the current directory (pass a folder name)", domain.ErrSourceDirNotFound, folder)
	}
	if !dirInfo.IsDir {
		return nil, fmt.Errorf("%w: %q is not a folder (pass a folder name)", domain.ErrSourceDirNotFound, folder)
	}

	// Discover the single regular file inside the folder.
	entries, err := fs.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("reading source folder: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("%w: %s", domain.ErrSourceFolderEmpty, sourceDir)
	}
	if len(entries) > 1 {
		return nil, fmt.Errorf("%w: %s contains %d entries", domain.ErrSourceFolderMultiple, sourceDir, len(entries))
	}

	entry := entries[0]
	if entry.IsDir || entry.IsSymlink {
		return nil, fmt.Errorf("%w: %s", domain.ErrSourceNotRegularFile, entry.Path)
	}

	d, err := domain.NewDotfile(filepath.Base(entry.Path), sourceDir, homeDir, backupDir)
	if err != nil {
		return nil, fmt.Errorf("invalid dotfile: %w", err)
	}

	return d, nil
}
