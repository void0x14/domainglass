// Package dg, domain.glass HTTP API istemcisidir.
//
// Sorumluluklar: istek başlıklarının site sözleşmesine uydurulması, hız sınırlama,
// 429/403 geçici kapılarında geri çekilmeli yeniden deneme ve tipli hata üretimi.
package dg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VarsayilanTabanURL, domain.glass kök adresidir.
const VarsayilanTabanURL = "https://domain.glass"

// VarsayilanTarayiciUA, site tarafındaki tarayıcı denetimini geçen gerçekçi bir UA dizesidir.
const VarsayilanTarayiciUA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

// HataTuru, istemci hatalarının sınıfıdır. Çağıran taraf kararını buna göre verir.
type HataTuru string

// Hata sınıfları.
const (
	HataGecici   HataTuru = "gecici"   // hız sınırı / geçici kapı
	HataYok      HataTuru = "yok"      // kayıt bulunamadı
	HataGecersiz HataTuru = "gecersiz" // doğrulama hatası
	HataErisim   HataTuru = "erisim"   // kalıcı erişim engeli
	HataAg       HataTuru = "ag"       // ağ hatası
	HataBicim    HataTuru = "bicim"    // çözümleme hatası
)

// Hata, domain.glass istemcisinin tipli hatasıdır.
type Hata struct {
	Tur           HataTuru
	HTTPKod       int
	Kod           string
	Mesaj         string
	Kaynak        string
	YenidenDeneme int
}

// Error, hata mesajını Türkçe biçimlendirir.
func (e *Hata) Error() string {
	if e.Kod != "" {
		return fmt.Sprintf("domain.glass %s hatası (HTTP %d, kod=%s): %s", e.Tur, e.HTTPKod, e.Kod, e.Mesaj)
	}
	return fmt.Sprintf("domain.glass %s hatası (HTTP %d): %s", e.Tur, e.HTTPKod, e.Mesaj)
}

// GeciciMi, hatanın yeniden denenebilir olduğunu bildirir.
func GeciciMi(err error) bool {
	var h *Hata
	return errors.As(err, &h) && h.Tur == HataGecici
}

// YokMu, hatanın "kayıt yok" sınıfında olduğunu bildirir.
func YokMu(err error) bool {
	var h *Hata
	return errors.As(err, &h) && h.Tur == HataYok
}

// IcerikHatasi, API gövdesindeki hata zarfıdır.
type IcerikHatasi struct {
	Error struct {
		Source  string `json:"source"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Secenek, istemci yapılandırmasıdır.
type Secenek func(*Client)

// ZamanAsimi, HTTP zaman aşımını ayarlar.
func ZamanAsimi(d time.Duration) Secenek {
	return func(c *Client) { c.zamanAsimi = d }
}

// AsgariAralik, iki istek arasındaki asgari süreyi ayarlar (hız sınırı).
func AsgariAralik(d time.Duration) Secenek {
	return func(c *Client) { c.asgariAralik = d }
}

// YenidenDeneme, geçici hatalarda deneme sayısını ayarlar.
func YenidenDeneme(n int) Secenek {
	return func(c *Client) { c.denemeSayisi = n }
}

// TabanURL, API kök adresini değiştirir (test için).
func TabanURL(u string) Secenek {
	return func(c *Client) { c.tabanURL = strings.TrimSuffix(u, "/") }
}

// KullaniciAraci, HTTP User-Agent değerini değiştirir.
func KullaniciAraci(ua string) Secenek {
	return func(c *Client) { c.ua = ua }
}

// VarsayilanAsgariAralik, domain.glass hız sınırına takılmamak için ölçülen
// güvenli aralıktır (canlı testte 3 istek/saniye sınırına takılmadan geçti,
// 1,1 sn aralık güvenli pay bırakır).
const VarsayilanAsgariAralik = 1100 * time.Millisecond

// Client, domain.glass API istemcisidir.
type Client struct {
	tabanURL     string
	ua           string
	http         *http.Client
	zamanAsimi   time.Duration
	asgariAralik time.Duration
	denemeSayisi int

	mu        sync.Mutex
	sonIstek  time.Time
	rand      *rand.Rand
	Kullanici *Kullanici
}

// Kullanici, eşzamanlılık ve oturum bilgisidir.
type Kullanici struct {
	IstekSayisi int64
	Basarisiz   int64
}

// Yeni, varsayılan ayarlarla bir istemci üretir.
func Yeni(secenekler ...Secenek) *Client {
	c := &Client{
		tabanURL:     VarsayilanTabanURL,
		ua:           VarsayilanTarayiciUA,
		zamanAsimi:   20 * time.Second,
		asgariAralik: VarsayilanAsgariAralik,
		denemeSayisi: 3,
		rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
		Kullanici:    &Kullanici{},
	}
	for _, s := range secenekler {
		s(c)
	}
	c.http = &http.Client{
		Timeout: c.zamanAsimi,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          32,
			MaxIdleConnsPerHost:   8,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
	return c
}

// TabanURL, istemcinin kök adresini döner.
func (c *Client) TabanURL() string { return c.tabanURL }

// AsgariAralikDegeri, hız sınırı aralığını döner.
func (c *Client) AsgariAralikDegeri() time.Duration { return c.asgariAralik }

// bekle, iki istek arasında asgari aralığı uygular ve istekleri sıraya sokar.
func (c *Client) bekle(ctx context.Context) error {
	c.mu.Lock()
	sonraki := c.sonIstek.Add(c.asgariAralik)
	simdi := time.Now()
	var uyku time.Duration
	if simdi.Before(sonraki) {
		uyku = sonraki.Sub(simdi)
	}
	c.sonIstek = simdi.Add(uyku)
	c.mu.Unlock()
	if uyku <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(uyku)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// istek, düşük seviyeli HTTP çağrısıdır; hız sınırı ve yeniden deneme uygular.
func (c *Client) istek(ctx context.Context, method, yol, action, referer string, gonder []byte) ([]byte, error) {
	var sonHata error
	denemeler := c.denemeSayisi
	if denemeler < 1 {
		denemeler = 1
	}
	for deneme := 0; deneme < denemeler; deneme++ {
		if err := c.bekle(ctx); err != nil {
			return nil, err
		}
		govde, err := c.tekIstek(ctx, method, yol, action, referer, gonder)
		if err == nil {
			c.mu.Lock()
			c.Kullanici.IstekSayisi++
			c.mu.Unlock()
			return govde, nil
		}
		var h *Hata
		if errors.As(err, &h) {
			h.YenidenDeneme = deneme
		}
		if !GeciciMi(err) {
			c.mu.Lock()
			c.Kullanici.Basarisiz++
			c.mu.Unlock()
			return nil, err
		}
		sonHata = err
		if deneme == denemeler-1 {
			break
		}
		// Üstel geri çekilme + jitter.
		bekleme := time.Duration(1<<uint(deneme)) * 1500 * time.Millisecond
		bekleme += time.Duration(c.rand.Int63n(int64(750 * time.Millisecond)))
		t := time.NewTimer(bekleme)
		select {
		case <-ctx.Done():
			t.Stop()
			return nil, ctx.Err()
		case <-t.C:
		}
	}
	c.mu.Lock()
	c.Kullanici.Basarisiz++
	c.mu.Unlock()
	return nil, sonHata
}

func (c *Client) tekIstek(ctx context.Context, method, yol, action, referer string, gonder []byte) ([]byte, error) {
	u := c.tabanURL + yol
	var govdeOkuyucu io.Reader
	if gonder != nil {
		govdeOkuyucu = bytes.NewReader(gonder)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, govdeOkuyucu)
	if err != nil {
		return nil, &Hata{Tur: HataAg, Mesaj: err.Error()}
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	if action != "" {
		req.Header.Set("X-Domain-Glass-Action", action)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
		req.Header.Set("Origin", c.tabanURL)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &Hata{Tur: HataAg, Mesaj: agMesaji(err)}
	}
	defer resp.Body.Close()
	govde, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, &Hata{Tur: HataAg, Mesaj: err.Error()}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return govde, nil
	}
	return nil, siniflandir(resp.StatusCode, govde, yol)
}

func agMesaji(err error) string {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "istek zaman aşımına uğradı"
	}
	return err.Error()
}

// siniflandir, HTTP durum kodunu ve API hata zarfını tipli hataya çevirir.
func siniflandir(status int, govde []byte, yol string) *Hata {
	var zarf IcerikHatasi
	_ = json.Unmarshal(govde, &zarf)
	mesaj := strings.TrimSpace(zarf.Error.Message)
	if mesaj == "" {
		mesaj = kisaGovde(govde)
	}
	h := &Hata{
		HTTPKod: status,
		Kod:     zarf.Error.Code,
		Mesaj:   mesaj,
		Kaynak:  zarf.Error.Source,
	}
	switch status {
	case http.StatusTooManyRequests:
		h.Tur = HataGecici
		if h.Mesaj == "" {
			h.Mesaj = "hız sınırı aşıldı"
		}
	case http.StatusForbidden:
		// Kapı hataları hız sınırıyla birlikte görülür; yeniden denenebilir.
		switch zarf.Error.Code {
		case "browser_request_required", "browser_required", "click_required", "page_required", "not_available":
			h.Tur = HataGecici
		default:
			h.Tur = HataErisim
		}
	case http.StatusNotFound:
		h.Tur = HataYok
		if h.Mesaj == "" {
			h.Mesaj = "uç nokta bulunamadı"
		}
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		h.Tur = HataGecersiz
	case http.StatusMethodNotAllowed:
		h.Tur = HataGecersiz
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		h.Tur = HataGecici
	default:
		if status >= 500 {
			h.Tur = HataGecici
		} else {
			h.Tur = HataErisim
		}
	}
	return h
}

func kisaGovde(govde []byte) string {
	s := strings.TrimSpace(string(govde))
	if strings.HasPrefix(s, "<") {
		return "sunucu HTML yanıtı döndürdü"
	}
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

func (c *Client) getJSON(ctx context.Context, yol, action, referer string, hedef any) error {
	govde, err := c.istek(ctx, http.MethodGet, yol, action, referer, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(govde, hedef); err != nil {
		return &Hata{Tur: HataBicim, Mesaj: fmt.Sprintf("%s yanıtı çözümlenemedi: %v", yol, err), HTTPKod: 200}
	}
	return nil
}

// AlanAdi, /api/v1/domain/{hedef} çağrısını yapar.
func (c *Client) AlanAdi(ctx context.Context, hedef string) (*AlanAdiYaniti, error) {
	var out AlanAdiYaniti
	yol := "/api/v1/domain/" + url.PathEscape(hedef)
	if err := c.getJSON(ctx, yol, "domain-page-load", c.tabanURL+"/"+hedef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Guvenlik, /api/v1/domain/{hedef}/safety çağrısını yapar.
func (c *Client) Guvenlik(ctx context.Context, hedef string) (*GuvenlikYaniti, error) {
	var out GuvenlikYaniti
	yol := "/api/v1/domain/" + url.PathEscape(hedef) + "/safety"
	if err := c.getJSON(ctx, yol, "domain-safety-load", c.tabanURL+"/"+hedef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Siniflandirma, /api/v1/classifications/{hedef} çağrısını yapar.
func (c *Client) Siniflandirma(ctx context.Context, hedef string) (*SiniflandirmaYaniti, error) {
	var out SiniflandirmaYaniti
	yol := "/api/v1/classifications/" + url.PathEscape(hedef)
	if err := c.getJSON(ctx, yol, "classification-load", c.tabanURL+"/"+hedef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Siralama, /api/v1/rankings/domain/{hedef}?source={kaynak} çağrısını yapar.
func (c *Client) Siralama(ctx context.Context, hedef, kaynak string) (*SiralamaYaniti, error) {
	var out SiralamaYaniti
	yol := "/api/v1/rankings/domain/" + url.PathEscape(hedef) + "?source=" + url.QueryEscape(kaynak)
	if err := c.getJSON(ctx, yol, "", c.tabanURL+"/"+hedef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TLS, /api/v1/tls/{tur}/{hedef} çağrısını yapar.
func (c *Client) TLS(ctx context.Context, tur, hedef string) (*TLSYaniti, error) {
	var out TLSYaniti
	yol := "/api/v1/tls/" + url.PathEscape(tur) + "/" + url.PathEscape(hedef)
	if err := c.getJSON(ctx, yol, "tls-page-load", c.tabanURL+"/"+hedef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WebProfili, /api/v1/web-profile/{hedef} çağrısını yapar.
func (c *Client) WebProfili(ctx context.Context, hedef string) (*WebProfiliYaniti, error) {
	var out WebProfiliYaniti
	yol := "/api/v1/web-profile/" + url.PathEscape(hedef)
	if err := c.getJSON(ctx, yol, "domain-web-profile-load", c.tabanURL+"/"+hedef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IP, /api/v1/ip/{ip} çağrısını yapar.
func (c *Client) IP(ctx context.Context, ip string) (*IPYaniti, error) {
	var out IPYaniti
	yol := "/api/v1/ip/" + url.PathEscape(ip)
	if err := c.getJSON(ctx, yol, "ip-page-load", c.tabanURL+"/"+ip, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ASN, /api/v1/asn/{numara} çağrısını yapar.
func (c *Client) ASN(ctx context.Context, numara int) (*ASNYaniti, error) {
	var out ASNYaniti
	s := strconv.Itoa(numara)
	yol := "/api/v1/asn/" + url.PathEscape(s)
	if err := c.getJSON(ctx, yol, "asn-page-load", c.tabanURL+"/asn/"+s, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TLDiana, /api/v1/tld-iana/{tld} çağrısını yapar.
func (c *Client) TLDiana(ctx context.Context, tld string) (*TLDianaYaniti, error) {
	var out TLDianaYaniti
	yol := "/api/v1/tld-iana/" + url.PathEscape(tld)
	if err := c.getJSON(ctx, yol, "", c.tabanURL+"/"+tld, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Arama, /api/v1/search?q= sorgusunu yapar.
func (c *Client) Arama(ctx context.Context, sorgu string) (*AramaYaniti, error) {
	var out AramaYaniti
	yol := "/api/v1/search?q=" + url.QueryEscape(sorgu)
	if err := c.getJSON(ctx, yol, "search-click", c.tabanURL+"/search?q="+url.QueryEscape(sorgu), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WHOIS, /api/v1/whois/{hedef} POST çağrısını yapar.
func (c *Client) WHOIS(ctx context.Context, hedef string) (*WHOISYaniti, error) {
	var out WHOISYaniti
	yol := "/api/v1/whois/" + url.PathEscape(hedef)
	govde, err := c.istek(ctx, http.MethodPost, yol, "whois-click", c.tabanURL+"/"+hedef, []byte{})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(govde, &out); err != nil {
		return nil, &Hata{Tur: HataBicim, Mesaj: fmt.Sprintf("WHOIS yanıtı çözümlenemedi: %v", err), HTTPKod: 200}
	}
	return &out, nil
}

// Akis, RSS akışını ham XML olarak indirir.
func (c *Client) Akis(ctx context.Context, ad string) ([]byte, error) {
	u := c.tabanURL + "/feeds/" + url.PathEscape(ad) + ".xml"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, &Hata{Tur: HataAg, Mesaj: err.Error()}
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")
	if err := c.bekle(ctx); err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &Hata{Tur: HataAg, Mesaj: agMesaji(err)}
	}
	defer resp.Body.Close()
	govde, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, &Hata{Tur: HataAg, Mesaj: err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, siniflandir(resp.StatusCode, govde, u)
	}
	return govde, nil
}

// SaglikKontrolu, /api/v1/domain/example.com üzerinden servis erişimini sınar.
func (c *Client) SaglikKontrolu(ctx context.Context) error {
	_, err := c.AlanAdi(ctx, "example.com")
	return err
}
