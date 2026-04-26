// Package main
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/prashanthsshetty/golsp/common/protocol"
)

func (c *Client) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		text := scanner.Text()

		if text == protocol.PingCommand {
			fmt.Fprintf(conn, "%s\n", protocol.PongReply)
			return
		}

		cmd, err := protocol.ParseCommand(text)
		if err != nil {
			fmt.Fprintf(conn, "ERROR: %v\n", err)
			fmt.Fprintf(conn, "%s\n", protocol.EOFMarker)
			return
		}

		c.executeCommand(cmd, conn)
	}
}

func (c *Client) executeCommand(cmd *protocol.Command, conn net.Conn) {
	f, _ := filepath.Abs(cmd.File)

	isNew := c.openFile(f) // handles open + change detection
	if isNew {
		c.waitForIndexing(f) // only wait on first open
	}

	reqID := c.send(cmd.Method, c.buildParams(cmd, f))

	c.mu.Lock()
	ch := c.pending[reqID]
	c.mu.Unlock()

	select {
	case res := <-ch:
		writeLocations(res, cmd.Method, conn)
	case <-time.After(15 * time.Second):
		fmt.Fprintf(conn, "ERROR: timeout\n")
		log.Printf("Command timed out: %s", cmd.Method)
	}

	fmt.Fprintf(conn, "%s\n", protocol.EOFMarker)
}

func (c *Client) buildParams(cmd *protocol.Command, f string) map[string]any {
	params := map[string]any{
		"textDocument": map[string]any{"uri": "file://" + f},
		"position": map[string]any{
			"line":      cmd.Line - 1,
			"character": cmd.Col - 1,
		},
	}
	if cmd.Method == "textDocument/references" {
		params["context"] = map[string]any{"includeDeclaration": true}
	}
	return params
}

func writeLocations(res *Response, method string, conn net.Conn) {
	switch method {
	case "textDocument/hover":
		var result struct {
			Contents struct {
				Value string `json:"value"`
			} `json:"contents"`
		}
		json.Unmarshal(res.Result, &result)
		fmt.Fprintf(conn, "%s\n", result.Contents.Value)

	default:
		var locs []Location
		if err := json.Unmarshal(res.Result, &locs); err != nil || len(locs) == 0 {
			var loc Location
			if err := json.Unmarshal(res.Result, &loc); err == nil && loc.URI != "" {
				locs = append(locs, loc)
			}
		}
		for _, l := range locs {
			p := strings.TrimPrefix(l.URI, "file://")
			fmt.Fprintf(conn, "%s:%d:%d\n", p, l.Range.Start.Line+1, l.Range.Start.Character+1)
		}
	}
}
