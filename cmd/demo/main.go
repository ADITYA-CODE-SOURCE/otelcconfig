// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Command demo demonstrates the otelcconfig supply chain end to end: declarative
// YAML (examples/demo.yaml) is validated and resolved, frozen into the baked
// package by `otelcconfig bake`, and consumed here through the runtime package.
// The hook never parses YAML and the binary's behavior is fixed at build time.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	_ "github.com/ADITYA-CODE-SOURCE/otelcconfig/baked"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/demo"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	rt := demo.Client(nil, func(o demo.Observation) {
		fmt.Printf("method=%s\n", o.Method)
		fmt.Printf("url=%s\n", o.URL)
		for _, name := range sortedKeys(o.CapturedHeaders) {
			fmt.Printf("header %s=%s\n", name, strings.Join(o.CapturedHeaders[name], ", "))
		}
	})
	client := &http.Client{Transport: rt}

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/demo?token=supersecret&keep=visible", nil)
	if err != nil {
		fmt.Printf("error=%v\n", err)
		return
	}
	req.Header.Set("User-Agent", "otelcconfig-demo/1.0")
	req.Header.Set("X-Request-Id", "demo-request-42")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error=%v\n", err)
		return
	}
	if err := resp.Body.Close(); err != nil {
		fmt.Printf("close=%v\n", err)
	}

	enabler := demo.Enabler{}
	if enabler.Enable() {
		fmt.Println("enabled=true")
		return
	}
	fmt.Println("enabled=false")
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
