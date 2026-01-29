# Ralph Skill

Ralph is an autonomous AI agent loop that runs Claude or Amp repeatedly until all PRD tasks are complete.

## Overview

Ralph works by:
1. Reading a `prd.json` file containing user stories
2. Running an AI tool (Claude or Amp) with a specialized prompt
3. The AI picks the highest priority incomplete story and implements it
4. Repeating until all stories pass or max iterations reached

## Quick Start

```bash
# Install
go install github.com/msoedov/ralph/cmd/ralph@latest

# Initialize a new project
ralph init

# Edit prd.json with your tasks
# Then run ralph
ralph
```

## Commands

| Command | Description |
|---------|-------------|
| `ralph` | Run the agent loop (default: claude, 10 iterations) |
| `ralph init` | Create prd.json in current directory |
| `ralph prompt claude` | Print the Claude prompt |
| `ralph prompt amp` | Print the Amp prompt |
| `ralph --tool amp` | Use Amp instead of Claude |
| `ralph 20` | Run up to 20 iterations |

## PRD Format

The `prd.json` file defines your project and tasks:

```json
{
  "project": "my-feature",
  "branchName": "ralph/my-feature",
  "description": "Add the new feature",
  "userStories": [
    {
      "id": "US-001",
      "title": "Task title",
      "description": "What needs to be done",
      "acceptanceCriteria": ["Criterion 1", "Criterion 2"],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

### Fields

| Field | Description |
|-------|-------------|
| `project` | Project/feature name |
| `branchName` | Git branch for this work |
| `description` | High-level description |
| `userStories` | Array of tasks to complete |

### User Story Fields

| Field | Description |
|-------|-------------|
| `id` | Unique identifier (e.g., US-001) |
| `title` | Short title |
| `description` | Detailed description |
| `acceptanceCriteria` | List of requirements to pass |
| `priority` | Lower = higher priority (1 is first) |
| `passes` | Set to true when complete |
| `notes` | Optional notes |

## How the Agent Works

Each iteration, the AI agent:

1. Reads `prd.json` and `progress.txt`
2. Checks out the correct branch
3. Picks the highest priority story where `passes: false`
4. Implements the story
5. Runs quality checks (typecheck, lint, test)
6. Commits with message: `feat: [Story ID] - [Story Title]`
7. Updates `prd.json` to set `passes: true`
8. Appends progress to `progress.txt`
9. Outputs `<promise>COMPLETE</promise>` when all stories pass

## Files Created

| File | Description |
|------|-------------|
| `prd.json` | Your project configuration |
| `progress.txt` | Log of completed work and learnings |
| `archive/` | Old runs archived when branch changes |
| `.ralph-branch` | Tracks the current branch |

## Tips

### Keep Stories Small

Each story should be completable in one AI iteration (one context window). If a story is too large, break it into smaller pieces.

### Use Clear Acceptance Criteria

The AI uses acceptance criteria to verify completion. Be specific:

```json
{
  "acceptanceCriteria": [
    "Function `calculateTotal` exists in src/utils.ts",
    "Unit tests pass",
    "TypeScript compiles without errors"
  ]
}
```

### Priority Ordering

Stories are processed by priority (lowest number first). Use this to ensure dependencies are handled:

```json
{
  "userStories": [
    { "id": "US-001", "title": "Add database schema", "priority": 1 },
    { "id": "US-002", "title": "Add API endpoint", "priority": 2 },
    { "id": "US-003", "title": "Add UI component", "priority": 3 }
  ]
}
```

### Check Progress

The `progress.txt` file contains:
- Completed work per iteration
- Learnings and patterns discovered
- Useful context for future iterations

## Architecture

```
ralph/
  cmd/ralph/
    main.go         # Entry point and main loop
    config.go       # CLI argument parsing
    prd.go          # PRD loading and file management
    tool.go         # Tool execution
    tool_claude.go  # Claude prompt
    tool_amp.go     # Amp prompt
    ui.go           # Terminal UI (banner, spinner, colors)
    version.go      # Version string
```

## Differences from Original

This Go implementation of [snarktank/ralph](https://github.com/snarktank/ralph):

- Single binary, no external files needed
- Prompts embedded in the binary
- Works in any directory (files stored in working dir)
- `init` command to create prd.json
- `prompt` command to inspect prompts
- Docker-style progress UI
