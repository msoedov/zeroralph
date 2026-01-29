package main

import (
	"fmt"
	"os"
	"time"
)

var version = "dev"

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
