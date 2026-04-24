// Package main
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/prashanthsshetty/golsp/common/socket"
)

func startGopls(cwd string) (*Client, error) {
	log.Printf("Starting gopls for: %s", cwd)
	sockPath := socket.GetGoplsSocket(cwd)
	cmd := exec.Command("gopls", "-mode=daemon", "-listen=unix;"+sockPath)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start gopls: %v", err)
	}
	log.Printf("gopls started PID: %d", cmd.Process.Pid)

	// Connect to gopls
	var conn net.Conn
	var err error
	for i := range 15 {
		conn, err = net.Dial("unix", sockPath)
		if err == nil {
			log.Printf("Connected to gopls on attempt %d", i+1)
			break
		}
		log.Printf("Attempt %d failed: %v", i+1, err)
		time.Sleep(200 * time.Millisecond)
	}

	if conn == nil {
		log.Fatal("Could not connect to gopls after 15 retries")
	}

	// go func() {
	// 	scanner := bufio.NewScanner(stderr)
	// 	for scanner.Scan() {
	// 		log.Printf("[gopls] %s", scanner.Text())
	// 	}
	// }()

	client := newClient(
		conn,
		cmd.Process,
		bufio.NewReader(conn),
		bufio.NewWriter(conn),
		cwd,
	)
	// go client.listen()
	return client, nil
}

func initializeClient(client *Client, cwd string) error {
	log.Printf("Initializing LSP...")
	id := client.send("initialize", map[string]any{
		"processId": os.Getpid(),
		"rootUri":   "file://" + cwd,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"references":     map[string]any{},
				"implementation": map[string]any{},
				"definition":     map[string]any{},
				"typeDefinition": map[string]any{},
			},
		},
	})

	client.mu.Lock()
	ch := client.pending[id]
	client.mu.Unlock()

	select {
	case <-ch:
		client.notify("initialized", map[string]any{})
		log.Printf("LSP initialized")
		return nil
	case <-time.After(15 * time.Second):
		return fmt.Errorf("initialize timed out")
	}
}
