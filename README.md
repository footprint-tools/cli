# Footprint

Footprint is a local tool that records activity from Git repositories and stores it on your machine.

It helps you keep a structured history of work without using external services or sending data anywhere.

## What it does

Footprint records events that happen in Git repositories and saves them locally in a SQLite database.

Recorded events include commits, merges, checkouts, rebases, and pushes.

The data can be inspected, filtered, or exported later. Everything stays under your control.

## What it does not do

* Footprint does not upload data
* Footprint does not track time
* Footprint does not monitor behavior
* Footprint does not depend on online services

## Installation

Build from source:

```
make build
```

This creates the `fp` binary in the current directory.

## Quick start

```
fp setup              # Install git hooks (in current repo)
fp track              # Start tracking the current repository
```

From now on, footprint records activity automatically. You can inspect it with:

```
fp activity           # Show recorded events
fp watch              # Watch for new events in real time
```

## Global flags

These flags work with any command:

```
fp --no-color <command>        # Disable colored output
fp --no-pager <command>        # Do not use pager for output
fp --pager=<cmd> <command>     # Use specified pager for this command
```

## Commands

### Getting started

Install git hooks in the current repository:

```
fp setup
fp setup --force           # Skip confirmation prompt
```

Install git hooks globally (applies to all repositories):

```
fp setup --global
```

Start tracking a repository:

```
fp track [path]
fp track --remote=<name> [path]   # Use specific remote instead of 'origin'
```

### Inspecting activity

Show recorded activity (newest first):

```
fp activity
```

Filter activity with flags:

```
fp activity --oneline              # Compact one-line format
fp activity --since=2024-01-01     # Events after date
fp activity --until=2024-12-31     # Events before date
fp activity --status=<status>      # pending, exported, orphaned, skipped
fp activity --source=<source>      # post-commit, post-merge, post-checkout, post-rewrite, pre-push, manual, backfill
fp activity --repo=<id>            # Filter by repository
fp activity --limit=50             # Limit results
```

Watch for new events in real time:

```
fp watch
fp watch --oneline
```

This runs continuously like `tail -f`. Press Ctrl+C to stop.

### Repository status

Check tracking status of current repository:

```
fp status [path]
```

List all tracked repositories:

```
fp repos
fp list              # Alias for repos
```

### Managing repositories

Stop tracking a repository:

```
fp untrack [path]
fp untrack --id=<repo-id>    # Untrack by ID (useful for orphaned repos)
```

Update repository ID after remote URL changes:

```
fp sync-remote [path]
```

### Exporting data

Events are automatically exported to CSV files after recording. You can also trigger manually:

```
fp export --force            # Export immediately
fp export --dry-run          # Preview what would be exported
fp export --open             # Open export directory in file manager
```

**CSV Structure (year-based rotation):**

```
~/.config/Footprint/exports/
├── commits.csv          # Current year
├── commits-2024.csv     # Events from 2024
└── commits-2023.csv     # Events from 2023
```

Each CSV contains enriched data: authored_at, repo, branch, commit, subject, author, files, additions, deletions, parents, committer, committed_at, source, and machine.

**Sync to remote repository:**

```
fp config set export_remote git@github.com:user/my-exports.git
```

When configured, exports are automatically pushed to the remote.

**Configuration:**

| Key | Description |
|-----|-------------|
| `export_remote` | Remote URL for syncing exports |
| `export_interval` | Seconds between auto-exports (default: 3600) |
| `export_repo` | Path to local export repository |

### Importing historical data

Import existing commits from a repository:

```
fp backfill [path]
fp backfill --since=2024-01-01     # Import commits after date
fp backfill --until=2024-12-31     # Import commits before date
fp backfill --limit=100            # Limit number of commits
fp backfill --branch=<name>        # Use specific branch name for all commits
fp backfill --dry-run              # Preview what would be imported
```

Events are inserted with source "BACKFILL" and status "pending". Run `fp export --force` afterward to export the backfilled events.

### Managing hooks

Check installed hooks:

```
fp check
fp check --global
```

Remove hooks:

```
fp teardown
fp teardown --global
fp teardown --force          # Skip confirmation prompt
```

### Configuration

```
fp config list                  # Show all config values
fp config get <key>             # Get a value
fp config set <key> <value>     # Set a value
fp config unset <key>           # Remove a value
fp config unset --all           # Remove all values
```

Configuration is stored in `~/.fprc`.

Available configuration keys:

| Key | Description |
|-----|-------------|
| `pager` | Override the default pager. Set to `cat` to disable paging. |
| `export_remote` | Remote URL for syncing exports (configures git remote) |
| `export_interval` | Seconds between automatic exports (default: 3600) |
| `export_repo` | Path to export repository |
| `theme` | Color theme (default, neon, aurora, mono, ocean, sunset, candy, contrast) |
| `log_enabled` | Enable/disable logging (true/false) |
| `log_level` | Log verbosity (debug, info, warn, error) |

Pager precedence:
1. `--no-pager` flag → direct output
2. stdout not a TTY → direct output
3. `--pager=<cmd>` flag → uses specified pager, `cat` bypasses
4. `pager` config → uses configured pager, `cat` bypasses
5. `$PAGER` env var → uses env pager, `cat` bypasses
6. Default → `less -FRSX`

Pagers can include arguments: `fp config set pager "less -R"`

### Help

```
fp --help                # Show all commands
fp help <command>        # Help for a specific command
fp help <topic>          # Conceptual documentation
fp version               # Show version
```

Available help topics: `overview`, `workflow`, `hooks`, `data`, `configuration`, `troubleshooting`.

## Data storage

Events are stored in a SQLite database at:

- macOS: `~/Library/Application Support/Footprint/store.db`
- Linux: `~/.config/Footprint/store.db`

Tracked repositories are stored in `~/.fprc`.

## Privacy

All data stays on your machine.
Nothing is shared unless you choose to export it.
There is no telemetry.

## License

MIT License. See the LICENSE file for details.
