package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDomainGlassClient_LookupDomain(t *testing.T) {
	mockResponse := `{
		"target": {"ascii": "example.com", "unicode": "example.com", "kind": "domain"},
		"dns": {
			"status": "NOERROR",
			"records": {
				"A": [{"name": "example.com", "type": "A", "ttl": 300, "value": "93.184.216.34"}]
			}
		},
		"rdap": {
			"registrationStatus": "registered",
			"registrar": {"name": "Test Registrar", "id": "123"}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/domain/example.com" {
			t.Errorf("Beklenmeyen URL path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	c := New(5 * time.Second)
	c.baseURL = server.URL

	ctx := context.Background()
	info, err := c.LookupDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("LookupDomain hatasi: %v", err)
	}

	if info.Target.Ascii != "example.com" {
		t.Errorf("Hedef eslesmedi: %s", info.Target.Ascii)
	}
	if len(info.DNS.Records.A) != 1 || info.DNS.Records.A[0].Value != "93.184.216.34" {
		t.Errorf("A kaydi hatali: %+v", info.DNS.Records.A)
	}
	if info.RDAP.Registrar.Name != "Test Registrar" {
		t.Errorf("Registrar hatali: %s", info.RDAP.Registrar.Name)
	}
}

func TestDomainGlassClient_LookupIP(t *testing.T) {
	mockResponse := `{
		"target": {"address": "1.1.1.1", "version": 4, "isPublic": true},
		"reverseDns": {"names": ["one.one.one.one"], "forwardConfirmed": true},
		"routing": {"announced": true, "prefix": "1.1.1.0/24"}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Domain-Glass-Action") != "ip-page-load" {
			t.Errorf("Eksik veya hatali action header: %s", r.Header.Get("X-Domain-Glass-Action"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	c := New(5 * time.Second)
	c.baseURL = server.URL

	ctx := context.Background()
	info, err := c.LookupIP(ctx, "1.1.1.1")
	if err != nil {
		t.Fatalf("LookupIP hatasi: %v", err)
	}

	if info.Target.Address != "1.1.1.1" {
		t.Errorf("IP eslesmedi: %s", info.Target.Address)
	}
	if !info.Routing.Announced || info.Routing.Prefix != "1.1.1.0/24" {
		t.Errorf("Routing verisi hatali: %+v", info.Routing)
	}
}
