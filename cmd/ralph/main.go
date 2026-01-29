package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	workDir, err := getWorkDir()
	if err != nil {
		logError("Getting working directory: %v", err)
		os.Exit(1)
	}
	cfg.workDir = workDir

	// Handle 'init' command
	if cfg.command == "init" {
		logInfo("Initializing ralph in %s", workDir)
		if err := initPRD(workDir); err != nil {
			logError("%v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle 'prompt' command
	if cfg.command == "prompt" {
		fmt.Println(getPrompt(cfg.tool))
		os.Exit(0)
	}

	// Run command - load PRD
	logInfo("Working directory: %s", workDir)

	p, exists, err := loadPRD(workDir)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	if !exists {
		logWarning("No prd.json found in %s", workDir)
		logInfo("Run 'ralph init' to create one, or create prd.json manually")
		logInfo("Continuing without PRD...")
		p = &prd{Project: "unknown", BranchName: "", Description: "No PRD"}
	} else {
		logSuccess("Loaded prd.json: project=%s branch=%s", p.Project, p.BranchName)
	}

	if p.BranchName != "" {
		if err := archivePreviousRun(workDir, p); err != nil {
			logError("Archiving previous run: %v", err)
			os.Exit(1)
		}

		if err := writeLastBranch(workDir, p.BranchName); err != nil {
			logError("Saving branch: %v", err)
			os.Exit(1)
		}
	}

	if err := initProgressFile(workDir); err != nil {
		logError("Initializing progress file: %v", err)
		os.Exit(1)
	}

	printBanner(cfg.tool, cfg.maxIterations, p, version)
	fmt.Println()

	totalStart := time.Now()

	for i := 1; i <= cfg.maxIterations; i++ {
		fmt.Printf("\n  %s%d/%d%s    %s\n", colorAccent, i, cfg.maxIterations, colorReset, progressBar(i-1, cfg.maxIterations, 24))

		startTime := time.Now()
		spin := newSpinner(fmt.Sprintf("%srunning %s%s", colorMuted, cfg.tool, colorReset))
		spin.Start()
		output, err := runTool(cfg)
		spin.Stop()
		elapsed := time.Since(startTime)

		// Print status on new line after spinner clears
		fmt.Println()
		if err != nil {
			printStatusLine(statusLine{id: fmt.Sprintf("iter%d", i), done: false, elapsed: elapsed})
		} else {
			printStatusLine(statusLine{id: fmt.Sprintf("iter%d", i), done: true, elapsed: elapsed})
		}

		if containsCompletion(output) {
			fmt.Println()
			totalElapsed := time.Since(totalStart)
			fmt.Printf("  %scomplete%s  finished in %d iterations\n", colorSuccess, colorReset, i)
			fmt.Printf("  %s          %s%s\n\n", colorMuted, totalElapsed.Round(time.Second), colorReset)
			os.Exit(0)
		}

		if i < cfg.maxIterations {
			spin := newSpinner(fmt.Sprintf("%swaiting%s", colorMuted, colorReset))
			spin.Start()
			time.Sleep(2 * time.Second)
			spin.Stop()
			fmt.Println()
		}
	}

	fmt.Println()
	totalElapsed := time.Since(totalStart)
	fmt.Printf("  %stimeout%s   max iterations reached (%d)\n", colorWarning, colorReset, cfg.maxIterations)
	fmt.Printf("  %s          %s%s\n", colorMuted, totalElapsed.Round(time.Second), colorReset)
	fmt.Printf("  %s          check progress.txt%s\n\n", colorMuted, colorReset)
	os.Exit(1)
}
