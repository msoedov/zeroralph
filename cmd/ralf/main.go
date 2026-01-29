package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Spinner for activity indication
type spinner struct {
	frames  []string
	current int
	message string
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
}

func newSpinner(message string) *spinner {
	return &spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (s *spinner) Start() {
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		defer close(s.done)

		for {
			select {
			case <-s.stop:
				fmt.Printf("\r\033[K")
				return
			case <-ticker.C:
				s.mu.Lock()
				frame := s.frames[s.current]
				msg := s.message
				s.current = (s.current + 1) % len(s.frames)
				s.mu.Unlock()
				fmt.Printf("\r%s%s%s %s", colorCyan, frame, colorReset, msg)
			}
		}
	}()
}

func (s *spinner) Stop() {
	close(s.stop)
	<-s.done
}

func (s *spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

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
		tool:          "claude",
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
  --tool     AI tool to use: amp or claude (default: claude)
  --version  Show version
  --help     Show this help

Arguments:
  max_iterations  Maximum iterations to run (default: 10)

Examples:
  ralf                    # Run with claude, 10 iterations
  ralf 20                 # Run with claude, 20 iterations
  ralf --tool amp         # Run with amp, 10 iterations
  ralf --tool amp 15      # Run with amp, 15 iterations`)
}

func getScriptDir() (string, error) {
	return os.Getwd()
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

func runTool(cfg *config) (string, error) {
	var cmd *exec.Cmd
	var stdinData []byte

	if cfg.tool == "amp" {
		promptPath := filepath.Join(cfg.scriptDir, "prompt.md")
		data, err := os.ReadFile(promptPath)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt.md: %w", err)
		}
		stdinData = data
		cmd = exec.Command("amp", "--dangerously-allow-all")
	} else {
		claudeMdPath := filepath.Join(cfg.scriptDir, "CLAUDE.md")
		data, err := os.ReadFile(claudeMdPath)
		if err != nil {
			return "", fmt.Errorf("failed to read CLAUDE.md: %w", err)
		}
		stdinData = data
		cmd = exec.Command("claude", "--dangerously-skip-permissions", "--print")
	}

	cmd.Dir = cfg.scriptDir

	var outputBuf bytes.Buffer
	teeWriter := io.MultiWriter(os.Stderr, &outputBuf)

	cmd.Stdout = teeWriter
	cmd.Stderr = teeWriter

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start %s: %w", cfg.tool, err)
	}

	stdin.Write(stdinData)
	stdin.Close()

	err = cmd.Wait()
	output := outputBuf.String()

	if err != nil {
		return output, fmt.Errorf("%s exited with error: %w", cfg.tool, err)
	}

	return output, nil
}

func containsCompletion(output string) bool {
	return strings.Contains(output, "<promise>COMPLETE</promise>")
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

// Logging helpers with colors
func logInfo(format string, args ...any) {
	fmt.Printf("%s[*]%s %s\n", colorBlue, colorReset, fmt.Sprintf(format, args...))
}

func logSuccess(format string, args ...any) {
	fmt.Printf("%s[+]%s %s\n", colorGreen, colorReset, fmt.Sprintf(format, args...))
}

func logWarning(format string, args ...any) {
	fmt.Printf("%s[!]%s %s\n", colorYellow, colorReset, fmt.Sprintf(format, args...))
}

func logError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s[-]%s %s\n", colorRed, colorReset, fmt.Sprintf(format, args...))
}

func logStep(step, total int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%d/%d]%s %s\n", colorCyan, step, total, colorReset, msg)
}

// Docker-like progress bar
func progressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	bar := strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		empty--
	}
	bar += strings.Repeat(" ", empty)
	percent := (current * 100) / total
	return fmt.Sprintf("[%s] %3d%%", bar, percent)
}

// Docker pull/build style status line
type statusLine struct {
	id      string
	status  string
	detail  string
	done    bool
	elapsed time.Duration
}

func printStatusLine(line statusLine) {
	checkmark := fmt.Sprintf("%s+%s", colorGreen, colorReset)
	if !line.done {
		checkmark = fmt.Sprintf("%s>%s", colorCyan, colorReset)
	}

	elapsed := ""
	if line.elapsed > 0 {
		elapsed = fmt.Sprintf(" %s%s%s", colorGray, line.elapsed.Round(time.Second), colorReset)
	}

	fmt.Printf(" %s %s%-12s%s %s%s\n", checkmark, colorBold, line.id, colorReset, line.status, elapsed)
}

func printBanner(tool string, maxIter int, project, branch string) {
	fmt.Printf("\n%s", colorCyan)
	fmt.Println(` ________  ___  ___  ___       ________ `)
	fmt.Println(`|\   __  \|\  \|\  \|\  \     |\  _____\`)
	fmt.Println(`\ \  \|\  \ \  \\\  \ \  \    \ \  \__/ `)
	fmt.Println(` \ \   _  _\ \  \\\  \ \  \    \ \   __\`)
	fmt.Println(`  \ \  \\  \\ \  \\\  \ \  \____\ \  \_|`)
	fmt.Println(`   \ \__\\ _\\ \_______\ \_______\ \__\ `)
	fmt.Println(`    \|__|\|__|\|_______|\|_______|\|__| `)
	fmt.Printf("%s\n", colorReset)
	fmt.Printf("  %sTool:%s      %s\n", colorGray, colorReset, tool)
	fmt.Printf("  %sProject:%s   %s\n", colorGray, colorReset, project)
	fmt.Printf("  %sBranch:%s    %s\n", colorGray, colorReset, branch)
	fmt.Printf("  %sMax iter:%s  %d\n\n", colorGray, colorReset, maxIter)
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	scriptDir, err := getScriptDir()
	if err != nil {
		logError("Getting script directory: %v", err)
		os.Exit(1)
	}
	cfg.scriptDir = scriptDir

	p, err := loadPRD(scriptDir)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	if err := archivePreviousRun(scriptDir, p); err != nil {
		logError("Archiving previous run: %v", err)
		os.Exit(1)
	}

	if p.BranchName != "" {
		if err := writeLastBranch(scriptDir, p.BranchName); err != nil {
			logError("Saving branch: %v", err)
			os.Exit(1)
		}
	}

	if err := initProgressFile(scriptDir); err != nil {
		logError("Initializing progress file: %v", err)
		os.Exit(1)
	}

	printBanner(cfg.tool, cfg.maxIterations, p.Project, p.BranchName)

	// Docker-style build header
	fmt.Printf("%s#1%s [internal] load configuration\n", colorGray, colorReset)
	printStatusLine(statusLine{id: "prd.json", status: "loaded", done: true})
	printStatusLine(statusLine{id: "CLAUDE.md", status: "loaded", done: true})
	fmt.Println()

	totalStart := time.Now()

	for i := 1; i <= cfg.maxIterations; i++ {
		// Docker build step style
		fmt.Printf("%s#%d%s %s %s\n", colorGray, i+1, colorReset, progressBar(i-1, cfg.maxIterations, 20), fmt.Sprintf("iteration %d/%d", i, cfg.maxIterations))

		startTime := time.Now()
		output, err := runTool(cfg)
		elapsed := time.Since(startTime)

		if err != nil {
			printStatusLine(statusLine{id: fmt.Sprintf("iter-%d", i), status: fmt.Sprintf("%swarning%s %v", colorYellow, colorReset, err), done: false, elapsed: elapsed})
		} else {
			printStatusLine(statusLine{id: fmt.Sprintf("iter-%d", i), status: "done", done: true, elapsed: elapsed})
		}

		if containsCompletion(output) {
			fmt.Println()
			fmt.Printf("%s#%d%s %s\n", colorGray, i+2, colorReset, "exporting results")
			printStatusLine(statusLine{id: "complete", status: fmt.Sprintf("%sCOMPLETE%s", colorGreen, colorReset), done: true})
			fmt.Println()
			totalElapsed := time.Since(totalStart)
			fmt.Printf(" %s=>%s %sfinished in %d iterations%s\n", colorGreen, colorReset, colorBold, i, colorReset)
			fmt.Printf("    %stotal time: %s%s\n\n", colorGray, totalElapsed.Round(time.Second), colorReset)
			os.Exit(0)
		}

		if i < cfg.maxIterations {
			// Docker-style waiting
			spin := newSpinner(fmt.Sprintf("%swaiting%s next iteration...", colorGray, colorReset))
			spin.Start()
			time.Sleep(2 * time.Second)
			spin.Stop()
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Printf("%s#%d%s %s\n", colorGray, cfg.maxIterations+2, colorReset, "exporting results")
	printStatusLine(statusLine{id: "incomplete", status: fmt.Sprintf("%smax iterations reached%s", colorYellow, colorReset), done: false})
	fmt.Println()
	totalElapsed := time.Since(totalStart)
	fmt.Printf(" %s=>%s %smax iterations reached (%d)%s\n", colorYellow, colorReset, colorBold, cfg.maxIterations, colorReset)
	fmt.Printf("    %stotal time: %s%s\n", colorGray, totalElapsed.Round(time.Second), colorReset)
	fmt.Printf("    %scheck progress.txt for status%s\n\n", colorGray, colorReset)
	os.Exit(1)
}
