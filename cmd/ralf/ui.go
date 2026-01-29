package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI color codes - modern palette inspired by popular CLI tools
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorDim   = "\033[2m"

	// Primary colors - softer, more readable
	colorRed     = "\033[38;5;203m" // soft red
	colorGreen   = "\033[38;5;114m" // muted green
	colorYellow  = "\033[38;5;221m" // warm yellow
	colorBlue    = "\033[38;5;75m"  // sky blue
	colorCyan    = "\033[38;5;80m"  // teal
	colorMagenta = "\033[38;5;177m" // soft purple
	colorGray    = "\033[38;5;245m" // medium gray

	// Semantic colors
	colorSuccess = "\033[38;5;114m" // green
	colorWarning = "\033[38;5;221m" // yellow
	colorError   = "\033[38;5;203m" // red
	colorInfo    = "\033[38;5;75m"  // blue
	colorMuted   = "\033[38;5;245m" // gray
	colorAccent  = "\033[38;5;80m"  // teal
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

// Docker pull/build style status line
type statusLine struct {
	id      string
	status  string
	detail  string
	done    bool
	elapsed time.Duration
}

func printStatusLine(line statusLine) {
	var marker string
	if line.done {
		marker = fmt.Sprintf("%s%s%s", colorSuccess, "done", colorReset)
	} else {
		marker = fmt.Sprintf("%s%s%s", colorAccent, "wait", colorReset)
	}

	elapsed := ""
	if line.elapsed > 0 {
		elapsed = fmt.Sprintf(" %s%s%s", colorMuted, line.elapsed.Round(time.Second), colorReset)
	}

	fmt.Printf("  %s%-6s%s %s%s%s%s\n", colorMuted, line.id, colorReset, marker, colorDim, elapsed, colorReset)
}

func progressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	bar := colorAccent + strings.Repeat("━", filled) + colorReset
	bar += colorMuted + strings.Repeat("─", empty) + colorReset
	percent := (current * 100) / total
	return fmt.Sprintf("%s %3d%%", bar, percent)
}

func printBanner(tool string, maxIter int, project, branch, ver string) {
	fmt.Printf("\n%s", colorAccent)
	fmt.Println(` ________  ___  ___  ___       ________ `)
	fmt.Println(`|\   __  \|\  \|\  \|\  \     |\  _____\`)
	fmt.Println(`\ \  \|\  \ \  \\\  \ \  \    \ \  \__/ `)
	fmt.Println(` \ \   _  _\ \  \\\  \ \  \    \ \   __\`)
	fmt.Println(`  \ \  \\  \\ \  \\\  \ \  \____\ \  \_|`)
	fmt.Println(`   \ \__\\ _\\ \_______\ \_______\ \__\ `)
	fmt.Println(`    \|__|\|__|\|_______|\|_______|\|__| `)
	fmt.Printf("%s\n", colorReset)
	fmt.Printf("  %sversion%s    %s\n", colorMuted, colorReset, ver)
	fmt.Printf("  %stool%s       %s%s%s\n", colorMuted, colorReset, colorBold, tool, colorReset)
	fmt.Printf("  %sproject%s    %s\n", colorMuted, colorReset, project)
	fmt.Printf("  %sbranch%s     %s%s%s\n", colorMuted, colorReset, colorAccent, branch, colorReset)
	fmt.Printf("  %smax iter%s   %d\n\n", colorMuted, colorReset, maxIter)
}

func logInfo(format string, args ...any) {
	fmt.Printf("  %sinfo%s  %s\n", colorInfo, colorReset, fmt.Sprintf(format, args...))
}

func logSuccess(format string, args ...any) {
	fmt.Printf("  %sdone%s  %s\n", colorSuccess, colorReset, fmt.Sprintf(format, args...))
}

func logWarning(format string, args ...any) {
	fmt.Printf("  %swarn%s  %s\n", colorWarning, colorReset, fmt.Sprintf(format, args...))
}

func logError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "  %sfail%s  %s\n", colorError, colorReset, fmt.Sprintf(format, args...))
}

func logStep(step, total int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s%d/%d%s  %s\n", colorAccent, step, total, colorReset, msg)
}
