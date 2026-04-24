// Package main entrypoint
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prashanthsshetty/golsp/common/socket"
)

func startService(root string) {
	servicePath, err := exec.LookPath("golsp-server")
	if err != nil {
		home, _ := os.UserHomeDir()
		candidates := []string{
			filepath.Join(home, "go", "bin", "golsp-server"),
			filepath.Join(home, ".local", "bin", "golsp-server"),
			"/usr/local/bin/golsp-server",
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				servicePath = p
				break
			}
		}
	}
	if servicePath == "" {
		fmt.Println("ERROR: golsp-server not found in PATH or common locations")
		fmt.Println("PATH:", os.Getenv("PATH"))
		os.Exit(1)
	}
	cmd := exec.Command(servicePath, root)
	cmd.Dir = root

	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		fmt.Printf("ERROR starting service: %v\n", err)
		os.Exit(1)
	}

	err = cmd.Process.Release()
	if err != nil {
		fmt.Printf("ERROR in service: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: golsp-cli <command> <file> <line> <col> [root]")
		fmt.Println("Commands: references, implementations, definition, typedef, declaration, hover")
		return
	}

	command := os.Args[1]
	file := os.Args[2]
	line := os.Args[3]
	col := os.Args[4]
	root := os.Args[5]

	ctrlSock := socket.GetCtrlSocket(root)

	// 1. Check if the service for THIS project is running
	conn, err := net.DialTimeout("unix", ctrlSock, 50*time.Millisecond)
	if err != nil {
		// Not running for this project? Start it.
		startService(root)

		// Wait for the hashed socket to appear
		for range 20 {
			conn, err = net.Dial("unix", ctrlSock)
			if err == nil {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
	}

	if conn == nil {
		fmt.Println("Failed to connect to project service.")
		return
	}
	defer conn.Close()

	absPath, _ := filepath.Abs(file)
	request := fmt.Sprintf("%s %s %s %s\n", command, absPath, line, col)
	_, err = fmt.Fprint(conn, request)
	if err != nil {
		fmt.Printf("Failed to send the request. %+v", request)
		return
	}
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOF" {
			break
		}
		fmt.Println(line)
	}
}
