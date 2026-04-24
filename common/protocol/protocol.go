// Package protocol
package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PingCommand = "ping"
	PongReply   = "pong"
	EOFMarker   = "EOF"
)

type Command struct {
	Method string
	File   string
	Line   int
	Col    int
}

var CommandMap = map[string]string{
	"references":      "textDocument/references",
	"implementations": "textDocument/implementation",
	"definition":      "textDocument/definition",
	"typedef":         "textDocument/typeDefinition",
	"declaration":     "textDocument/declaration",
	"hover":           "textDocument/hover",
}

func ParseCommand(text string) (*Command, error) {
	parts := strings.Fields(text)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid command: %s", text)
	}

	line, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid line: %s", parts[2])
	}

	col, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid col: %s", parts[3])
	}

	method, ok := CommandMap[parts[0]]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", parts[0])
	}

	return &Command{
		Method: method,
		File:   parts[1],
		Line:   line,
		Col:    col,
	}, nil
}
