# zero ralph

Autonomous AI agent loop. Runs Claude or Amp in iterations until task completion.


This is a Go reimplementation of [snarktank/ralph](https://github.com/snarktank/ralph), rewritten for simplicity of setup - single binary, no external dependencies or template files.


Zero Ralph is a zero-config / file dependency version of Ralph.



<img src="docs/demo.png" width="400" alt="demo">

## Installation


### From source

```bash
go install github.com/msoedov/zeroralph/cmd/ralph@latest
```

Or build locally:

```bash
go install ./cmd/ralph/
```

## Usage

```bash
ralph [command] [--tool amp|claude] [max_iterations]
```

### Commands

- `run` - Start the AI agent loop (default)
- `prompt` - Print the embedded prompt for a tool (claude or amp)
- `skill` - Print a skill instruction (prd or ralph)
- `setup` - Print first-time setup commands for Claude skills
- `clean` - Remove prd.json, progress.txt, and .ralph-branch

### Options

- `--tool` - AI tool to use: `amp` or `claude` (default: `claude`)
- `--version`, `-v` - Show version
- `--help`, `-h` - Show help

### Arguments

- `max_iterations` - Maximum iterations to run (default: 10)

### Examples

```bash
ralph                    # Run with claude, 10 iterations
ralph setup              # Print first-time setup commands
ralph clean              # Remove progress files
ralph prompt claude      # Print the Claude prompt to stdout
ralph prompt amp         # Print the Amp prompt to stdout
ralph skill prd          # Print the PRD generator skill
ralph skill ralph        # Print the Ralph converter skill
ralph 20                 # Run with claude, 20 iterations
ralph --tool amp         # Run with amp, 10 iterations
```

## File Locations

All files are stored in the **current working directory** (where you run ralph):

| File | Description |
|------|-------------|
| `progress.txt` | Progress log (created automatically on first run) |
| `archive/` | Previous runs archived when branch changes |
| `.ralph-branch` | Tracks the last used branch |

### prd.json format

```json
{
  "project": "my-project",
  "branchName": "ralph/feature-name",
  "description": "Implement the new feature",
  "userStories": [
    {
      "id": "US-001",
      "title": "Add database schema",
      "description": "As a developer...",
      "acceptanceCriteria": ["Table exists", "Typecheck passes"],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

## How it works

1. Reads `prd.json` from current directory (warns if missing)
2. Archives previous run if branch changed (copies prd.json + progress.txt to archive/)
3. Creates/updates `progress.txt` for tracking
4. Runs the AI tool in a loop, piping the embedded prompt to stdin
5. Checks output for `<promise>COMPLETE</promise>` marker
6. Exits on completion or max iterations

## Embedded Prompts and Skills

All prompts and skills are embedded in the binary - no external files needed.

### Prompts

Use `ralph prompt claude` or `ralph prompt amp` to inspect the prompts passed to AI tools.

The prompts instruct the AI to:
- Read prd.json and progress.txt from the current directory
- Pick the highest priority user story where `passes: false`
- Implement that story, run quality checks, commit
- Update prd.json to mark the story as complete
- Append progress to progress.txt
- Output `<promise>COMPLETE</promise>` when all stories pass

### Skills

Skills are instruction sets you can pipe to AI tools for specific tasks:

| Skill | Description |
|-------|-------------|
| `prd` | Generate PRDs from feature descriptions |
| `ralph` | Convert existing PRDs to prd.json format |

Example usage with Claude:
```bash
ralph skill prd | claude
ralph skill ralph | claude
```

## Compatibility with original ralph

| Feature | ralph.sh | ralph (Go) |
|---------|----------|------------|
| Tool selection (`--tool`) | Yes | Yes |
| Max iterations argument | Yes | Yes |
| PRD loading | Yes | Yes |
| Branch-based archiving | Yes | Yes |
| Progress file | Yes | Yes |
| Completion detection | Yes | Yes |
| `prompt` command | No | Yes |
| `skill` command | No | Yes |
| `setup` command | No | Yes |
| `clean` command | No | Yes |
| Self-contained binary | No | Yes |
| ASCII banner & UI | No | Yes |
| Progress bar & Spinner | No | Yes |
| Colored output | No | Yes |

Default tool changed from `amp` to `claude`.

## Documentation

See [docs/SKILL.md](docs/SKILL.md) for detailed documentation including:
- PRD format and field descriptions
- How the agent works step by step
- Tips for writing effective user stories
- Architecture overview

## Project Structure

```
cmd/ralph/
  main.go           # Entry point, main loop
  config.go         # CLI parsing
  prd.go            # PRD/progress file handling
  tool.go           # Tool execution
  tool_claude.go    # Claude prompt (embedded)
  tool_amp.go       # Amp prompt (embedded)
  skill_prd.go      # PRD generator skill (embedded)
  skill_ralph.go    # Ralph converter skill (embedded)
  ui.go             # Terminal UI
  version.go        # Version
```
