// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetTokenFromHeader(r http.Header) (string, error) {
	token := r.Get("Authorization")
	if token != "" {
		if len(token) < 8 || !strings.HasPrefix(token, "Bearer ") {
			return "", fmt.Errorf("invalid authorization header format")
		}
		return RemoveSpaces(token[7:]), nil
	}

	wsProtocol := r.Get("Sec-WebSocket-Protocol")
	if wsProtocol != "" {
		parts := strings.Split(wsProtocol, ",")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "Bearer" {
			return RemoveSpaces(strings.TrimSpace(parts[1])), nil
		}
		return "", errors.New("invalid websocket protocol header format")
	}

	return "", errors.New("no token provided")
}

func GetClusterTokenFromHeader(r http.Header) (string, error) {
	if v := r.Get("ClusterToken"); v != "" {
		if len(v) < 8 || !strings.HasPrefix(v, "Bearer ") {
			return "", fmt.Errorf("invalid ClusterToken header format")
		}
		return RemoveSpaces(v[7:]), nil
	}

	if v := r.Get("X-Cluster-Authorization"); v != "" {
		if len(v) < 8 || !strings.HasPrefix(v, "Bearer ") {
			return "", fmt.Errorf("invalid X-Cluster-Authorization header format")
		}
		return RemoveSpaces(v[7:]), nil
	}

	if v := r.Get("X-Cluster-Token"); v != "" {
		if len(v) < 8 || !strings.HasPrefix(v, "Bearer ") {
			return "", fmt.Errorf("invalid X-Cluster-Token header format")
		}
		return RemoveSpaces(v[7:]), nil
	}

	return "", errors.New("no cluster token provided")
}

func GetIdFromParam(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func FlatHeaders(c *gin.Context) map[string]string {
	var flatHeaders = make(map[string]string)
	for key, value := range c.Request.Header {
		if len(value) > 0 {
			flatHeaders[key] = value[0]
		}
	}
	return flatHeaders
}

func HTTPPostJSON(url string, payload any, headers map[string]string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	default:
		reader = resp.Body
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(reader)
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func HTTPPostJSONRead(url string, payload any, headers map[string]string) ([]byte, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   6 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	} else {
		reader = resp.Body
	}

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("http error %d: %s", resp.StatusCode, string(b))
	}

	return b, resp.StatusCode, nil
}
