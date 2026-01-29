package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantCmd       string
		wantTool      string
		wantMaxIter   int
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name:        "defaults",
			args:        []string{},
			wantCmd:     "run",
			wantTool:    "claude",
			wantMaxIter: 10,
		},
		{
			name:        "init command",
			args:        []string{"init"},
			wantCmd:     "init",
			wantTool:    "claude",
			wantMaxIter: 10,
		},
		{
			name:        "run command explicit",
			args:        []string{"run", "5"},
			wantCmd:     "run",
			wantTool:    "claude",
			wantMaxIter: 5,
		},
		{
			name:        "tool flag amp",
			args:        []string{"--tool", "amp"},
			wantCmd:     "run",
			wantTool:    "amp",
			wantMaxIter: 10,
		},
		{
			name:        "tool flag claude",
			args:        []string{"--tool", "claude"},
			wantCmd:     "run",
			wantTool:    "claude",
			wantMaxIter: 10,
		},
		{
			name:        "tool equals syntax",
			args:        []string{"--tool=claude"},
			wantCmd:     "run",
			wantTool:    "claude",
			wantMaxIter: 10,
		},
		{
			name:        "max iterations",
			args:        []string{"5"},
			wantCmd:     "run",
			wantTool:    "claude",
			wantMaxIter: 5,
		},
		{
			name:        "tool and iterations",
			args:        []string{"--tool", "claude", "20"},
			wantCmd:     "run",
			wantTool:    "claude",
			wantMaxIter: 20,
		},
		{
			name:        "skills command prd",
			args:        []string{"skills", "prd"},
			wantCmd:     "skills",
			wantTool:    "prd",
			wantMaxIter: 10,
		},
		{
			name:        "skills command ralph",
			args:        []string{"skills", "ralph"},
			wantCmd:     "skills",
			wantTool:    "ralph",
			wantMaxIter: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrSubstr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if cfg.command != tt.wantCmd {
				t.Errorf("command = %q, want %q", cfg.command, tt.wantCmd)
			}
			if cfg.tool != tt.wantTool {
				t.Errorf("tool = %q, want %q", cfg.tool, tt.wantTool)
			}
			if cfg.maxIterations != tt.wantMaxIter {
				t.Errorf("maxIterations = %d, want %d", cfg.maxIterations, tt.wantMaxIter)
			}
		})
	}
}

func TestLoadPRD(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("valid prd", func(t *testing.T) {
		prdPath := filepath.Join(tmpDir, "prd.json")
		os.WriteFile(prdPath, []byte(`{"project":"test","branchName":"main"}`), 0644)

		p, err := loadPRD(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Project != "test" {
			t.Errorf("project = %q, want %q", p.Project, "test")
		}
		if p.BranchName != "main" {
			t.Errorf("branchName = %q, want %q", p.BranchName, "main")
		}
	})

	t.Run("missing prd initializes files", func(t *testing.T) {
		emptyDir := t.TempDir()
		p, err := loadPRD(emptyDir)
		if err != nil {
			t.Errorf("expected auto-initialization, got error: %v", err)
		}
		if p.Project == "" {
			t.Error("expected initialized PRD project name to be non-empty")
		}

		// Verify files exist
		files := []string{"prd.json", "CLAUDE.md", "prompt.md", "AGENTS.md"}
		for _, f := range files {
			if _, err := os.Stat(filepath.Join(emptyDir, f)); os.IsNotExist(err) {
				t.Errorf("expected %s to be initialized", f)
			}
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidDir := t.TempDir()
		os.WriteFile(filepath.Join(invalidDir, "prd.json"), []byte(`{invalid`), 0644)
		_, err := loadPRD(invalidDir)
		if err == nil {
			t.Error("expected error for invalid json")
		}
	})
}

func TestInitProgressFile(t *testing.T) {
	t.Run("creates new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := initProgressFile(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		progressPath := filepath.Join(tmpDir, "progress.txt")
		data, err := os.ReadFile(progressPath)
		if err != nil {
			t.Fatalf("failed to read progress file: %v", err)
		}
		if len(data) == 0 {
			t.Error("progress file is empty")
		}
	})

	t.Run("preserves existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		progressPath := filepath.Join(tmpDir, "progress.txt")
		os.WriteFile(progressPath, []byte("existing content"), 0644)

		err := initProgressFile(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, _ := os.ReadFile(progressPath)
		if string(data) != "existing content" {
			t.Errorf("file was overwritten: got %q", string(data))
		}
	})
}

func TestArchiveDetection(t *testing.T) {
	t.Run("no archive when same branch", func(t *testing.T) {
		tmpDir := t.TempDir()
		writeLastBranch(tmpDir, "main")
		os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(`{}`), 0644)

		p := &prd{BranchName: "main"}
		err := archivePreviousRun(tmpDir, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		archiveDir := filepath.Join(tmpDir, "archive")
		if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
			t.Error("archive folder should not exist for same branch")
		}
	})

	t.Run("no archive when no last branch", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(`{}`), 0644)

		p := &prd{BranchName: "new-branch"}
		err := archivePreviousRun(tmpDir, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		archiveDir := filepath.Join(tmpDir, "archive")
		if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
			t.Error("archive folder should not exist when no previous branch")
		}
	})

	t.Run("archives on branch change", func(t *testing.T) {
		tmpDir := t.TempDir()
		writeLastBranch(tmpDir, "old-branch")
		os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(`{}`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "progress.txt"), []byte("old progress"), 0644)

		p := &prd{BranchName: "new-branch"}
		err := archivePreviousRun(tmpDir, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		archiveDir := filepath.Join(tmpDir, "archive")
		if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
			t.Error("archive folder should exist after branch change")
		}
	})
}

func TestContainsCompletion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "contains completion",
			output: "some output\n<promise>COMPLETE</promise>\nmore output",
			want:   true,
		},
		{
			name:   "no completion",
			output: "some output without marker",
			want:   false,
		},
		{
			name:   "partial marker",
			output: "<promise>INCOMPLET</promise>",
			want:   false,
		},
		{
			name:   "empty output",
			output: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsCompletion(tt.output)
			if got != tt.want {
				t.Errorf("containsCompletion(%q) = %v, want %v", tt.output, got, tt.want)
			}
		})
	}
}
