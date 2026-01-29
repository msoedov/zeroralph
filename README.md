# ralf

Autonomous AI agent loop. Runs Claude or Amp in iterations until task completion.

## Installation

### From releases

Download the binary for your platform from [releases](https://github.com/user/ralf/releases).

### From source

```bash
go install github.com/user/ralf/cmd/ralf@latest
```

Or build locally:

```bash
go build -o ralf ./cmd/ralf/
```

## Usage

```bash
ralf [--tool amp|claude] [max_iterations]
```

### Options

- `--tool` - AI tool to use: `amp` or `claude` (default: `claude`)
- `--version`, `-v` - Show version
- `--help`, `-h` - Show help

### Arguments

- `max_iterations` - Maximum iterations to run (default: 10)

### Examples

```bash
ralf                    # Run with claude, 10 iterations
ralf 20                 # Run with claude, 20 iterations
ralf --tool amp         # Run with amp, 10 iterations
ralf --tool amp 15      # Run with amp, 15 iterations
```

## Configuration

ralf expects these files in the current directory:

- `prd.json` - Project configuration with fields:
  - `project` - Project name
  - `branchName` - Git branch name
  - `description` - Task description
- `CLAUDE.md` - Instructions for Claude (when using `--tool claude`)
- `prompt.md` - Instructions for Amp (when using `--tool amp`)

### prd.json example

```json
{
  "project": "my-project",
  "branchName": "feature/new-feature",
  "description": "Implement the new feature"
}
```

## How it works

1. Loads `prd.json` from current directory
2. Archives previous run if branch changed
3. Initializes `progress.txt` for tracking
4. Runs the AI tool in a loop
5. Checks output for `<promise>COMPLETE</promise>` marker
6. Exits on completion or max iterations

## Output files

- `progress.txt` - Progress log updated by the AI
- `archive/` - Previous runs archived by date and branch
- `.last-branch` - Tracks the last used branch

## Compatibility

ralf is a Go reimplementation of `scripts/ralph/ralph.sh`. It provides full feature parity plus enhancements:

| Feature | ralph.sh | ralf |
|---------|----------|------|
| Tool selection (`--tool amp\|claude`) | Yes | Yes |
| Max iterations argument | Yes | Yes |
| PRD loading (`prd.json`) | Yes | Yes |
| Branch-based archiving | Yes | Yes |
| Progress file initialization | Yes | Yes |
| Completion detection (`<promise>COMPLETE</promise>`) | Yes | Yes |
| 2-second sleep between iterations | Yes | Yes |
| Output tee to stderr | Yes | Yes |
| Exit codes (0 success, 1 max reached) | Yes | Yes |
| `--version` flag | No | Yes |
| `--help` flag | No | Yes |
| ASCII banner | No | Yes |
| Docker-style progress bar | No | Yes |
| Spinner animation | No | Yes |
| Colored output | No | Yes |
| Elapsed time tracking | No | Yes |

Default tool changed from `amp` to `claude`.
