// Package main
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Message struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type Response struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   any             `json:"error,omitempty"`
}

type OpenedFile struct {
	ModTime time.Time
}
type Location struct {
	URI   string `json:"uri"`
	Range struct {
		Start struct{ Line, Character int } `json:"start"`
	} `json:"range"`
}

// Client manages the LSP connection to gopls
type Client struct {
	conn         net.Conn
	goplsProcess *os.Process
	reader       *bufio.Reader
	writer       *bufio.Writer
	pending      map[int]chan *Response
	openedFiles  map[string]*OpenedFile
	mu           sync.Mutex
	idCounter    int
	projectDir   string
	sockPath     string
	ctrlPath     string
}

func newClient(conn net.Conn, process *os.Process, reader *bufio.Reader, writer *bufio.Writer, projectDir string) *Client {
	return &Client{
		conn:         conn,
		reader:       reader,
		writer:       writer,
		pending:      make(map[int]chan *Response),
		openedFiles:  make(map[string]*OpenedFile),
		projectDir:   projectDir,
		goplsProcess: process,
	}
}

func (c *Client) Shutdown() {
	if c.goplsProcess != nil {
		fmt.Printf("Shutting down gopls (PID: %d)...\n", c.goplsProcess.Pid)
		// Send SIGTERM to let gopls close gracefully
		c.goplsProcess.Signal(syscall.SIGTERM)
		os.Remove(c.ctrlPath)
		os.Remove(c.sockPath)
		os.Exit(0)
	}
}

func (c *Client) send(method string, params any) int {
	c.mu.Lock()
	c.idCounter++
	id := c.idCounter
	c.pending[id] = make(chan *Response, 1)
	c.mu.Unlock()

	msg := Message{Jsonrpc: "2.0", ID: id, Method: method, Params: params}
	body, _ := json.Marshal(msg)
	fmt.Fprintf(c.writer, "Content-Length: %d\r\n\r\n%s", len(body), body)
	c.writer.Flush()
	return id
}

func (c *Client) notify(method string, params any) {
	msg := Message{Jsonrpc: "2.0", Method: method, Params: params}
	body, _ := json.Marshal(msg)
	fmt.Fprintf(c.writer, "Content-Length: %d\r\n\r\n%s", len(body), body)
	c.writer.Flush()
}

// listen loop handles incoming data from gopls continuously
func (c *Client) listen() {
	for {
		resp, err := c.readOne()
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			return
		}
		if resp.ID > 0 {
			c.mu.Lock()
			ch, ok := c.pending[resp.ID]
			if ok {
				ch <- resp
				delete(c.pending, resp.ID)
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) readOne() (*Response, error) {
	var contentLength int
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			contentLength, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:")))
		}
	}
	body := make([]byte, contentLength)
	_, err := io.ReadFull(c.reader, body)
	if err != nil {
		return nil, err
	}
	var resp Response
	json.Unmarshal(body, &resp)
	return &resp, nil
}

func (c *Client) waitForIndexing(f string) {
	log.Printf("Waiting for gopls to index %s...", f)
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			log.Printf("Indexing timed out")
			return
		case <-ticker.C:
			reqID := c.send("textDocument/hover", map[string]any{
				"textDocument": map[string]any{"uri": "file://" + f},
				"position":     map[string]any{"line": 0, "character": 0},
			})

			c.mu.Lock()
			ch := c.pending[reqID]
			c.mu.Unlock()

			select {
			case resp := <-ch:
				if resp != nil {
					log.Printf("gopls ready!")
					return
				}
			case <-time.After(500 * time.Millisecond):
				log.Printf("Still indexing...")
			}
		}
	}
}

func (c *Client) openFile(f string) (isNew bool) {
	info, err := os.Stat(f)
	if err != nil {
		log.Printf("Failed to stat file: %v", err)
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	existing, alreadyOpen := c.openedFiles[f]

	if alreadyOpen {
		if info.ModTime().After(existing.ModTime) {
			// File changed — send didChange
			log.Printf("File changed, syncing: %s", f)
			c.syncFile(f, info.ModTime())
			return false // not new, but updated
		}
		return false // not new, not changed
	}

	// New file — send didOpen
	content, err := os.ReadFile(f)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return false
	}

	c.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri":        "file://" + f,
			"languageId": "go",
			"version":    1,
			"text":       string(content),
		},
	})

	c.openedFiles[f] = &OpenedFile{ModTime: info.ModTime()}
	log.Printf("Opened file: %s", f)
	return true
}

func (c *Client) syncFile(f string, modTime time.Time) {
	content, err := os.ReadFile(f)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return
	}

	// Increment version
	c.openedFiles[f].ModTime = modTime

	c.notify("textDocument/didChange", map[string]any{
		"textDocument": map[string]any{
			"uri":     "file://" + f,
			"version": 2,
		},
		"contentChanges": []map[string]any{
			{"text": string(content)}, // full file sync
		},
	})
	log.Printf("Synced file: %s", f)
}
