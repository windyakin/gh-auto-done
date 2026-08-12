# gh-auto-done

A [gh](https://cli.github.com/) extension that automatically marks GitHub notifications as "Done" when their associated Pull Request or Issue has been closed or merged.

## Installation

```
gh extension install windyakin/gh-auto-done
```

## Usage

### Mark notifications as done

```
gh auto-done [--hostname <host>] [--dry-run]
```

Fetches your unread notifications, checks if each PR/Issue has been closed or merged, and marks matching notifications as done.

| Flag | Description |
|---|---|
| `--hostname` | Target a specific GitHub host |
| `--dry-run`, `-n` | Preview what would be done without making changes |

Example:

```
# github.com
gh auto-done

# GitHub Enterprise
gh auto-done --hostname git.pepabo.com

# Preview only
gh auto-done --dry-run
```

### Scheduled execution

Install a scheduled job to run `gh auto-done` periodically. The scheduler is selected automatically based on your OS:

| OS | Scheduler |
|---|---|
| macOS | launchd |
| Linux | systemd timer |
| Windows | Task Scheduler |

```
gh auto-done schedule install [--hostname <host>] [--interval <seconds>]
```

| Flag | Description |
|---|---|
| `--hostname` | GitHub host for the scheduled job (default: `github.com`) |
| `--interval` | Interval in seconds between runs (default: `300`) |

Example:

```
# Run every 5 minutes for github.com
gh auto-done schedule install
```

To remove the scheduled job:

```
gh auto-done schedule uninstall [--hostname <host>]
```

## Requirements

- [gh](https://cli.github.com/) CLI
- Authenticated via `gh auth login` (or `gh auth login --hostname <host>` for GHE)

## License

MIT
