// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoHandlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sylve/internal/logger"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WindowSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
	X    uint16
	Y    uint16
}

var WSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleTerminalWebsocket(c *gin.Context) {
	_, exists := c.Get("Token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	checkSession := exec.Command("tmux", "has-session", "-t", id)
	if err := checkSession.Run(); err != nil {
		createSession := exec.Command("tmux", "new-session", "-s", id, "-d")
		if err := createSession.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create tmux session"})
			return
		}
	}

	w := c.Writer
	r := c.Request
	conn, err := WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.L.Error().Msgf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	var wsWriteMu sync.Mutex

	safeWrite := func(messageType int, data []byte) error {
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		return conn.WriteMessage(messageType, data)
	}

	cmd := exec.Command("tmux", "attach-session", "-t", id)
	cmd.Env = append(os.Environ(), "TERM=xterm")

	tty, err := pty.Start(cmd)
	if err != nil {
		safeWrite(websocket.TextMessage, []byte(err.Error()))
		return
	}

	defer func() {
		cmd.Process.Kill()
		cmd.Process.Wait()
		tty.Close()
	}()

	done := make(chan struct{})
	defer close(done)

	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-done:
				return
			default:
				n, err := tty.Read(buf)
				if err != nil {
					safeWrite(websocket.TextMessage, []byte("Terminal session closed."))
					return
				}
				if err := safeWrite(websocket.BinaryMessage, buf[:n]); err != nil {
					return
				}
			}
		}
	}()

	for {
		messageType, reader, err := conn.NextReader()
		if err != nil {
			return
		}

		if messageType == websocket.TextMessage {
			safeWrite(websocket.TextMessage, []byte("Unexpected text message"))
			continue
		}

		dataTypeBuf := make([]byte, 1)
		read, err := reader.Read(dataTypeBuf)
		if err != nil || read != 1 {
			return
		}
		switch dataTypeBuf[0] {
		case 0:
			_, err := io.Copy(tty, reader)
			if err != nil {
				logger.L.Error().Msg("Error writing to terminal")
			}
		case 1:
			var resizeMessage WindowSize
			if err := json.NewDecoder(reader).Decode(&resizeMessage); err != nil {
				safeWrite(websocket.TextMessage, []byte("Error decoding resize message: "+err.Error()))
				continue
			}

			_, _, errno := syscall.Syscall(
				syscall.SYS_IOCTL,
				tty.Fd(),
				syscall.TIOCSWINSZ,
				uintptr(unsafe.Pointer(&resizeMessage)),
			)
			if errno != 0 {
				safeWrite(websocket.TextMessage, []byte("Error resizing terminal: "+errno.Error()))
			}
		case 2:
			var killMessage struct {
				Kill string `json:"kill"`
			}
			if err := json.NewDecoder(reader).Decode(&killMessage); err != nil {
				logger.L.Error().Msgf("Error decoding kill message: %v", err)
				continue
			}

			sessionID := killMessage.Kill
			if sessionID == "" {
				sessionID = id
			}

			cmd := exec.Command("tmux", "kill-session", "-t", sessionID)
			if err := cmd.Run(); err != nil {
				logger.L.Error().Msgf("Error killing tmux session %s: %v", sessionID, err)
				safeWrite(websocket.TextMessage, []byte("Error killing session: "+err.Error()))
			} else {
				safeWrite(websocket.TextMessage, []byte("Session killed: "+sessionID))
			}

			if sessionID == id {
				return
			}
		default:
			safeWrite(websocket.TextMessage, []byte("Unknown message type"))
		}
	}
}
