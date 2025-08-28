package clusterHandlers

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/alchemillahq/sylve/internal"
	"github.com/alchemillahq/sylve/internal/config"
	"github.com/alchemillahq/sylve/internal/services/cluster"
	"github.com/gin-gonic/gin"
)

func mapRaftAddrToAPI(raftAddr string) (string, error) {
	host, _, err := net.SplitHostPort(raftAddr)
	if err != nil {
		return "", err
	}

	scheme := "https"
	apiPort := config.ParsedConfig.Port

	return (&url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(apiPort)),
	}).String(), nil
}

func ReverseProxy(c *gin.Context, backend string) {
	remote, err := url.Parse(backend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse proxy URL"})
		return
	}

	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		if !strings.Contains(err.Error(), "context canceled") {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
	}

	q := c.Request.URL.Query()

	clusterKey := c.GetString("ClusterKey")
	if clusterKey != "" {
		q.Set("clusterkey", clusterKey)
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	proxy.ServeHTTP(c.Writer, c.Request)
}

func forwardToLeader(c *gin.Context, cS *cluster.Service) {
	leaderAddr, _ := cS.Raft.LeaderWithID()
	if leaderAddr == "" {
		_ = cS.Raft.VerifyLeader().Error()
		c.JSON(http.StatusServiceUnavailable, internal.APIResponse[any]{
			Status: "error", Message: "leader_unknown",
		})
		return
	}

	base, err := mapRaftAddrToAPI(string(leaderAddr))
	if err != nil {
		c.JSON(http.StatusBadGateway, internal.APIResponse[any]{
			Status: "error", Message: "map_leader_api_failed", Error: err.Error(),
		})
		return
	}

	ReverseProxy(c, base)
}
