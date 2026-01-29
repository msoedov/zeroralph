package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var version = "dev"

type config struct {
	tool          string
	maxIterations int
	scriptDir     string
}

type prd struct {
	Project     string `json:"project"`
	BranchName  string `json:"branchName"`
	Description string `json:"description"`
}

func parseArgs(args []string) (*config, error) {
	cfg := &config{
		tool:          "amp",
		maxIterations: 10,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--version" || arg == "-v":
			fmt.Printf("ralf %s\n", version)
			os.Exit(0)
		case arg == "--help" || arg == "-h":
			printUsage()
			os.Exit(0)
		case arg == "--tool":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--tool requires a value")
			}
			i++
			cfg.tool = args[i]
		case len(arg) > 7 && arg[:7] == "--tool=":
			cfg.tool = arg[7:]
		default:
			if n, err := strconv.Atoi(arg); err == nil && n > 0 {
				cfg.maxIterations = n
			}
		}
		i++
	}

	if cfg.tool != "amp" && cfg.tool != "claude" {
		return nil, fmt.Errorf("invalid tool '%s': must be 'amp' or 'claude'", cfg.tool)
	}

	return cfg, nil
}

func printUsage() {
	fmt.Println(`ralf - autonomous AI agent loop

Usage: ralf [--tool amp|claude] [max_iterations]

Options:
  --tool     AI tool to use: amp or claude (default: amp)
  --version  Show version
  --help     Show this help

Arguments:
  max_iterations  Maximum iterations to run (default: 10)

Examples:
  ralf                    # Run with amp, 10 iterations
  ralf 20                 # Run with amp, 20 iterations
  ralf --tool claude      # Run with claude, 10 iterations
  ralf --tool claude 15   # Run with claude, 15 iterations`)
}

func getScriptDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
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

	fmt.Printf("Archiving previous run: %s\n", lastBranch)
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

	fmt.Printf("   Archived to: %s\n", archiveFolder)

	return resetProgressFile(scriptDir)
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	scriptDir, err := getScriptDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting script directory: %v\n", err)
		os.Exit(1)
	}
	cfg.scriptDir = scriptDir

	p, err := loadPRD(scriptDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := archivePreviousRun(scriptDir, p); err != nil {
		fmt.Fprintf(os.Stderr, "Error archiving previous run: %v\n", err)
		os.Exit(1)
	}

	if p.BranchName != "" {
		if err := writeLastBranch(scriptDir, p.BranchName); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving branch: %v\n", err)
			os.Exit(1)
		}
	}

	if err := initProgressFile(scriptDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing progress file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting Ralf - Tool: %s - Max iterations: %d\n", cfg.tool, cfg.maxIterations)
	fmt.Printf("Project: %s - Branch: %s\n", p.Project, p.BranchName)
}
