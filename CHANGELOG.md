# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]


### Added

- Add shell completions command with auto-install support


### Changed

- update docs, improve test coverage
- Improve test coverage and rename demo target to simulate-activity


### Fixed

- Fix completions: create parent dirs, add to global setup, simplify messages


---
## [0.0.11] - 2026-01-26


### Added

- add database migrations system
- add defer db.Close() to prevent database leaks
- add config defaults with code fallback
- add backfill command to import historical commits
- add enriched export with per-repo CSV files
- add --id flag to untrack orphaned repos
- add UI improvements with color and pager support
- add test scripts for hooks, export, and backfill
- add ParsedFlags for typed flag access and remove duplicate helpers
- add themes configuration and log domain
- add tests for dispatchers package (99.6% coverage)
- add error case tests for log package (90.5% coverage)
- add tests for theme actions and refactor pick model (76.5% coverage)
- add tests for logs actions (71.4% coverage)
- add tests for ui/style package (98.2% coverage)
- add tests for tracking actions
- add tests for config package
- add tests for git, hooks, paths, repo and cli packages
- add tests for store package
- add tests for usage, ui pager and help topics
- add tests for setup actions and test utilities
- add help browser action for interactive help navigation
- Add Border color to theme system for interactive UI elements
- Add command suggestion algorithm using Levenshtein distance
- Add domain interfaces for dependency injection
- Add provider implementations for domain interfaces
- Add app factory for wiring application dependencies
- Add watch interactive TUI mode
- Add tests for tracking actions
- Add splitpanel UI component for consistent two-pane layouts
- Add UIActive and UIDim colors to theme system
- Add internal documentation for application and export flows
- Add git sync to export with offline mode and deduplication
- add shell quoting for safe path embedding in hooks
- add path traversal validation and support for git:// file:// URLs
- add input validation for git date args and commit hashes
- add RWMutex for thread-safe global logger access
- add thread-safe getter/setter for InteractiveBrowserFunc
- add mutex protection and connection cleanup for singleton
- add warning log on chmod failure
- add error handling for hook backup remove
- add committed flag pattern for transaction rollback handling
- add CloseDB helper that logs close errors
- add bounds checking for cursor and path slicing
- add cursor bounds validation in View method
- add cursor negative check and optimize buffer prepend
- add minimum height check to prevent underflow
- add signal handling with context cancellation for clean shutdown
- add goroutine to reap background process
- add retry logic, CSV validation, field count checks, and file sync
- add security note about pager command execution
- add auto-update check and release automation
- Add interactive logs viewer, configurable date/time display, and repo tracking fixes
- add interactive logs viewer and repos manager
- add repos scan command, improve setup/teardown with path args, clarify hooks docs


### Changed

- improve error handling in dispatcher, activity, and record commands
- handle ambiguous remotes in track command
- handle --flag=value format in dispatcher validation
- update dependencies
- simplified the store schema and add the --enrich flag to the watch and activity commands
- progress on coverage of logging and addes more themes
- update changelog for v0.0.8
- Replace help-browser command with interactive help flag (-i)
- Refactor store package for cleaner database operations
- Simplify action dependencies and minor fixes
- Update changelog for v0.0.8 and unreleased changes
- Update module dependencies
- Refactor theme picker with split panel UI
- Refactor interactive help with split panel UI
- Refactor watch view layout
- Refactor export with new CSV format and year-based rotation
- Update CLI tree and remove deprecated flags
- Update git utilities
- Improve config set and hooks install
- Update help topics for configuration and data
- Update README with new export format and configuration options
- Make hook scripts fail-safe with || true
- replace shell command with os.UserHomeDir()
- implement atomic config write with temp file and rename
- preserve inline comments when updating config values
- use parameterized query for LIMIT clause
- use CloseDB helper and add event ID validation
- update tests for new behavior
- update module path to github.com/footprint-tools/footprint-cli
- Fix lint issues, remove docs from tracking
- remove unused commands (track, untrack, status, sync-remote)
- improve test coverage across packages
- simplify and clarify all help text
- minor improvements to setup, theme, and UI components
- update changelog for v0.0.11


### Fixed

- fix SQL bug, remove dead code, improve variable names, and unify test script colors
- use Go 1.24 for CI compatibility
- fix: use Go 1.24 in CI
- configure git user in export test for CI environment


---
## [0.0.6] - 2026-01-16


### Added

- add README and LICENSE
- add EventFilter struct to support flexible event queries
- add help package with embedded topic documentation files
- add description field to commands for detailed help output
- add topic resolution to help and exit code 1 for bare invocation
- add detailed descriptions to all commands and update flag definitions
- add CLAUDE.md to gitignore


### Changed

- set base for testing of actions
- reorganize actions package into config, setup and tracking subpackages
- simplify hook script command from 'repo record' to 'record'
- reorganize command categories to follow user journey
- improve help output with git-like formatting and topic support
- handle resolution exit code in main
- rename telemetry package to store
- implemented fp log


---
## [0.0.5] - 2026-01-14


### Fixed

- fix hooks install


---
## [0.0.4] - 2026-01-13


### Changed

- implemented telemetry module so it saves events to sqlite db and prepare the process of them


---
## [0.0.3] - 2026-01-13


### Added

- add repo domain and telemetry


---
## [0.0.2] - 2026-01-13


### Fixed

- fix Makefile and make App/info var instead of const


---
## [0.0.1] - 2026-01-13


### Changed

- Initial commit
- Implemented al config sub-command: get, set, unset (--all), list
- implement version injection in build


---

[Unreleased]: https://github.com/footprint-tools/cli/compare/v0.0.11...HEAD
[0.0.11]: https://github.com/footprint-tools/cli/compare/v0.0.6...v0.0.11
[0.0.6]: https://github.com/footprint-tools/cli/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/footprint-tools/cli/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/footprint-tools/cli/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/footprint-tools/cli/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/footprint-tools/cli/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/footprint-tools/cli/releases/tag/v0.0.1
