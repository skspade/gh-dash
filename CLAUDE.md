# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gh-dash is a terminal UI dashboard for GitHub, built as a `gh` CLI extension. It displays PRs and issues in customizable sections with vim-style navigation. The project uses Go with the Bubbletea TUI framework.

## Common Commands

```bash
# Development (requires Devbox or manual Go setup)
task                    # Run the application
task debug              # Run with debug logging (writes to debug.log)
task logs               # Tail debug.log in a separate terminal

# Testing
task test               # Run all tests with prism
task test:one           # Run a single test interactively with gotip
task test:rerun         # Rerun the last test

# Code Quality
task lint               # Run golangci-lint
task lint-fix           # Run golangci-lint with --fix
task fmt                # Format code with gofumpt

# Debugging
go run gh-dash.go --debug  # Run with debug flag
tail -f debug.log          # Watch debug output

# Build & Install
go build -o gh-dash . && gh extension remove dash && gh extension install .
```

## Architecture

### Bubbletea/Elm Architecture
The app follows the Elm architecture pattern (Model-Update-View):
- **Model**: Application state defined in `internal/tui/ui.go` (`Model` struct)
- **Update**: Message handling in `Model.Update()` processes events and returns commands
- **View**: `Model.View()` renders the current state to the terminal

### Directory Structure
- `cmd/` - CLI entry point using Cobra
- `internal/config/` - YAML config parsing with koanf, supports global + per-repo configs
- `internal/data/` - GitHub GraphQL API queries via `go-gh` and `githubv4`
- `internal/tui/` - All TUI components
  - `components/` - Reusable UI components (sections, views, sidebar, etc.)
  - `context/` - Shared program context passed to all components
  - `keys/` - Keybinding definitions
  - `theme/` - Color and styling configuration

### Key Patterns
- **Sections**: PRs and Issues are displayed in configurable sections (`prssection`, `issuessection`)
- **Views**: Three view types - PRs, Issues, and Repo (branches)
- **Sidebar**: Details panel showing selected item info (`prview`, `issueview`, `branchsidebar`)
- **Context**: `context.ProgramContext` is passed to all components for shared state

### Data Flow
1. Config is loaded from `~/.config/gh-dash/config.yml` (or per-repo `.gh-dash.yml`)
2. Sections fetch data via GraphQL queries in `internal/data/`
3. Results are cached using `maypok86/otter`
4. UI updates via Bubbletea message passing

## Configuration

Config file locations (in priority order):
1. `--config` flag
2. `.gh-dash.yml` in current git repo
3. `$GH_DASH_CONFIG` env var
4. `$XDG_CONFIG_HOME/gh-dash/config.yml`

### Adding GitHub Actions

To enable the Actions view, add `actionsSections` to your config:

```yaml
actionsSections:
  - title: My Workflows
    repos:
      - owner/repo-name
      - owner/another-repo
    limit: 20  # optional
```

Press `s` to cycle views: PRs → Issues → Actions → Repo. The Actions view only appears when at least one section is configured.

## Development Environment

The project uses [Devbox](https://github.com/jetpack-io/devbox) for reproducible development:
```bash
devbox shell    # Enter dev environment with all tools
```

Tools provided: Go 1.23, golangci-lint, gofumpt, go-task, nerdfix, fd
