# Footprint

Local Git activity tracker. Records commits, merges, checkouts, rebases, and pushes to a SQLite database on your machine.

## Install

```bash
# Build from source
make build

# Or install to GOPATH/bin
make install
```

## Quick Start

```bash
fp setup                  # Install hooks in current repo
fp activity               # View recorded events
```

## Commands

### Getting Started

```bash
fp setup                     # Install hooks in current repo
fp setup ~/projects/myapp    # Install in specific repo
fp setup --core-hooks-path   # Set global hooks (git core.hooksPath)
fp repos check               # Verify hooks are installed
```

### View Activity

```bash
fp activity                  # Show recent events
fp activity -n 50            # Limit to 50 events
fp activity -e               # Include commit messages
fp activity --repo <id>      # Filter by repository

fp watch                     # Stream events in real time
fp watch -i                  # Interactive dashboard
```

### Manage Repositories

```bash
fp repos list                # List repos with activity
fp repos scan                # Find repos and show hook status
fp repos scan --root ~/dev   # Scan from specific path
fp repos check               # Verify hooks in current repo
fp repos -i                  # Interactive hook manager

fp teardown                  # Remove hooks from current repo
fp teardown ~/projects/app   # Remove from specific repo
```

### Import History

```bash
fp backfill                  # Import all past commits
fp backfill --since 2024-01-01
fp backfill --limit 100
fp backfill --dry-run        # Preview only
```

### Export Data

```bash
fp export --now              # Export immediately
fp export --dry-run          # Preview
fp export --open             # Open export folder
```

Exports go to `~/.config/Footprint/exports/` as CSV files.

### Configuration

```bash
fp config list               # Show all settings
fp config get <key>          # Get a value
fp config set <key> <value>  # Set a value
fp config unset <key>        # Remove a value
fp config -i                 # Interactive settings editor
```

Settings:

| Key | Description |
|-----|-------------|
| `theme` | Color theme (neon-dark, ocean-light, etc.) |
| `display_date` | Date format (dd/mm/yyyy, mm/dd/yyyy, yyyy-mm-dd) |
| `display_time` | Time format (12h, 24h) |
| `pager` | Pager command (default: less -FRSX) |
| `enable_log` | Enable logging (true/false) |

### Themes

```bash
fp theme list                # Show available themes
fp theme set neon-dark       # Apply a theme
fp theme -i                  # Interactive theme picker
```

Themes: default, neon, aurora, mono, ocean, sunset, candy, contrast (each with -dark/-light variants)

### Other

```bash
fp version                   # Show version
fp update                    # Update to latest version
fp logs                      # View fp logs
fp logs -i                   # Interactive log viewer
fp help                      # Show help
fp help -i                   # Interactive help browser
```

## Global Flags

```bash
fp --no-color <command>      # Disable colors
fp --no-pager <command>      # Disable pager
fp --pager=<cmd> <command>   # Use specific pager
```

## Data Storage

- Database: `~/.config/Footprint/store.db`
- Config: `~/.fprc`
- Exports: `~/.config/Footprint/exports/`
- Logs: `~/.config/Footprint/fp.log`

## Privacy

All data stays local. No telemetry. No network requests except `fp update`.

## License

MIT
