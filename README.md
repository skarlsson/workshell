<p align="center">
  <img src="logo.svg" width="200" alt="workshell logo">
</p>

# workshell

CLI tool that orchestrates **kitty** sessions, **zellij** layouts, and **git** branches to manage multiple AI-assisted development workspaces.

Each workspace gets its own kitty window with a zellij layout (claude + build + shell panes), and git branch tracking for task management.

## Prerequisites

**Required** (core features):
- **kitty** — terminal emulator
- **zellij** — terminal multiplexer
- **git** — branch/task management

**Optional** (monitor/window management, GNOME-specific):
- **xdotool** — window move/focus/minimize (`ws focus`, `ws rotate`, `ws capture`)
- **xprop** — window title updates (cosmetic, silently skipped if missing)
- **gdbus** — monitor detection via GNOME/Mutter (`ws focus`, `ws rotate`)

No kitty configuration needed. `ws open` launches its own kitty instances with remote control enabled and per-workspace sockets at `$XDG_RUNTIME_DIR/kitty-ws-<name>` (falls back to `/tmp`).

## Install

```bash
curl -fsSL https://github.com/skarlsson/workshell/releases/latest/download/install.sh | bash
```

To update an existing install:

```bash
ws update
```

### Build from source

```bash
bash build.sh install
```

## Usage

### Dashboard

The dashboard is the primary interface for managing everything:

```bash
ws dashboard
```

From the dashboard you can create/open/close workspaces, manage tasks, configure settings (hosts, keybindings, dependencies), and more.

```
↑/↓: navigate  Enter: open  n: new  t: tasks  d: detach  x: kill  D: delete  s: settings  Esc: quit
```

### CLI commands

All dashboard actions are also available as subcommands:

```bash
ws open myproject                    # launch kitty + zellij session
ws close myproject                   # tear down session
ws list                              # show all workspaces
ws open clawdbot1:myproject          # open remote workspace via SSH
ws task start auth-refactor          # create + checkout task branch
ws task switch fix-bug               # stash + switch task
ws focus myproject                   # move window to work monitor
ws rotate                            # cycle workspaces on work monitor
```

See `ws --help` for the full list.

## Default Zellij Layout

```
┌──────────────────┬─────────────────┐
│                  │     build       │
│     claude       ├─────────────────┤
│                  │     shell       │
└──────────────────┴─────────────────┘
```

Set `auto_claude: false` in workspace config to get an editor pane instead.

Custom layouts can be placed in `~/.config/workshell/layouts/`.

## File Locations

| Path | Purpose |
|------|---------|
| `~/.config/workshell/config.yaml` | Global config (default layout, work monitor) |
| `~/.config/workshell/hosts.yaml` | Remote host definitions |
| `~/.config/workshell/workspaces/` | Workspace configs |
| `~/.config/workshell/layouts/` | Zellij layout files |
| `~/.local/state/workshell/` | Runtime state (window IDs, session names) |

## Building

```bash
bash build.sh           # build ./ws
bash build.sh install   # build + copy to ~/.local/bin/ws
VERSION=1.0.0 bash build.sh  # build with specific version
```
