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

type prd struct {
	Project     string `json:"project"`
	BranchName  string `json:"branchName"`
	Description string `json:"description"`
}

func loadPRD(scriptDir string) (*prd, error) {
	prdPath := filepath.Join(scriptDir, "prd.json")
	data, err := os.ReadFile(prdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("prd.json not found in %s", scriptDir)
		}
		return nil, err
	}

	var p prd
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid prd.json: %w", err)
	}

	return &p, nil
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
