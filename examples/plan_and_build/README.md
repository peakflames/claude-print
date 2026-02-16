# Plan & Build Example

Demonstrates a two-phase headless workflow using `claude-print`:

1. **Planning** — Opus creates a detailed plan with restricted permissions (`--permission-mode plan`)
2. **Building** — Sonnet executes the plan with full permissions

## Usage

```sh
./plan_and_build.sh "<task-description>"
```

The script automatically generates a plan file, creates the implementation plan, then executes it.

## Example

```sh
./plan_and_build.sh "Build a clone of the game Wordle. Stick to html, js, css, and no frameworks. Use multiple files and place source code in ./wordle/"
```

## How It Works

- **Phase 1:** Opus analyzes requirements, generates `<name>-plan.md` with detailed steps
- **Phase 2:** Sonnet reads the plan and implements it, using Task agents for parallel work
- Plan files are saved to `examples/plan_and_build/` for review

## Key Features

- Non-interactive, fully autonomous operation
- Permission-restricted planning prevents premature execution
- Model separation: expensive Opus for planning, efficient Sonnet for building
- Cross-platform compatible (handles Windows/Unix differences)