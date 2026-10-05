package recon

import (
	"context"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"domainglass/pkg/client"
	"domainglass/pkg/models"
)

// IsIP verilen değerin geçerli bir IP olup olmadığını doğrular.
func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

// SanitizeHost alan adı, URL veya port içeren host stringlerini temizler ve normalize eder.
func SanitizeHost(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "*.")
	s = strings.TrimSuffix(s, ".")
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		s = u.Host
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	return s
}

// ExtractDiscoveredHosts, DNS kayıtları, TLS SAN'ları ve Web profilinden alt alan adlarını,
// ilişkili hostları ve IP'leri ayıklar.
func ExtractDiscoveredHosts(target string, d *models.DomainInfo, t *models.TLSInfo, w *models.WebProfileInfo) models.DiscoveredHosts {
	subdomains := make(map[string]bool)
	related := make(map[string]bool)
	ips := make(map[string]bool)

	rootTarget := target

	processHost := func(h string) {
		h = SanitizeHost(h)
		if h == "" || h == target {
			return
		}
		if IsIP(h) {
			ips[h] = true
			return
		}
		if strings.HasSuffix(h, "."+rootTarget) {
			subdomains[h] = true
		} else {
			related[h] = true
		}
	}

	if d != nil {
		for _, rec := range d.DNS.Records.A {
			if rec.Value != "" {
				ips[rec.Value] = true
			}
		}
		for _, rec := range d.DNS.Records.AAAA {
			if rec.Value != "" {
				ips[rec.Value] = true
			}
		}
		for _, rec := range d.DNS.Records.CNAME {
			processHost(rec.Value)
		}
		for _, rec := range d.DNS.Records.MX {
			parts := strings.Fields(rec.Value)
			if len(parts) >= 2 {
				processHost(parts[1])
			} else {
				processHost(rec.Value)
			}
		}
		for _, rec := range d.DNS.Records.NS {
			processHost(rec.Value)
		}
		for _, ns := range d.RDAP.Nameservers {
			processHost(ns)
		}
	}

	if t != nil && t.Certificate.DNSNames != nil {
		for _, san := range t.Certificate.DNSNames {
			processHost(san.Value)
		}
		for _, ip := range t.Certificate.IPAddresses {
			ips[ip] = true
		}
	}

	if w != nil {
		for _, u := range w.DiscoveredURLs {
			if parsed, err := url.Parse(u); err == nil {
				processHost(parsed.Host)
			}
		}
	}

	res := models.DiscoveredHosts{
		Subdomains: make([]string, 0, len(subdomains)),
		Related:    make([]string, 0, len(related)),
		IPs:        make([]string, 0, len(ips)),
	}
	for s := range subdomains {
		res.Subdomains = append(res.Subdomains, s)
	}
	for r := range related {
		res.Related = append(res.Related, r)
	}
	for i := range ips {
		res.IPs = append(res.IPs, i)
	}
	sort.Strings(res.Subdomains)
	sort.Strings(res.Related)
	sort.Strings(res.IPs)
	return res
}

// RunRecon, verilen hedef domain veya IP için tüm domain.glass ve yedek kontrolleri paralel yürütür.
func RunRecon(ctx context.Context, c *client.DomainGlassClient, target string) (*models.TargetReport, error) {
	report := &models.TargetReport{Target: target}

	if IsIP(target) {
		report.Kind = "ip"
		ipInfo, err := c.LookupIP(ctx, target)
		if err == nil {
			report.IP = ipInfo
		}
		tlsInfo, _ := c.GetTLS(ctx, target, "ip")
		if tlsInfo != nil && tlsInfo.Status == "available" {
			report.TLS = tlsInfo
		} else {
			dirTLS, _ := client.ProbeDirectTLS(target, 5*time.Second)
			report.DirectTLS = dirTLS
		}
		return report, nil
	}

	report.Kind = "domain"
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		d, err := c.LookupDomain(ctx, target)
		if err == nil {
			report.Domain = d
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := c.GetRanking(ctx, target, "cisco")
		if err == nil {
			report.CiscoRank = r
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := c.GetRanking(ctx, target, "tranco")
		if err == nil {
			report.TrancoRank = r
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := c.GetSafety(ctx, target)
		if err == nil {
			report.Safety = s
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		t, err := c.GetTLS(ctx, target, "domain")
		if err == nil && t.Status == "available" {
			report.TLS = t
		} else {
			dirTLS, _ := client.ProbeDirectTLS(target, 5*time.Second)
			report.DirectTLS = dirTLS
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		w, err := c.GetWebProfile(ctx, target)
		if err == nil {
			report.WebProfile = w
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := c.Search(ctx, target)
		if err == nil {
			report.Search = s
		}
	}()

	wg.Wait()

	report.DiscoveredHosts = ExtractDiscoveredHosts(target, report.Domain, report.TLS, report.WebProfile)
	return report, nil
}
