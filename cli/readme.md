# golsp-cli

`golsp-cli` is a command-line interface designed to interact with the `golsp-server`. It allows users to query Go language symbols directly from the terminal or via editor integrations (like Vim, Emacs, or VS Code).

## Key Features

- **Auto-Managed Server**: If the backend server for a specific project isn't running, the CLI automatically spawns a background `golsp-server` process.
- **Project Isolation**: Uses project-specific Unix sockets to ensure queries are handled by the correct server instance for that specific codebase.
- **LSP Query Support**: Supports standard Language Server Protocol actions including Definition, References, and Hover.

## Installation

### 1. Build the CLI

```bash
go build -o golsp-cli .
```

### 2. Install

Move the binary to your path:

```bash
go install .
```

### 3. Ensure Server Availability

The CLI expects the `golsp-server` binary to be present in your system. It searches in the following order:

1. Your system `$PATH`
2. `$HOME/go/bin/golsp-server`
3. `$HOME/.local/bin/golsp-server`
4. `/usr/local/bin/golsp-server`

## Usage

The CLI requires specific arguments to pinpoint the code location you are querying.

### Command Syntax

```bash
golsp-cli <command> <file_path> <line> <column> <project_root>
```

### Available Commands

- `definition`: Jump to the symbol definition.
- `references`: List all usages of the symbol.
- `implementations`: Find interface implementations.
- `typedef`: Find the type definition.
- `hover`: Get documentation/type info for the symbol under the cursor.

### Example

To find the definition of a function at line 15, column 10 in `main.go`:

```bash
golsp-cli definition ./main.go 15 10 /home/user/my-go-project
```

## How It Works

1.  **Socket Lookup**: The CLI calculates a unique Unix socket path based on the provided `<project_root>`.
2.  **Connection Check**: It attempts to connect to the socket. If the connection fails (server not running), it launches `golsp-server` in the background as a detached process.
3.  **Request**: It sends a plain-text request formatted as: `command file_path line col`.
4.  **Response**: It reads the server output and prints it to `stdout` until an `EOF` signal is received.

## Troubleshooting

- **"Failed to connect to project service"**: This usually means the server failed to start. Check if `gopls` is installed on your system, as the server depends on it.
- **Incorrect Results**: Ensure the `<project_root>` provided is the absolute path to the directory containing the `go.mod` file.
- **Logs**: Check the server logs (usually located in a temp directory or the project root depending on your `socket` package configuration) for internal errors.
