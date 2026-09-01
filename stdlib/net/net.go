// Package net provides HTTP client and networking capabilities for NilLang.
package net

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

var defaultClient = &http.Client{
	Timeout: 30 * time.Second,
}

// HttpResponse represents the result of an HTTP request.
type HttpResponse struct {
	StatusCode int               `json:"statusCode"`
	Status     string            `json:"status"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	OK         bool              `json:"ok"`
}

// Get performs an HTTP GET request and returns the response body.
func Get(url string) (string, error) {
	resp, err := defaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("net.get error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("net.get read error: %w", err)
	}
	return string(bodyBytes), nil
}

// Post performs an HTTP POST request with a payload and contentType.
func Post(url, contentType, body string) (string, error) {
	resp, err := defaultClient.Post(url, contentType, bytes.NewBufferString(body))
	if err != nil {
		return "", fmt.Errorf("net.post error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("net.post read error: %w", err)
	}
	return string(bodyBytes), nil
}

// Fetch executes a full HTTP request with method, headers, and body.
func Fetch(method, url string, headers map[string]string, body string) (*HttpResponse, error) {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("net.fetch request error: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("net.fetch error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("net.fetch read error: %w", err)
	}

	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	return &HttpResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(bodyBytes),
		Headers:    respHeaders,
		OK:         resp.StatusCode >= 200 && resp.StatusCode < 300,
	}, nil
}
