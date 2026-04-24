// Package main entry point
package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/prashanthsshetty/golsp/common/socket"
)

func main() {
	var cwd string
	if len(os.Args) > 1 {
		cwd = os.Args[1]
	} else {
		cwd, _ = os.Getwd()
	}

	// Setup logging
	logPath := socket.LogPath(cwd)
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Printf("=== golsp-server starting, root: %s ===", cwd)

	ctrlPath := socket.GetCtrlSocket(cwd)
	os.Remove(ctrlPath)

	// Start gopls
	client, err := startGopls(cwd)
	if err != nil {
		log.Fatalf("Failed to start gopls: %v", err)
	}

	go client.listen()

	// Initialize LSP
	if err := initializeClient(client, cwd); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Control socket
	l, err := net.Listen("unix", ctrlPath)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Ready at %s", ctrlPath)

	// Signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sig
		client.Shutdown()
		os.Exit(0)
	}()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go client.handleConn(c)
	}
}
