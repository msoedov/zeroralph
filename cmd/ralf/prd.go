package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/*
var templates embed.FS

type userStory struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Priority           int      `json:"priority"`
	Passes             bool     `json:"passes"`
	Notes              string   `json:"notes"`
}

type prd struct {
	Project     string      `json:"project"`
	BranchName  string      `json:"branchName"`
	Description string      `json:"description"`
	UserStories []userStory `json:"userStories"`
}

func loadPRD(scriptDir string) (*prd, error) {
	prdPath := filepath.Join(scriptDir, "prd.json")
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		if err := initMissingFiles(scriptDir); err != nil {
			return nil, fmt.Errorf("initializing missing files: %w", err)
		}
	}

	data, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, err
	}

	var p prd
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid prd.json: %w", err)
	}

	return &p, nil
}

func initMissingFiles(scriptDir string) error {
	// Files to initialize if missing
	files := []struct {
		name     string
		template string
	}{
		{"prd.json", "templates/prd.json"},
		{"CLAUDE.md", "templates/CLAUDE.md"},
		{"prompt.md", "templates/prompt.md"},
		{"AGENTS.md", "templates/AGENTS.md"},
	}

	for _, f := range files {
		targetPath := filepath.Join(scriptDir, f.name)
		if _, err := os.Stat(targetPath); err == nil {
			continue // Already exists
		}

		logInfo("Initializing %s...", f.name)
		content, err := templates.ReadFile(f.template)
		if err != nil {
			return fmt.Errorf("reading template for %s: %w", f.name, err)
		}

		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", f.name, err)
		}
	}

	return nil
}

func initProgressFile(scriptDir string) error {
	progressPath := filepath.Join(scriptDir, "progress.txt")
	if _, err := os.Stat(progressPath); err == nil {
		return nil
	}

	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC1123))
	return os.WriteFile(progressPath, []byte(content), 0644)
}

func resetProgressFile(scriptDir string) error {
	progressPath := filepath.Join(scriptDir, "progress.txt")
	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC1123))
	return os.WriteFile(progressPath, []byte(content), 0644)
}

func readLastBranch(scriptDir string) string {
	data, err := os.ReadFile(filepath.Join(scriptDir, ".last-branch"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeLastBranch(scriptDir, branch string) error {
	return os.WriteFile(filepath.Join(scriptDir, ".last-branch"), []byte(branch), 0644)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func archivePreviousRun(scriptDir string, p *prd) error {
	lastBranch := readLastBranch(scriptDir)
	if lastBranch == "" || lastBranch == p.BranchName {
		return nil
	}

	prdPath := filepath.Join(scriptDir, "prd.json")
	progressPath := filepath.Join(scriptDir, "progress.txt")

	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil
	}

	folderName := strings.TrimPrefix(lastBranch, "ralph/")
	archiveFolder := filepath.Join(scriptDir, "archive", time.Now().Format("2006-01-02")+"-"+folderName)

	logInfo("Archiving previous run: %s", lastBranch)
	if err := os.MkdirAll(archiveFolder, 0755); err != nil {
		return err
	}

	if err := copyFile(prdPath, filepath.Join(archiveFolder, "prd.json")); err != nil {
		return err
	}
	if _, err := os.Stat(progressPath); err == nil {
		if err := copyFile(progressPath, filepath.Join(archiveFolder, "progress.txt")); err != nil {
			return err
		}
	}

	logSuccess("Archived to: %s", archiveFolder)

	return resetProgressFile(scriptDir)
}
