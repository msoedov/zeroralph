package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
