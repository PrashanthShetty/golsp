// Package socket
package socket

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
)

const (
	CtrlPrefix  = "/tmp/ctrl-"
	GoplsPrefix = "/tmp/gopls-"
	SocketExt   = ".sock"
)

func GetCtrlSocket(root string) string {
	return socketPath(CtrlPrefix, root)
}

func GetGoplsSocket(root string) string {
	return socketPath(GoplsPrefix, root)
}

func LogPath(root string) string {
	absPath, _ := filepath.Abs(root)
	hash := md5.Sum([]byte(absPath))
	shortHash := fmt.Sprintf("%x", hash)[:8]
	name := filepath.Base(root)
	return fmt.Sprintf("/tmp/golsp-%s-%s.log", name, shortHash)
}

func socketPath(prefix, root string) string {
	absPath, _ := filepath.Abs(root)
	hash := md5.Sum([]byte(absPath))
	shortHash := fmt.Sprintf("%x", hash)[:8]
	name := filepath.Base(root)
	return fmt.Sprintf("%s%s-%s%s", prefix, name, shortHash, SocketExt)
}
