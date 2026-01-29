package main

import (
	"fmt"
	"os"
	"strconv"
)

type config struct {
	command       string
	tool          string
	maxIterations int
	scriptDir     string
}

func parseArgs(args []string) (*config, error) {
	cfg := &config{
		command:       "run",
		tool:          "claude",
		maxIterations: 10,
	}

	i := 0
	if len(args) > 0 {
		switch args[0] {
		case "init":
			cfg.command = "init"
			i = 1
		case "run":
			cfg.command = "run"
			i = 1
		}
	}

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

Usage: ralf [command] [--tool amp|claude] [max_iterations]

Commands:
  run   Start the AI agent loop (default)
  init  Initialize workspace with templates (prd.json, CLAUDE.md, etc.)

Options:
  --tool     AI tool to use: amp or claude (default: claude)
  --version  Show version
  --help     Show this help

Arguments:
  max_iterations  Maximum iterations to run (default: 10)

Examples:
  ralf                    # Run with claude, 10 iterations
  ralf init               # Initialize workspace
  ralf 20                 # Run with claude, 20 iterations
  ralf --tool amp         # Run with amp, 10 iterations`)
}

func getScriptDir() (string, error) {
	return os.Getwd()
}
