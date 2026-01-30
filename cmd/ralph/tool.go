package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Prompts are defined in:
//   - tool_claude.go (claudePrompt)
//   - tool_amp.go (ampPrompt)
//
// Skills are defined in:
//   - skill_prd.go (skillPRD)
//   - skill_ralph.go (skillRalph)

func getPrompt(tool string) string {
	if tool == "amp" {
		return ampPrompt
	}
	return claudePrompt
}

func getSkill(name string) string {
	switch name {
	case "prd":
		return skillPRD
	case "ralph":
		return skillRalph
	default:
		return "Unknown skill: " + name + "\nAvailable skills: prd, ralph"
	}
}

func runTool(cfg *config) (string, error) {
	var cmd *exec.Cmd
	var stdinData string

	if cfg.tool == "amp" {
		stdinData = ampPrompt
		cmd = exec.Command("amp", "--dangerously-allow-all")
	} else {
		stdinData = claudePrompt
		cmd = exec.Command("claude", "--dangerously-skip-permissions", "--print")
	}

	cmd.Dir = cfg.workDir

	var outputBuf bytes.Buffer
	teeWriter := io.MultiWriter(os.Stderr, &outputBuf)

	cmd.Stdout = teeWriter
	cmd.Stderr = teeWriter

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	stdin.Write([]byte(stdinData))
	stdin.Close()

	err = cmd.Wait()
	output := outputBuf.String()

	return output, err
}

func containsCompletion(output string) bool {
	return strings.Contains(output, "<promise>COMPLETE</promise>")
}

func printSetup() {
	println("# Run the following commands:")
	println("mkdir -p ~/.claude/skills/ralph-prd")
	println("ralph skill prd > ~/.claude/skills/ralph-prd/SKILL.md")
	println("mkdir -p ~/.claude/skills/ralph-ralph")
	println("ralph skill ralph > ~/.claude/skills/ralph-ralph/SKILL.md")
}
