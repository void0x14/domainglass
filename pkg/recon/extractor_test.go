package recon

import (
	"testing"

	"domainglass/pkg/models"
)

func TestExtractDiscoveredHosts(t *testing.T) {
	domainInfo := &models.DomainInfo{}
	domainInfo.DNS.Records.A = []models.DNSRecord{
		{Value: "93.184.216.34"},
	}
	domainInfo.DNS.Records.CNAME = []models.DNSRecord{
		{Value: "cdn.example.com"},
	}
	domainInfo.DNS.Records.NS = []models.DNSRecord{
		{Value: "ns1.iana.org."},
	}

	tlsInfo := &models.TLSInfo{}
	tlsInfo.Certificate.DNSNames = []struct {
		Value        string `json:"value"`
		Unicode      string `json:"unicode"`
		Wildcard     bool   `json:"wildcard"`
		Relationship string `json:"relationship"`
	}{
		{Value: "example.com"},
		{Value: "*.example.com"},
		{Value: "sub.example.com"},
		{Value: "otherdomain.com"},
	}

	webProfile := &models.WebProfileInfo{
		DiscoveredURLs: []string{
			"https://api.example.com/v1",
			"https://partner.com/login",
		},
	}

	discovered := ExtractDiscoveredHosts("example.com", domainInfo, tlsInfo, webProfile)

	// Check subdomains
	foundCDN := false
	foundSub := false
	foundAPI := false
	for _, sub := range discovered.Subdomains {
		if sub == "cdn.example.com" {
			foundCDN = true
		}
		if sub == "sub.example.com" {
			foundSub = true
		}
		if sub == "api.example.com" {
			foundAPI = true
		}
	}
	if !foundCDN || !foundSub || !foundAPI {
		t.Fatalf("Subdomainler eksik: %+v", discovered.Subdomains)
	}

	// Check related domains
	foundIANA := false
	foundOther := false
	for _, rel := range discovered.Related {
		if rel == "ns1.iana.org" {
			foundIANA = true
		}
		if rel == "otherdomain.com" {
			foundOther = true
		}
	}
	if !foundIANA || !foundOther {
		t.Fatalf("İlişkili alan adları eksik: %+v", discovered.Related)
	}

	// Check IPs
	foundIP := false
	for _, ip := range discovered.IPs {
		if ip == "93.184.216.34" {
			foundIP = true
		}
	}
	if !foundIP {
		t.Fatalf("IP adresi eksik: %+v", discovered.IPs)
	}
}

func TestSanitizeHost(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"*.test.com", "test.com"},
		{"https://app.test.com/path", "app.test.com"},
		{"ns1.test.com.", "ns1.test.com"},
		{"  Sub.Test.COM:8080  ", "sub.test.com"},
	}

	for _, c := range cases {
		got := SanitizeHost(c.input)
		if got != c.expected {
			t.Errorf("SanitizeHost(%q) = %q; beklenen %q", c.input, got, c.expected)
		}
	}
}
