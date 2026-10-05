package dg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func testIstemci(sunucu *httptest.Server) *Client {
	return Yeni(TabanURL(sunucu.URL), AsgariAralik(0), YenidenDeneme(3))
}

// TestBasliklar, istemcinin site sözleşmesine uyan başlıkları gönderdiğini doğrular.
func TestBasliklar(t *testing.T) {
	var gorulen http.Header
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gorulen = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"target":{"ascii":"example.com"}}`))
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	if _, err := c.AlanAdi(context.Background(), "example.com"); err != nil {
		t.Fatalf("AlanAdi hata: %v", err)
	}
	if got := gorulen.Get("User-Agent"); got == "" || got == "Go-http-client/1.1" {
		t.Errorf("tarayıcı User-Agent bekleniyordu, alınan %q", got)
	}
	if got := gorulen.Get("Accept"); got != "application/json" {
		t.Errorf("Accept istenen application/json, alınan %q", got)
	}
	for _, h := range []string{"Sec-Fetch-Site", "Sec-Fetch-Mode", "Sec-Fetch-Dest"} {
		if gorulen.Get(h) == "" {
			t.Errorf("%s başlığı eksik", h)
		}
	}
	if got := gorulen.Get("X-Domain-Glass-Action"); got != "domain-page-load" {
		t.Errorf("action başlığı istenen domain-page-load, alınan %q", got)
	}
	if got := gorulen.Get("Origin"); got != sunucu.URL {
		t.Errorf("Origin istenen %q, alınan %q", sunucu.URL, got)
	}
	if got := gorulen.Get("Referer"); got != sunucu.URL+"/example.com" {
		t.Errorf("Referer istenen %q, alınan %q", sunucu.URL+"/example.com", got)
	}
}

// TestWHOISPost, WHOIS çağrısının POST ve whois-click action ile yapıldığını doğrular.
func TestWHOISPost(t *testing.T) {
	var method, action string
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		action = r.Header.Get("X-Domain-Glass-Action")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"registered":true,"registrar":{"name":"Test"}}`))
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	if _, err := c.WHOIS(context.Background(), "example.com"); err != nil {
		t.Fatalf("WHOIS hata: %v", err)
	}
	if method != http.MethodPost {
		t.Errorf("POST bekleniyordu, alınan %s", method)
	}
	if action != "whois-click" {
		t.Errorf("whois-click action bekleniyordu, alınan %q", action)
	}
}

// TestGecici403YenidenDeneme, tarayıcı kapısı hatasının geçici sayıldığını ve
// yeniden denendiğini doğrular. Gerçek sunucu, hız sınırında bu kodu döner.
func TestGecici403YenidenDeneme(t *testing.T) {
	var sayac int32
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&sayac, 1)
		if n < 3 {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":"browser_request_required","message":"Certificate analysis is available only from the matching domain.glass detail page."}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"target":"example.com","status":"available"}`))
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	yanit, err := c.TLS(context.Background(), "domain", "example.com")
	if err != nil {
		t.Fatalf("yeniden deneme sonrası başarı bekleniyordu: %v", err)
	}
	if yanit.Status != "available" {
		t.Fatalf("durum istenen available, alınan %q", yanit.Status)
	}
	if atomic.LoadInt32(&sayac) != 3 {
		t.Fatalf("3 deneme bekleniyordu, %d yapıldı", sayac)
	}
}

// Test429YenidenDeneme, hız sınırı yanıtının geçici sayıldığını doğrular.
func Test429YenidenDeneme(t *testing.T) {
	var sayac int32
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&sayac, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"domain":"example.com","provider":"cisco","listed":true,"currentRank":10}`))
	}))
	defer sunucu.Close()

	c := Yeni(TabanURL(sunucu.URL), AsgariAralik(0), YenidenDeneme(3))
	s, err := c.Siralama(context.Background(), "example.com", "cisco")
	if err != nil {
		t.Fatalf("hata: %v", err)
	}
	if !s.Listed || s.CurrentRank == nil || *s.CurrentRank != 10 {
		t.Fatalf("sıralama hatalı: %+v", s)
	}
}

// TestKaliciHataYenidenDenenmez, doğrulama hatasının yeniden denenmediğini doğrular.
func TestKaliciHataYenidenDenenmez(t *testing.T) {
	var sayac int32
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&sayac, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"source":"validation","code":"ip_address","message":"IP addresses are not domain lookups."}}`))
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	_, err := c.AlanAdi(context.Background(), "1.1.1.1")
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
	if !GeciciMi(err) && !isGecersiz(err) {
		t.Fatalf("geçersiz sınıfı bekleniyordu, alınan: %v", err)
	}
	if atomic.LoadInt32(&sayac) != 1 {
		t.Fatalf("1 deneme bekleniyordu, %d yapıldı", sayac)
	}
}

func isGecersiz(err error) bool {
	var h *Hata
	if e, ok := err.(*Hata); ok {
		h = e
	}
	return h != nil && h.Tur == HataGecersiz
}

// TestYokHatasi, 404 yanıtının HataYok olarak sınıflandığını doğrular.
func TestYokHatasi(t *testing.T) {
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found","message":"The requested route was not found."}}`))
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	_, err := c.TLDiana(context.Background(), "com")
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
	if !YokMu(err) {
		t.Fatalf("HataYok bekleniyordu, alınan: %v", err)
	}
}

// TestHizSinirlama, asgari istek aralığının uygulandığını doğrular.
func TestHizSinirlama(t *testing.T) {
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"target":{"ascii":"example.com"}}`))
	}))
	defer sunucu.Close()

	c := Yeni(TabanURL(sunucu.URL), AsgariAralik(120*time.Millisecond), YenidenDeneme(1))
	baslangic := time.Now()
	for i := 0; i < 3; i++ {
		if _, err := c.AlanAdi(context.Background(), "example.com"); err != nil {
			t.Fatalf("hata: %v", err)
		}
	}
	gecen := time.Since(baslangic)
	// İlk istek beklemesiz; sonraki ikisi en az 120 ms arayla: toplam >= 240 ms.
	if gecen < 240*time.Millisecond {
		t.Fatalf("hız sınırı uygulanmadı: %v", gecen)
	}
}

// TestBicimHatasi, geçersiz JSON yanıtının HataBicim olarak sınıflandığını doğrular.
func TestBicimHatasi(t *testing.T) {
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{bu gecerli json degil`))
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	_, err := c.AlanAdi(context.Background(), "example.com")
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
	h, ok := err.(*Hata)
	if !ok || h.Tur != HataBicim {
		t.Fatalf("HataBicim bekleniyordu, alınan: %v", err)
	}
}

// TestAlanAdiCozumleme, gerçek API yanıt biçiminin doğru çözümlendiğini doğrular.
func TestAlanAdiCozumleme(t *testing.T) {
	ornek := map[string]any{
		"target": map[string]any{"ascii": "example.com", "kind": "domain", "tld": "com"},
		"dns": map[string]any{
			"status": "NOERROR",
			"records": map[string]any{
				"A": []map[string]any{{"name": "example.com", "type": "A", "ttl": 88, "value": "104.20.23.154"}},
			},
		},
		"rdap": map[string]any{"registrationStatus": "registered", "registrar": map[string]any{"name": "IANA", "id": "376"}},
	}
	govde, _ := json.Marshal(ornek)
	sunucu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(govde)
	}))
	defer sunucu.Close()

	c := testIstemci(sunucu)
	bilgi, err := c.AlanAdi(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("hata: %v", err)
	}
	if bilgi.Target.Ascii != "example.com" || bilgi.DNS.Status != "NOERROR" {
		t.Fatalf("çözümleme hatalı: %+v", bilgi)
	}
	if len(bilgi.DNS.Records.A) != 1 || bilgi.DNS.Records.A[0].Value != "104.20.23.154" {
		t.Fatalf("A kaydı hatalı: %+v", bilgi.DNS.Records.A)
	}
	if bilgi.RDAP.Registrar.Name != "IANA" {
		t.Fatalf("RDAP hatalı: %+v", bilgi.RDAP)
	}
}
