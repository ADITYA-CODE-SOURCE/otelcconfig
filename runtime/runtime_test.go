// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"testing"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/generated/types"
)

func TestRegisterAndAccessor(t *testing.T) {
	Register(ConfigSnapshot{
		NetHTTPClient: types.NetHTTPClientConfig{
			Enabled:                  false,
			RequestCapturedHeaders:   []string{"user-agent"},
			SensitiveQueryParameters: []string{"token"},
		},
	})

	got := NetHTTPClient()
	if got.Enabled {
		t.Fatal("Enabled = true, want false")
	}
	if len(got.RequestCapturedHeaders) != 1 || got.RequestCapturedHeaders[0] != "user-agent" {
		t.Fatalf("RequestCapturedHeaders = %v, want [user-agent]", got.RequestCapturedHeaders)
	}
	if len(got.SensitiveQueryParameters) != 1 || got.SensitiveQueryParameters[0] != "token" {
		t.Fatalf("SensitiveQueryParameters = %v, want [token]", got.SensitiveQueryParameters)
	}
}

func TestAccessorReturnsCopy(t *testing.T) {
	Register(ConfigSnapshot{
		NetHTTPClient: types.NetHTTPClientConfig{
			RequestCapturedHeaders:   []string{"user-agent"},
			SensitiveQueryParameters: []string{"token"},
		},
	})

	a := NetHTTPClient()
	a.RequestCapturedHeaders[0] = "mutated"
	a.SensitiveQueryParameters = append(a.SensitiveQueryParameters, "added")

	b := NetHTTPClient()
	if b.RequestCapturedHeaders[0] != "user-agent" {
		t.Fatalf("RequestCapturedHeaders mutated by caller: %v", b.RequestCapturedHeaders)
	}
	if len(b.SensitiveQueryParameters) != 1 {
		t.Fatalf("SensitiveQueryParameters mutated by caller: %v", b.SensitiveQueryParameters)
	}
}

func TestCanReplaceSnapshot(t *testing.T) {
	Register(ConfigSnapshot{NetHTTPClient: types.NetHTTPClientConfig{Enabled: false}})
	Register(ConfigSnapshot{NetHTTPClient: types.NetHTTPClientConfig{Enabled: true}})
	if got := NetHTTPClient(); !got.Enabled {
		t.Fatal("Enabled = false, want true after second Register")
	}
}

func TestAccessorPanicsWithoutSnapshot(t *testing.T) {
	orig := lookupSnapshot
	lookupSnapshot = func() (ConfigSnapshot, bool) { return ConfigSnapshot{}, false }
	defer func() { lookupSnapshot = orig }()

	defer func() {
		if recover() == nil {
			t.Fatal("NetHTTPClient did not panic without a baked snapshot")
		}
	}()
	_ = NetHTTPClient()
}

func TestCopyConfig(t *testing.T) {
	src := types.NetHTTPClientConfig{
		RequestCapturedHeaders:   []string{"a", "b"},
		SensitiveQueryParameters: []string{"c"},
	}
	dst := copyConfig(src)
	dst.RequestCapturedHeaders[0] = "x"
	if src.RequestCapturedHeaders[0] != "a" {
		t.Fatalf("copyConfig shared backing array: %v", src.RequestCapturedHeaders)
	}
}
