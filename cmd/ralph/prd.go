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

func checkClaudeMD(workDir string) bool {
	locations := []string{
		filepath.Join(workDir, "CLAUDE.md"),
		filepath.Join(workDir, ".claude", "CLAUDE.md"),
	}
	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func cleanWorkDir(workDir string) error {
	files := []string{"prd.json", "progress.txt", ".ralph-branch"}
	removed := 0
	for _, f := range files {
		path := filepath.Join(workDir, f)
		err := os.Remove(path)
		if err == nil {
			logInfo("Removed %s", f)
			removed++
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", f, err)
		}
	}
	if removed > 0 {
		logSuccess("Cleaned working directory")
	} else {
		logInfo("Nothing to clean")
	}
	return nil
}
