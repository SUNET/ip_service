package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var host = "172.17.0.1:8181"

func init() {
	if h := os.Getenv("IP_SERVICE_HOST"); h != "" {
		host = h
	}
}

type testCase struct {
	name   string
	method string
	path   string
	body   string
	accept string
}

func main() {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}

	testIPs := []string{
		"1.7.83.133",
		"81.233.235.203",
		"2001:67c:2564::1",
		"2a02:d040::1",
	}

	var tests []testCase

	// Index endpoint (plain, json, html)
	tests = append(tests,
		testCase{name: "GET / (plain)", method: "GET", path: "/", accept: "text/plain"},
		testCase{name: "GET / (json)", method: "GET", path: "/", accept: "application/json"},
		testCase{name: "GET / (html)", method: "GET", path: "/", accept: "text/html"},
	)

	// Per-IP endpoints
	for _, ip := range testIPs {
		tests = append(tests,
			testCase{name: fmt.Sprintf("GET /lookup/%s", ip), method: "GET", path: "/lookup/" + ip, accept: "application/json"},
			testCase{name: fmt.Sprintf("GET /whois/%s", ip), method: "GET", path: "/whois/" + ip, accept: "application/json"},
		)
	}

	// Geo endpoints (uses client IP)
	geoEndpoints := []string{"/city", "/asn", "/country", "/country-iso", "/coordinates", "/all"}
	for _, ep := range geoEndpoints {
		tests = append(tests,
			testCase{name: fmt.Sprintf("GET %s (json)", ep), method: "GET", path: ep, accept: "application/json"},
			testCase{name: fmt.Sprintf("GET %s (plain)", ep), method: "GET", path: ep, accept: "text/plain"},
		)
	}

	// Health
	tests = append(tests, testCase{name: "GET /health", method: "GET", path: "/health", accept: "application/json"})

	// Collision endpoint
	tests = append(tests, testCase{
		name:   "POST /collision",
		method: "POST",
		path:   "/collision",
		body:   `{"ip_1": "192.168.1.0/24", "ip_2": "192.168.1.128/25"}`,
		accept: "application/json",
	})

	now := time.Now()
	passed, failed := 0, 0

	for _, tc := range tests {
		var bodyReader io.Reader
		if tc.body != "" {
			bodyReader = strings.NewReader(tc.body)
		}

		url := fmt.Sprintf("http://%s%s", host, tc.path)
		req, err := http.NewRequestWithContext(ctx, tc.method, url, bodyReader)
		if err != nil {
			fmt.Printf("FAIL  %s - create request: %v\n", tc.name, err)
			failed++
			continue
		}
		req.Header.Set("Accept", tc.accept)
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("FAIL  %s - %v\n", tc.name, err)
			failed++
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Try to pretty-print JSON responses
			if strings.Contains(resp.Header.Get("Content-Type"), "json") {
				var j any
				if json.Unmarshal(body, &j) == nil {
					compact, _ := json.Marshal(j)
					if len(compact) > 120 {
						compact = append(compact[:117], '.', '.', '.')
					}
					fmt.Printf("PASS  %s [%d] %s\n", tc.name, resp.StatusCode, string(compact))
				} else {
					fmt.Printf("PASS  %s [%d] (non-json body)\n", tc.name, resp.StatusCode)
				}
			} else {
				text := strings.TrimSpace(string(body))
				if len(text) > 80 {
					text = text[:77] + "..."
				}
				fmt.Printf("PASS  %s [%d] %s\n", tc.name, resp.StatusCode, text)
			}
			passed++
		} else {
			fmt.Printf("FAIL  %s [%d] %s\n", tc.name, resp.StatusCode, strings.TrimSpace(string(body)))
			failed++
		}
	}

	fmt.Printf("\n--- Results: %d passed, %d failed, %d total in %s ---\n", passed, failed, len(tests), time.Since(now).Round(time.Millisecond))
	if failed > 0 {
		os.Exit(1)
	}
}
