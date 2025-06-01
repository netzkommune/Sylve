package vncHandler

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sylve/internal/logger"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func VNCProxyHandler(c *gin.Context) {
	port := c.Param("port")
	if port == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'port' query parameter"})
		return
	}

	vncAddress := fmt.Sprintf("localhost:%s", port)
	vncConn, err := net.Dial("tcp", vncAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to connect to VNC server on port %s", port)})
		return
	}
	defer vncConn.Close()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()

	done := make(chan bool)
	var once sync.Once

	closeDone := func() {
		once.Do(func() {
			close(done)
		})
	}

	go func() {
		defer closeDone()
		buffer := make([]byte, 1024)
		for {
			n, err := vncConn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					logger.L.Debug().Err(err).Msg("Error reading from VNC connection")
				}
				return
			}
			err = conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
			if err != nil {
				return
			}
		}
	}()

	go func() {
		defer closeDone()
		for {
			_, message, err := conn.ReadMessage()
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			_, err = vncConn.Write(message)
			if err != nil {
				return
			}
		}
	}()

	<-done
}
