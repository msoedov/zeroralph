package main

import (
	"fmt"
	"os"
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
