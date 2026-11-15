package adapter

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/henrique/dotset/internal/domain"
	"github.com/henrique/dotset/internal/usecase"
)

func sampleDotfile() *domain.Dotfile {
	return &domain.Dotfile{
		Name:       ".bashrc",
		SourcePath: "/repo/bash/.bashrc",
		TargetPath: "/home/user/.bashrc",
		BackupPath: "/home/user/.local/state/dotset/.bashrc.bak",
	}
}

func TestPrintSuccess(t *testing.T) {
	var out, errOut bytes.Buffer
	cli := NewCLI(&out, &errOut)

	cli.PrintSuccess(&usecase.AdoptResult{Dotfile: sampleDotfile(), BackedUp: true})

	got := out.String()
	for _, want := range []string{"Adopted .bashrc", "Source : /repo/bash/.bashrc", "Backup :"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q missing %q", got, want)
		}
	}
	if errOut.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", errOut.String())
	}
}

func TestPrintSuccessWithoutBackup(t *testing.T) {
	var out bytes.Buffer
	cli := NewCLI(&out, &bytes.Buffer{})

	cli.PrintSuccess(&usecase.AdoptResult{Dotfile: sampleDotfile()})

	if strings.Contains(out.String(), "Backup :") {
		t.Fatalf("did not expect a backup line: %q", out.String())
	}
}

func TestPrintUndoSuccess(t *testing.T) {
	var out bytes.Buffer
	cli := NewCLI(&out, &bytes.Buffer{})

	cli.PrintUndoSuccess(&usecase.UndoResult{Dotfile: sampleDotfile(), Restored: true})

	got := out.String()
	if !strings.Contains(got, "Unlinked .bashrc") || !strings.Contains(got, "Restored:") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPrintUndoSuccessWithoutBackup(t *testing.T) {
	var out bytes.Buffer
	cli := NewCLI(&out, &bytes.Buffer{})

	cli.PrintUndoSuccess(&usecase.UndoResult{Dotfile: sampleDotfile(), BackupMissing: true})

	if !strings.Contains(out.String(), "original not restored") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestPrintError(t *testing.T) {
	var out, errOut bytes.Buffer
	cli := NewCLI(&out, &errOut)

	cli.PrintError(errors.New("boom"))

	if out.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", out.String())
	}
	if got := errOut.String(); !strings.Contains(got, "dotset: boom") {
		t.Fatalf("unexpected stderr output: %q", got)
	}
}

func TestPrintVersion(t *testing.T) {
	var out bytes.Buffer
	cli := NewCLI(&out, &bytes.Buffer{})

	cli.PrintVersion("dotset v1.2.3")

	if got := strings.TrimSpace(out.String()); got != "dotset v1.2.3" {
		t.Fatalf("unexpected version output: %q", got)
	}
}
