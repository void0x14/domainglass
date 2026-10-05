package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"domainglass/pkg/models"
)

const (
	DefaultBaseURL   = "https://domain.glass"
	DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
)

type DomainGlassClient struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
}

func New(timeout time.Duration) *DomainGlassClient {
	return &DomainGlassClient{
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *DomainGlassClient) get(ctx context.Context, endpoint string, action string, referer string) ([]byte, error) {
	reqURL := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")

	if action != "" {
		req.Header.Set("X-Domain-Glass-Action", action)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
		req.Header.Set("Origin", c.baseURL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// LookupDomain, domain.glass'ın temel DNS & RDAP verilerini çeker.
func (c *DomainGlassClient) LookupDomain(ctx context.Context, target string) (*models.DomainInfo, error) {
	body, err := c.get(ctx, "/api/v1/domain/"+url.PathEscape(target), "", "")
	if err != nil {
		return nil, err
	}
	var res models.DomainInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// GetRanking, Cisco Umbrella veya Tranco popülerlik sıralamasını döndürür.
func (c *DomainGlassClient) GetRanking(ctx context.Context, target string, source string) (*models.RankingInfo, error) {
	endpoint := fmt.Sprintf("/api/v1/rankings/domain/%s?source=%s", url.PathEscape(target), url.QueryEscape(source))
	body, err := c.get(ctx, endpoint, "", "")
	if err != nil {
		return nil, err
	}
	var res models.RankingInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// GetSafety, Quad9, AdGuard, CleanBrowsing, ControlD DNS filtreleme verisini çeker.
func (c *DomainGlassClient) GetSafety(ctx context.Context, target string) (*models.SafetyInfo, error) {
	referer := c.baseURL + "/" + target
	body, err := c.get(ctx, "/api/v1/domain/"+url.PathEscape(target)+"/safety", "domain-safety-load", referer)
	if err != nil {
		return nil, err
	}
	var res models.SafetyInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// GetTLS, domain.glass üzerinden port 443 sertifika detaylarını sorgular.
func (c *DomainGlassClient) GetTLS(ctx context.Context, target string, kind string) (*models.TLSInfo, error) {
	referer := c.baseURL + "/" + target
	endpoint := fmt.Sprintf("/api/v1/tls/%s/%s", kind, url.PathEscape(target))
	body, err := c.get(ctx, endpoint, "tls-page-load", referer)
	if err != nil {
		return nil, err
	}
	var res models.TLSInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// GetWebProfile, ana sayfa HTTP başlıkları, meta etiketleri ve link keşiflerini çeker.
func (c *DomainGlassClient) GetWebProfile(ctx context.Context, target string) (*models.WebProfileInfo, error) {
	referer := c.baseURL + "/" + target
	body, err := c.get(ctx, "/api/v1/web-profile/"+url.PathEscape(target), "domain-web-profile-load", referer)
	if err != nil {
		return nil, err
	}
	var res models.WebProfileInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// Search, domain.glass ilişkili arama sonuçlarını döndürür.
func (c *DomainGlassClient) Search(ctx context.Context, query string) (*models.SearchInfo, error) {
	endpoint := "/api/v1/search?q=" + url.QueryEscape(query)
	body, err := c.get(ctx, endpoint, "search-click", c.baseURL)
	if err != nil {
		return nil, err
	}
	var res models.SearchInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// LookupIP, IP adresi için BGP, ASN, lokasyon ve RIR RDAP bilgilerini çeker.
func (c *DomainGlassClient) LookupIP(ctx context.Context, ip string) (*models.IPInfo, error) {
	referer := c.baseURL + "/" + ip
	body, err := c.get(ctx, "/api/v1/ip/"+url.PathEscape(ip), "ip-page-load", referer)
	if err != nil {
		return nil, err
	}
	var res models.IPInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
	}
	return &res, nil
}

// ProbeDirectTLS, domain.glass TLS API'si yanıt vermezse canlı port 443 bağlantısı ile yedek çalışır.
func ProbeDirectTLS(target string, timeout time.Duration) (*models.DirectTLSResult, error) {
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(target, "443"), &tls.Config{
		ServerName:         target,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("sertifika bulunamadi")
	}

	cert := certs[0]
	issuer := ""
	if len(cert.Issuer.Organization) > 0 {
		issuer = strings.Join(cert.Issuer.Organization, ", ")
	}

	days := int(time.Until(cert.NotAfter).Hours() / 24)

	return &models.DirectTLSResult{
		SubjectCN:     cert.Subject.CommonName,
		IssuerOrg:     issuer,
		SANs:          cert.DNSNames,
		ValidFrom:     cert.NotBefore.Format(time.RFC3339),
		ValidTo:       cert.NotAfter.Format(time.RFC3339),
		DaysRemaining: days,
	}, nil
}
