# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

---

## [0.0.10] - 2026-01-23

### Fixed
- Fixed all golangci-lint errcheck issues across the codebase
- Properly handle return values from Printf, Println, Close, Setenv calls
- Migrated deprecated bubbletea mouse API to new Button/Action pattern

### Changed
- Removed docs folder from version control (kept locally for development)

---

## [0.0.9] - 2026-01-22

### Added
- Auto-update check: notifies when new version is available (non-blocking)
- `fp update` command to download and install latest version
- GitHub Actions workflows for CI and automated releases
- Interactive help browser with `fp help -i` flag (replaces `help-browser` command)
- Command suggestions for typos using Levenshtein distance
- Border color in theme system for interactive UI elements
- Watch interactive TUI mode
- Domain interfaces for dependency injection
- App factory for wiring application dependencies
- Tests for tracking actions (adopt, backfill, export)

### Changed
- Improved Makefile: added `lint`, `fmt`, `clean`, `install`, `integration` targets
- Replaced `help-browser` command with `-i` flag on `help`
- Refactored store package for cleaner database operations
- Simplified action dependencies

---

## [0.0.8] - 2026-01-21

### Added
- Theme system with `theme list`, `theme set`, and `theme pick` commands
- 8 built-in themes with dark/light variants: default, neon, aurora, mono, ocean, sunset, candy, contrast
- Application logging system with `logs` command
- `log_level` and `log_enabled` configuration options
- Interactive help browser for navigating commands and topics

### Changed
- Improved test coverage for logging, themes, and all packages

---

## [0.0.7] - 2026-01-19

### Added
- Database migrations system for schema updates
- `backfill` command to import historical commits from repositories
- Enriched CSV export with per-repo directory structure
- `--id` flag to `untrack` for removing orphaned repositories
- `--enrich` flag to `activity` and `watch` commands
- UI improvements with colored output
- Pager support with `--no-pager`, `--pager=<cmd>` flags and `pager` config
- `--flag=value` format support in addition to `--flag value`
- `ParsedFlags` for typed flag access
- Config defaults with code fallback
- Test scripts for hooks, export, and backfill

### Changed
- Renamed `log` command to `watch` for clarity
- Simplified store schema
- Handle ambiguous remotes in `track` command
- Improved error handling in dispatcher, activity, and record commands

### Fixed
- SQL bug in event queries
- Database connection leaks (added proper `db.Close()`)
- Removed dead code and improved variable naming

---

## [0.0.6] - 2026-01-16

### Added
- README and LICENSE documentation
- Base testing infrastructure for actions
- Help package with embedded topic documentation (`overview`, `workflow`, `hooks`, `data`)
- `EventFilter` struct for flexible event queries
- `log` command for streaming events in real time
- Detailed descriptions for all commands
- Topic resolution in help system
- Exit code 1 for bare `fp` invocation

### Changed
- Reorganized actions package into `config`, `setup`, and `tracking` subpackages
- Simplified hook script command from `repo record` to `record`
- Reorganized command categories to follow user journey
- Improved help output with git-like formatting
- Renamed `telemetry` package to `store`

---

## [0.0.5] - 2026-01-14

### Fixed
- Hooks installation issues

---

## [0.0.4] - 2026-01-13

### Added
- SQLite event storage for commits, merges, checkouts, rebases, and pushes
- `record` command (plumbing, invoked by hooks)
- `activity` command to view recorded events
- `export` command for CSV export

---

## [0.0.3] - 2026-01-13

### Added
- Repository tracking with `track`, `untrack`, `repos`, `status` commands
- Git hooks installation (`setup`, `teardown`, `check`)
- `sync-remote` command for updating repo IDs after URL changes

---

## [0.0.2] - 2026-01-12

### Fixed
- Makefile build process
- Changed App/info from const to var for version injection

---

## [0.0.1] - 2026-01-12

### Added
- Initial project structure
- CLI dispatcher with command tree
- `config` subcommands: `get`, `set`, `unset --all`, `list`
- `version` command with build-time injection

[Unreleased]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.10...HEAD
[0.0.10]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.9...v0.0.10
[0.0.9]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/footprint-tools/footprint-cli/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/footprint-tools/footprint-cli/releases/tag/v0.0.1
