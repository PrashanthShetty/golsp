# golsp: A Lightweight LSP Proxy for Go

`golsp` is a client-server toolset designed to provide fast Language Server Protocol (LSP) queries (like references and implementations) via Unix domain sockets. It is specifically optimized for integration with minimalist editors like **Zed** or **Neovim**.

## Architecture Overview

1.  **`golsp-server`**: A background service that wraps `gopls`. It maintains a persistent connection to the Go language server and exposes a project-specific Unix socket.
2.  **`golsp-cli`**: A stateless client that connects to the server’s socket, sends a query, and prints the results to `stdout`. It automatically starts the server if it isn't running.
3.  **Zed Integration**: Bash scripts that bridge the CLI with **fzf** and **Zed's Task system** for a seamless UI experience.

---

## Installation

### 1. Build and Install Binaries

Compile both components and ensure they are in your system path:

```bash
# Install the server
go install ./golsp-server

# Install the CLI
go install ./golsp-cli
```

### 2. Dependencies

Ensure the official Go language server is installed:

```bash
go install golang.org/x/tools/gopls@latest
```

---

## Zed Editor Integration

To use `golsp` inside the Zed editor for finding references or implementations, follow these steps:

### 1. Setup Helper Scripts

Copy your script files (e.g., `zed-fzf-golsp.sh` and `golsp-preview.sh`) to your Zed configuration folder and make them executable:

```bash
cp zed-fzf-golsp.sh golsp-preview.sh ~/.config/zed/
chmod +x ~/.config/zed/*.sh
```

### 2. Configure Zed Tasks

Add the following to your `tasks.json` in Zed:

```json
[
    {
        "label": "Find References",
        "command": "zed \"$(~/.config/zed/zed-fzf-golsp.sh references $ZED_FILE $ZED_ROW $ZED_COLUMN $ZED_WORKTREE_ROOT)\"",
        "tags": ["go"],
        "allow_concurrent_runs": false,
        "hide": "always",
        "use_new_terminal": false,
        "cwd": "$ZED_WORKTREE_ROOT"
    },
    {
        "label": "Find Implementations",
        "command": "zed \"$(~/.config/zed/zed-fzf-golsp.sh implementations $ZED_FILE $ZED_ROW $ZED_COLUMN $ZED_WORKTREE_ROOT)\"",
        "tags": ["go"],
        "allow_concurrent_runs": false,
        "hide": "always",
        "use_new_terminal": false,
        "cwd": "$ZED_WORKTREE_ROOT"
    }
]
```

### 3. Add Keyboard Shortcuts

Bind the tasks to keys in your `keymap.json`:

```json
[
    {
        "context": "Editor && mode == full",
        "bindings": {
            "cmd-shift-r": [
                "task::Spawn",
                {
                    "task_name": "Find References",
                    "reveal_target": "center"
                }
            ]
        }
    }
]
```

---

## Manual CLI Usage

You can also use the CLI manually for debugging or custom scripting:

```bash
golsp-cli <command> <file> <line> <col> <project_root>
```

**Supported Commands:**

- `references`
- `implementations`
- `definition`
- `typedef`
- `declaration`
- `hover`

---

## Technical Details

- **Socket Path**: The server creates a unique socket for every project based on a hash of the `<project_root>`. This allows you to work on multiple Go projects simultaneously without interference.
- **Auto-Cleanup**: The server handles `SIGTERM` and `SIGHUP` to clean up socket files and gracefully shut down the underlying `gopls` instance.
- **Logging**: Logs are automatically written to a path determined by the `common/socket` package, making it easy to troubleshoot initialization errors.
