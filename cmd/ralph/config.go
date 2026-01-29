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
	workDir       string // current working directory where prd.json/progress.txt live
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
		case "prompt":
			cfg.command = "prompt"
			i = 1
		case "skill":
			cfg.command = "skill"
			i = 1
		}
	}

	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--version" || arg == "-v":
			fmt.Printf("ralph %s\n", version)
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
			if cfg.command == "prompt" && cfg.tool == "claude" {
				// For prompt command, first positional argument is tool name
				cfg.tool = arg
			} else if cfg.command == "skill" && cfg.tool == "claude" {
				// For skill command, first positional argument is skill name
				cfg.tool = arg
			} else if n, err := strconv.Atoi(arg); err == nil && n > 0 {
				cfg.maxIterations = n
			}
		}
		i++
	}

	if cfg.command != "prompt" && cfg.command != "skill" && cfg.tool != "amp" && cfg.tool != "claude" {
		return nil, fmt.Errorf("invalid tool '%s': must be 'amp' or 'claude'", cfg.tool)
	}

	return cfg, nil
}

func printUsage() {
	fmt.Println(`ralph - autonomous AI agent loop

Usage: ralph [command] [--tool amp|claude] [max_iterations]

Commands:
  run       Start the AI agent loop (default)
  init      Initialize prd.json in current directory
  prompt    Print the prompt for a tool (claude or amp)
  skill     Print a skill instruction (prd or ralph)

Options:
  --tool     AI tool to use: amp or claude (default: claude)
  --version  Show version
  --help     Show this help

Arguments:
  max_iterations  Maximum iterations to run (default: 10)

Examples:
  ralph                    # Run with claude, 10 iterations
  ralph init               # Create prd.json in current directory
  ralph prompt claude      # Print the Claude prompt
  ralph prompt amp         # Print the Amp prompt
  ralph skill prd          # Print the PRD generator skill
  ralph skill ralph        # Print the Ralph converter skill
  ralph 20                 # Run with claude, 20 iterations
  ralph --tool amp         # Run with amp, 10 iterations

File Locations:
  prd.json      Current working directory (created by 'ralph init')
  progress.txt  Current working directory (created automatically)
  archive/      Current working directory (for archiving old runs)

Skills:
  prd     Generate PRDs from feature descriptions
  ralph   Convert PRDs to prd.json format

The prompts and skills are embedded in the binary.`)
}

func getWorkDir() (string, error) {
	return os.Getwd()
}
