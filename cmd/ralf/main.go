package main

import (
	"fmt"
	"os"
	"time"
)

var version = "0.1.0"

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

	if cfg.command == "init" {
		if err := initMissingFiles(scriptDir); err != nil {
			logError("Initialization: %v", err)
			os.Exit(1)
		}
		logSuccess("Workspace initialized successfully.")
		os.Exit(0)
	}

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

	printBanner(cfg.tool, cfg.maxIterations, p, version)

	fmt.Printf("  %sforging%s  workspace\n", colorMuted, colorReset)
	printStatusLine(statusLine{id: "prd", done: true})
	printStatusLine(statusLine{id: "claude", done: true})
	fmt.Println()

	totalStart := time.Now()

	for i := 1; i <= cfg.maxIterations; i++ {
		fmt.Printf("  %s%d/%d%s    %s\n", colorAccent, i, cfg.maxIterations, colorReset, progressBar(i-1, cfg.maxIterations, 24))

		startTime := time.Now()
		spin := newSpinner(fmt.Sprintf("%srunning %s%s", colorMuted, cfg.tool, colorReset))
		spin.Start()
		output, err := runTool(cfg)
		spin.Stop()
		elapsed := time.Since(startTime)

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
