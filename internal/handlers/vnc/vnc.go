// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package vncHandler

import (
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alchemillahq/sylve/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    8 * 1024,
	WriteBufferSize:   8 * 1024,
	EnableCompression: false,
	CheckOrigin:       func(r *http.Request) bool { return true },
}

func VNCProxyHandler(c *gin.Context) {
	port := c.Param("port")
	if port == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'port' parameter"})
		return
	}

	rawConn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to VNC"})
		return
	}
	defer rawConn.Close()

	if tcp, ok := rawConn.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
		_ = tcp.SetKeepAlive(true)
		_ = tcp.SetKeepAlivePeriod(30 * time.Second)
	}

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer wsConn.Close()

	// WS keepalive
	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wsConn.SetPongHandler(func(string) error {
		wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	done := make(chan struct{})
	var once sync.Once
	closeDone := func() { once.Do(func() { close(done) }) }

	buf := make([]byte, 32*1024)

	// WS → VNC
	go func() {
		defer closeDone()
		for {
			msgType, reader, err := wsConn.NextReader()
			if err != nil {
				return
			}
			if msgType != websocket.BinaryMessage {
				_, _ = io.Copy(io.Discard, reader)
				continue
			}
			if _, err := io.Copy(rawConn, reader); err != nil {
				return
			}
		}
	}()

	// VNC → WS
	go func() {
		defer closeDone()
		for {
			n, err := rawConn.Read(buf)
			if err != nil {
				if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
					logger.L.Debug().Err(err).Msg("VNC read error")
				}
				return
			}
			writer, err := wsConn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			if _, err := writer.Write(buf[:n]); err != nil {
				_ = writer.Close()
				return
			}
			if err := writer.Close(); err != nil {
				return
			}
		}
	}()

	<-done
}
