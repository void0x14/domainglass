package dns

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// Saglayici, bir DNS çözümleyici uç noktasını tanımlar.
type Saglayici struct {
	ID         string
	Ad         string
	Adres      string
	Aciklama   string
	JSON       bool
	URL        string
	FiltreRolu string
}

// Saglayicilar, kullanılabilir çözümleyicilerin listesidir.
var Saglayicilar = []Saglayici{
	{ID: "sistem", Ad: "Yerel sistem çözümleyici", Aciklama: "İşletim sisteminin yapılandırılmış çözümleyicisi."},
	{ID: "cloudflare", Ad: "Cloudflare 1.1.1.1", Adres: "1.1.1.1", Aciklama: "Genel amaçlı, günlük tutmayan çözümleyici.", JSON: true, URL: "https://cloudflare-dns.com/dns-query"},
	{ID: "cloudflare-guvenlik", Ad: "Cloudflare 1.1.1.1 for Families (Güvenlik)", Adres: "1.1.1.2", Aciklama: "Zararlı yazılım ve oltalama alan adlarını engeller.", JSON: true, URL: "https://security.cloudflare-dns.com/dns-query", FiltreRolu: "malware"},
	{ID: "cloudflare-aile", Ad: "Cloudflare 1.1.1.1 for Families (Aile)", Adres: "1.1.1.3", Aciklama: "Zararlı yazılım, oltalama ve yetişkin içeriği engeller.", JSON: true, URL: "https://family.cloudflare-dns.com/dns-query", FiltreRolu: "adult"},
	{ID: "google", Ad: "Google Public DNS", Adres: "8.8.8.8", Aciklama: "Genel amaçlı genel çözümleyici.", JSON: true, URL: "https://dns.google/resolve"},
	{ID: "quad9", Ad: "Quad9", Adres: "9.9.9.9", Aciklama: "Zararlı yazılım, oltalama ve C2 alan adlarını engeller.", URL: "https://dns.quad9.net/dns-query", FiltreRolu: "malware"},
	{ID: "adguard", Ad: "AdGuard DNS", Adres: "94.140.14.14", Aciklama: "Reklam, izleyici ve zararlı içerik engeller.", URL: "https://dns.adguard-dns.com/dns-query", FiltreRolu: "advertising"},
	{ID: "cleanbrowsing", Ad: "CleanBrowsing Güvenlik Filtresi", Adres: "185.228.168.9", Aciklama: "Zararlı yazılım, oltalama ve yetişkin içerik engeller.", URL: "https://doh.cleanbrowsing.org/doh/security-filter/", FiltreRolu: "adult"},
	{ID: "controld", Ad: "Control D (Free)", Adres: "76.76.2.0", Aciklama: "Reklam, izleyici ve zararlı içerik engeller.", URL: "https://freedns.controld.com/p0", FiltreRolu: "advertising"},
}

// SaglayiciGetir, kimliğe göre sağlayıcı bulur.
func SaglayiciGetir(id string) (Saglayici, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, s := range Saglayicilar {
		if s.ID == id {
			return s, true
		}
	}
	return Saglayici{}, false
}

// Cozumleyici, DoH ve sistem çözümleyicisi üzerinden DNS sorgusu yapar.
type Cozumleyici struct {
	http  *http.Client
	zaman time.Duration
	rand  *rand.Rand
}

// YeniCozumleyici, verilen zaman aşımıyla bir çözümleyici üretir.
func YeniCozumleyici(zamanAsimi time.Duration) *Cozumleyici {
	if zamanAsimi <= 0 {
		zamanAsimi = 10 * time.Second
	}
	return &Cozumleyici{
		http:  &http.Client{Timeout: zamanAsimi},
		zaman: zamanAsimi,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Cozumle, sağlayıcı üzerinden tek bir kayıt tipini sorgular.
func (c *Cozumleyici) Cozumle(ctx context.Context, saglayiciID, ad string, t Type) (*Response, error) {
	s, ok := SaglayiciGetir(saglayiciID)
	if !ok {
		return nil, fmt.Errorf("bilinmeyen çözümleyici: %s", saglayiciID)
	}
	if s.ID == "sistem" {
		return c.sistemCozumle(ctx, ad, t)
	}
	if s.JSON {
		return c.jsonCozumle(ctx, s, ad, t)
	}
	return c.ikiliCozumle(ctx, s, ad, t)
}

// CozumleTum, tüm standart kayıt tiplerini tek sağlayıcı üzerinden sorgular.
func (c *Cozumleyici) CozumleTum(ctx context.Context, saglayiciID, ad string) (map[string][]Record, error) {
	sonuc := make(map[string][]Record, len(AllTypes))
	var sonHata error
	for _, t := range AllTypes {
		resp, err := c.Cozumle(ctx, saglayiciID, ad, t)
		if err != nil {
			sonHata = err
			continue
		}
		sonuc[TypeName(t)] = resp.Records
	}
	if len(sonuc) == 0 && sonHata != nil {
		return nil, sonHata
	}
	return sonuc, nil
}

// Engelli, sağlayıcının bu adı engelleyip engellemediğini döner.
// DNS filtreleri engellenen adları genellikle 0.0.0.0 veya :: adresine yönlendirir.
func Engelli(resp *Response) bool {
	for _, r := range resp.Records {
		if r.Type != "A" && r.Type != "AAAA" {
			continue
		}
		v := strings.TrimSpace(r.Value)
		if v == "0.0.0.0" || v == "::" || v == "127.0.0.1" || v == "::1" {
			return true
		}
	}
	return false
}

func (c *Cozumleyici) sistemCozumle(ctx context.Context, ad string, t Type) (*Response, error) {
	if t == TypeA || t == TypeAAAA {
		ips, err := net.DefaultResolver.LookupIP(ctx, ipAg(t), ad)
		if err != nil {
			return &Response{Rcode: rcodeFromErr(err), Records: []Record{}}, nil
		}
		resp := &Response{Rcode: "NOERROR", Records: []Record{}}
		for _, ip := range ips {
			resp.Records = append(resp.Records, Record{Name: ad, Type: TypeName(t), Value: ip.String()})
		}
		return resp, nil
	}
	if t == TypeMX {
		mxs, err := net.DefaultResolver.LookupMX(ctx, ad)
		if err != nil {
			return &Response{Rcode: rcodeFromErr(err), Records: []Record{}}, nil
		}
		resp := &Response{Rcode: "NOERROR", Records: []Record{}}
		for _, mx := range mxs {
			resp.Records = append(resp.Records, Record{Name: ad, Type: "MX", Value: fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, "."))})
		}
		return resp, nil
	}
	if t == TypeNS {
		nss, err := net.DefaultResolver.LookupNS(ctx, ad)
		if err != nil {
			return &Response{Rcode: rcodeFromErr(err), Records: []Record{}}, nil
		}
		resp := &Response{Rcode: "NOERROR", Records: []Record{}}
		for _, ns := range nss {
			resp.Records = append(resp.Records, Record{Name: ad, Type: "NS", Value: strings.TrimSuffix(ns.Host, ".")})
		}
		return resp, nil
	}
	if t == TypeTXT {
		txts, err := net.DefaultResolver.LookupTXT(ctx, ad)
		if err != nil {
			return &Response{Rcode: rcodeFromErr(err), Records: []Record{}}, nil
		}
		resp := &Response{Rcode: "NOERROR", Records: []Record{}}
		for _, txt := range txts {
			resp.Records = append(resp.Records, Record{Name: ad, Type: "TXT", Value: txt})
		}
		return resp, nil
	}
	if t == TypeCNAME {
		cname, err := net.DefaultResolver.LookupCNAME(ctx, ad)
		if err != nil {
			return &Response{Rcode: rcodeFromErr(err), Records: []Record{}}, nil
		}
		cname = strings.TrimSuffix(cname, ".")
		if cname == ad {
			return &Response{Rcode: "NOERROR", Records: []Record{}}, nil
		}
		return &Response{Rcode: "NOERROR", Records: []Record{{Name: ad, Type: "CNAME", Value: cname}}}, nil
	}
	if t == TypePTR {
		names, err := net.DefaultResolver.LookupAddr(ctx, ad)
		if err != nil {
			return &Response{Rcode: rcodeFromErr(err), Records: []Record{}}, nil
		}
		resp := &Response{Rcode: "NOERROR", Records: []Record{}}
		for _, n := range names {
			resp.Records = append(resp.Records, Record{Name: ad, Type: "PTR", Value: strings.TrimSuffix(n, ".")})
		}
		return resp, nil
	}
	// CAA ve SOA için sistem çözümleyicisi doğrudan API sunmaz; DoH kullanılır.
	return nil, fmt.Errorf("sistem çözümleyici %s tipini desteklemiyor", TypeName(t))
}

func ipAg(t Type) string {
	if t == TypeAAAA {
		return "ip6"
	}
	return "ip4"
}

func rcodeFromErr(err error) string {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsNotFound {
			return "NXDOMAIN"
		}
	}
	return "SERVFAIL"
}

type jsonYanit struct {
	Status int `json:"Status"`
	Answer []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
		TTL  int    `json:"TTL"`
		Data string `json:"data"`
	} `json:"Answer"`
	Comment []string `json:"Comment"`
}

func (c *Cozumleyici) jsonCozumle(ctx context.Context, s Saglayici, ad string, t Type) (*Response, error) {
	u := s.URL
	if strings.Contains(u, "?") {
		u += "&"
	} else {
		u += "?"
	}
	u += "name=" + urlSorgu(ad) + "&type=" + TypeName(t)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	govde, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s HTTP %d döndürdü", s.Ad, resp.StatusCode)
	}
	var jy jsonYanit
	if err := json.Unmarshal(govde, &jy); err != nil {
		return nil, fmt.Errorf("%s yanıtı ayrıştırılamadı: %w", s.Ad, err)
	}
	sonuc := &Response{Rcode: Rcode(jy.Status).String(), Records: []Record{}}
	for _, a := range jy.Answer {
		sonuc.Records = append(sonuc.Records, Record{
			Name:  strings.TrimSuffix(a.Name, "."),
			Type:  TypeName(Type(a.Type)),
			TTL:   uint32(a.TTL),
			Value: a.Data,
		})
	}
	return sonuc, nil
}

func (c *Cozumleyici) ikiliCozumle(ctx context.Context, s Saglayici, ad string, t Type) (*Response, error) {
	id := uint16(c.rand.Intn(65535))
	sorgu, err := EncodeQuery(id, ad, t)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(sorgu))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	govde, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s HTTP %d döndürdü", s.Ad, resp.StatusCode)
	}
	return DecodeResponse(govde)
}

func urlSorgu(s string) string {
	return strings.NewReplacer(" ", "%20", "+", "%2B", "&", "%26", "?", "%3F", "#", "%23").Replace(s)
}

// KodlaJSON, JSON DoH sorgusu için yardımcıdır (test ve hata ayıklama).
func KodlaJSON(ad string, t Type) string {
	b, _ := json.Marshal(map[string]string{"name": ad, "type": TypeName(t)})
	return string(b)
}

// Base64URL, ikili DoH sorgusunu GET parametresi olarak kodlar.
func Base64URL(sorgu []byte) string {
	return base64.RawURLEncoding.EncodeToString(sorgu)
}
