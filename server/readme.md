# golsp-server

`golsp-server` is a Go-based Language Server Protocol (LSP) proxy. It manages an underlying `gopls` instance and provides a Unix domain socket interface for communication, making it easier to integrate Go language features into custom editors or tools.

## Prerequisites

- **Go**: Version 1.18 or higher recommended.
- **gopls**: Ensure the official Go language server is installed and in your `PATH`.
    ```bash
    go install golang.org/x/tools/gopls@latest
    ```

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/prashanthsshetty/golsp.git
cd golsp
```

### 2. Build the Binary

To compile the program into an executable named `golsp-server`:

```bash
go build -o golsp-server .
```

### 3. Install to System Path

To make the command available globally:

```bash
go install .
```

_Note: This moves the binary to your `$GOPATH/bin` or `$HOME/go/bin`._

---

## Usage

### Starting the Server

By default, the server initializes in your current working directory. You can also specify a target directory as an argument.

```bash
# Start in current directory
./golsp-server

# Start in a specific project directory
./golsp-server /path/to/your/project
```

### How it Works

1.  **Logging**: The server automatically creates a log file in a project-specific path (determined by the `socket` package). It also pipes logs to `stderr`.
2.  **Control Socket**: It creates a Unix domain socket (e.g., `ctrl.sock`) used for LSP communication.
3.  **Process Management**: It handles system signals (`SIGTERM`, `SIGINT`) to ensure `gopls` and the socket connections are shut down gracefully.

---

## Configuration & Logs

The server uses the `github.com/prashanthsshetty/golsp` common library to determine socket locations and log paths.

- **Logs**: Check the console output or the generated log file in your workspace to debug connection issues.
- **Socket**: The socket path is dynamically generated based on the working directory to avoid conflicts between different projects.

## Troubleshooting

> **Error: Failed to start gopls**
> Ensure `gopls` is installed and accessible. Run `gopls version` in your terminal to verify.

> **Error: Failed to listen (Unix Socket)**
> The program attempts to clean up old sockets on startup. If you encounter permission errors, ensure you have write access to the directory or check if another instance is already running.
