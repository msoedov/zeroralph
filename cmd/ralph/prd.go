package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Default prd.json template
const defaultPRD = `{
  "project": "MyProject",
  "branchName": "ralph/feature",
  "description": "Feature description",
  "userStories": [
    {
      "id": "US-001",
      "title": "First task",
      "description": "As a developer, I need to implement...",
      "acceptanceCriteria": [
        "Acceptance criterion 1",
        "Typecheck passes"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
`

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

func loadPRD(workDir string) (*prd, bool, error) {
	prdPath := filepath.Join(workDir, "prd.json")

	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil, false, nil
	}

	data, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, true, fmt.Errorf("reading prd.json: %w", err)
	}

	var p prd
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, true, fmt.Errorf("parsing prd.json: %w", err)
	}

	return &p, true, nil
}

func initPRD(workDir string) error {
	prdPath := filepath.Join(workDir, "prd.json")

	if _, err := os.Stat(prdPath); err == nil {
		logInfo("prd.json already exists at %s", prdPath)
		return nil
	}

	logInfo("Creating prd.json at %s", prdPath)
	if err := os.WriteFile(prdPath, []byte(defaultPRD), 0644); err != nil {
		return fmt.Errorf("writing prd.json: %w", err)
	}

	logSuccess("Created prd.json - edit it with your project details")
	return nil
}

func initProgressFile(workDir string) error {
	progressPath := filepath.Join(workDir, "progress.txt")
	if _, err := os.Stat(progressPath); err == nil {
		logInfo("progress.txt exists at %s", progressPath)
		return nil
	}

	logInfo("Creating progress.txt at %s", progressPath)
	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC1123))
	return os.WriteFile(progressPath, []byte(content), 0644)
}

func resetProgressFile(workDir string) error {
	progressPath := filepath.Join(workDir, "progress.txt")
	logInfo("Resetting progress.txt")
	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC1123))
	return os.WriteFile(progressPath, []byte(content), 0644)
}

func readLastBranch(workDir string) string {
	data, err := os.ReadFile(filepath.Join(workDir, ".ralph-branch"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeLastBranch(workDir, branch string) error {
	return os.WriteFile(filepath.Join(workDir, ".ralph-branch"), []byte(branch), 0644)
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

func archivePreviousRun(workDir string, p *prd) error {
	lastBranch := readLastBranch(workDir)
	if lastBranch == "" {
		logInfo("No previous branch recorded")
		return nil
	}
	if lastBranch == p.BranchName {
		logInfo("Same branch as last run: %s", lastBranch)
		return nil
	}

	prdPath := filepath.Join(workDir, "prd.json")
	progressPath := filepath.Join(workDir, "progress.txt")

	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil
	}

	folderName := strings.TrimPrefix(lastBranch, "ralph/")
	archiveFolder := filepath.Join(workDir, "archive", time.Now().Format("2006-01-02")+"-"+folderName)

	logInfo("Branch changed: %s -> %s", lastBranch, p.BranchName)
	logInfo("Archiving previous run to %s", archiveFolder)

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

	return resetProgressFile(workDir)
}
