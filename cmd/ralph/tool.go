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

func getPrompt(tool string) string {
	if tool == "amp" {
		return ampPrompt
	}
	return claudePrompt
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
