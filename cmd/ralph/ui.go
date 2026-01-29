package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorDim   = "\033[2m"

	// Orcish colors - WarCraft 3 inspired
	colorOrcBlood  = "\033[38;5;124m" // dark red
	colorOrcIron   = "\033[38;5;244m" // metallic gray
	colorOrcGold   = "\033[38;5;214m" // highlight gold
	colorOrcRust   = "\033[38;5;130m" // brownish rust
	colorOrcSparks = "\033[38;5;226m" // bright spark yellow

	// Semantic colors
	colorSuccess = colorOrcGold
	colorWarning = colorOrcRust
	colorError   = colorOrcBlood
	colorInfo    = colorOrcIron
	colorMuted   = "\033[38;5;240m"
	colorAccent  = colorOrcBlood
)

// Forge animation (Hammer & Anvil)
var forgeFrames = []string{
	"  🔨      ",
	"   🔨     ",
	"    🔨    ",
	"     🔨   ",
	"      🔨  ",
	"      ⛰  ",
	"      🔨  ",
	"     🔨   ",
	"    🔨    ",
	"   🔨     ",
}

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
		frames:  forgeFrames,
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
				fmt.Printf("\r%s%s%s %s", colorOrcIron, frame, colorReset, msg)
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
		marker = fmt.Sprintf("%s%s%s", colorSuccess, "ready", colorReset)
	} else {
		marker = fmt.Sprintf("%s%s%s", colorOrcIron, "forging", colorReset)
	}

	elapsed := ""
	if line.elapsed > 0 {
		elapsed = fmt.Sprintf(" %s%s%s", colorMuted, line.elapsed.Round(time.Second), colorReset)
	}

	fmt.Printf("  %s[%-6s]%s %s %s%s%s\n", colorOrcRust, line.id, colorReset, marker, colorDim, elapsed, colorReset)
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

func printBanner(tool string, maxIter int, p *prd, ver string) {
	fmt.Printf("\n%s", colorOrcIron)
	fmt.Println(`             /\          /\          /\`)
	fmt.Printf("      ______/  \\________/  \\________/  \\______\n")
	fmt.Printf("     |  %s██████╗  █████╗ ██╗   ██████╗ ██╗  ██╗%s  |\n", colorOrcBlood, colorOrcIron)
	fmt.Printf("     |  %s██╔══██╗██╔══██╗██║   ██╔══██╗██║  ██║%s  |\n", colorOrcBlood, colorOrcIron)
	fmt.Printf("     |  %s██████╔╝███████║██║   ██████╔╝███████║%s  |\n", colorOrcBlood, colorOrcIron)
	fmt.Printf("     |  %s██╔══██╗██╔══██║██║   ██╔═══╝ ██╔══██║%s  |\n", colorOrcBlood, colorOrcIron)
	fmt.Printf("     |  %s██║  ██║██║  ██║█████╗██║     ██║  ██║%s  |\n", colorOrcBlood, colorOrcIron)
	fmt.Printf("     |  %s╚═╝  ╚═╝╚═╝  ╚═╝╚════╝╚═╝     ╚═╝  ╚═╝%s  |\n", colorOrcBlood, colorOrcIron)
	fmt.Printf("      \\______    ________    ________    ______/\n")
	fmt.Println(`             \/          \/          \/`)
	fmt.Printf("%s\n", colorReset)
	fmt.Printf("  %s[TOOL  ]%s %s%s%s (v%s)\n", colorMuted, colorReset, colorBold, tool, colorReset, ver)
	fmt.Printf("  %s[GOAL  ]%s  %s\n", colorMuted, colorReset, p.Project)
	fmt.Printf("  %s[BRANCH]%s     %s%s%s\n", colorMuted, colorReset, colorOrcGold, p.BranchName, colorReset)
	fmt.Printf("  %s[LIMIT ]%s    %d\n\n", colorMuted, colorReset, maxIter)

	if len(p.UserStories) > 0 {
		fmt.Printf("  %sTASKS:%s\n", colorBold, colorReset)
		for _, s := range p.UserStories {
			status := " "
			if s.Passes {
				status = fmt.Sprintf("%s✔%s", colorSuccess, colorReset)
			} else {
				status = fmt.Sprintf("%s○%s", colorOrcIron, colorReset)
			}
			fmt.Printf("    [%s] %-8s %s\n", status, s.ID, s.Title)
		}
		fmt.Println()
	}
}

func logInfo(format string, args ...any) {
	fmt.Printf("  %s[ZUG ZUG]%s  %s\n", colorInfo, colorReset, fmt.Sprintf(format, args...))
}

func logSuccess(format string, args ...any) {
	fmt.Printf("  %s[DABU]%s     %s\n", colorSuccess, colorReset, fmt.Sprintf(format, args...))
}

func logWarning(format string, args ...any) {
	fmt.Printf("  %s[SWAB]%s     %s\n", colorWarning, colorReset, fmt.Sprintf(format, args...))
}

func logError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "  %s[LOK-TAR!]%s %s\n", colorError, colorReset, fmt.Sprintf(format, args...))
}

func logStep(step, total int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s[%d/%d]%s  %s\n", colorAccent, step, total, colorReset, msg)
}
