// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package vncHandler

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/alchemillahq/sylve/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func VNCProxyHandler(c *gin.Context) {
	port := c.Param("port")
	if port == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'port' parameter"})
		return
	}

	vncAddress := fmt.Sprintf("localhost:%s", port)
	rawConn, err := net.Dial("tcp", vncAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to connect to VNC on port %s", port)})
		return
	}
	defer rawConn.Close()

	// Disable Nagle's algorithm for lower latency
	if tcpConn, ok := rawConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}
	defer wsConn.Close()

	done := make(chan struct{})
	var once sync.Once
	closeDone := func() { once.Do(func() { close(done) }) }

	const bufSize = 32 * 1024
	buffer := make([]byte, bufSize)

	go func() {
		defer closeDone()
		for {
			msgType, reader, err := wsConn.NextReader()
			if err != nil {
				return
			}

			if msgType != websocket.BinaryMessage {
				io.Copy(io.Discard, reader)
				continue
			}

			if _, err := io.Copy(rawConn, reader); err != nil {
				return
			}
		}
	}()

	go func() {
		defer closeDone()
		for {
			n, err := rawConn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					if !strings.Contains(err.Error(), "use of closed network connection") {
						logger.L.Debug().Err(err).Msg("Error reading from VNC connection")
					}
				}
				return
			}
			writer, err := wsConn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			if _, err := writer.Write(buffer[:n]); err != nil {
				writer.Close()
				return
			}
			if err := writer.Close(); err != nil {
				return
			}
		}
	}()

	<-done
}
