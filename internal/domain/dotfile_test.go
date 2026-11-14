package domain

import (
	"testing"
)

func TestNewDotfile(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		workingDir string
		homeDir    string
		backupDir  string
		wantErr    bool
		wantTarget string
		wantSource string
	}{
		{
			name:       "dotfile with leading dot",
			filename:   ".bashrc",
			workingDir: "/home/user/dotfiles",
			homeDir:    "/home/user",
			backupDir:  "/home/user/.local/state/dotset",
			wantTarget: "/home/user/.bashrc",
			wantSource: "/home/user/dotfiles/.bashrc",
		},
		{
			name:       "filename without leading dot (auto-prepend)",
			filename:   "bashrc",
			workingDir: "/home/user/dotfiles",
			homeDir:    "/home/user",
			backupDir:  "/home/user/.local/state/dotset",
			wantTarget: "/home/user/.bashrc",
			wantSource: "/home/user/dotfiles/bashrc",
		},
		{
			name:       "empty filename",
			filename:   "",
			workingDir: "/home/user/dotfiles",
			homeDir:    "/home/user",
			backupDir:  "/home/user/.local/state/dotset",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDotfile(tt.filename, tt.workingDir, tt.homeDir, tt.backupDir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.SourcePath != tt.wantSource {
				t.Errorf("SourcePath = %q, want %q", d.SourcePath, tt.wantSource)
			}
			if d.TargetPath != tt.wantTarget {
				t.Errorf("TargetPath = %q, want %q", d.TargetPath, tt.wantTarget)
			}
		})
	}
}
