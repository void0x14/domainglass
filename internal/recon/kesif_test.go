package recon

import (
	"testing"

	"github.com/void0x14/domainglass/internal/model"
)

func TestHostTemizle(t *testing.T) {
	durumlar := []struct{ girdi, beklenen string }{
		{"*.test.com", "test.com"},
		{"https://app.test.com/path", "app.test.com"},
		{"ns1.test.com.", "ns1.test.com"},
		{"  Sub.Test.COM:8080  ", "sub.test.com"},
		{"[2606:4700::1111]", "2606:4700::1111"},
	}
	for _, d := range durumlar {
		if got := HostTemizle(d.girdi); got != d.beklenen {
			t.Errorf("HostTemizle(%q) = %q; beklenen %q", d.girdi, got, d.beklenen)
		}
	}
}

func TestIPMi(t *testing.T) {
	if !IPMi("1.1.1.1") || !IPMi("2606:4700::1111") {
		t.Fatal("geçerli IP adresleri tanınmadı")
	}
	if IPMi("example.com") || IPMi("") {
		t.Fatal("IP olmayan girdiler IP sayıldı")
	}
}

func TestKesifKorelasyonu(t *testing.T) {
	d := &model.AlanAdiBilgisi{}
	d.DNS.Records.A = []model.DNSRecord{{Value: "93.184.216.34"}}
	d.DNS.Records.CNAME = []model.DNSRecord{{Value: "cdn.example.com"}}
	d.DNS.Records.NS = []model.DNSRecord{{Value: "ns1.iana.org."}}
	d.DNS.Records.MX = []model.DNSRecord{{Value: "10 mail.example.com."}}
	d.RDAP.Nameservers = []string{"elliott.ns.cloudflare.com"}

	tls := &model.TLSBilgisi{}
	tls.Certificate = &struct {
		Subject struct {
			CommonNames         []string `json:"commonNames"`
			Organizations       []string `json:"organizations"`
			OrganizationalUnits []string `json:"organizationalUnits"`
			Localities          []string `json:"localities"`
			States              []string `json:"states"`
			Countries           []string `json:"countries"`
		} `json:"subject"`
		Issuer struct {
			CommonNames         []string `json:"commonNames"`
			Organizations       []string `json:"organizations"`
			OrganizationalUnits []string `json:"organizationalUnits"`
			Localities          []string `json:"localities"`
			States              []string `json:"states"`
			Countries           []string `json:"countries"`
		} `json:"issuer"`
		ValidFrom          string `json:"validFrom"`
		ValidTo            string `json:"validTo"`
		DaysRemaining      *int   `json:"daysRemaining"`
		SignatureAlgorithm string `json:"signatureAlgorithm"`
		KeyType            string `json:"keyType"`
		KeyBits            int    `json:"keyBits"`
		KeyCurve           string `json:"keyCurve"`
		SerialNumber       string `json:"serialNumber"`
		FingerprintSHA256  string `json:"fingerprintSha256"`
		DNSNames           []struct {
			Value        string `json:"value"`
			Unicode      string `json:"unicode"`
			Wildcard     bool   `json:"wildcard"`
			Relationship string `json:"relationship"`
		} `json:"dnsNames"`
		IPAddresses     []string `json:"ipAddresses"`
		OmittedSANCount int      `json:"omittedSanCount"`
	}{}
	tls.Certificate.DNSNames = []struct {
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

	web := &model.WebProfili{
		DiscoveredURLs: []string{"https://api.example.com/v1", "https://partner.com/login"},
	}

	kesif := Kesif(KesifGirdisi{Hedef: "example.com", Kok: "example.com", AlanAdi: d, TLS: tls, Web: web})

	bekle := func(liste []model.KesifOgesi, deger string) *model.KesifOgesi {
		for i := range liste {
			if liste[i].Deger == deger {
				return &liste[i]
			}
		}
		return nil
	}

	for _, ad := range []string{"cdn.example.com", "sub.example.com", "api.example.com", "mail.example.com"} {
		if bekle(kesif.AltAlanlar, ad) == nil {
			t.Errorf("alt alan adı eksik: %s (mevcut: %+v)", ad, kesif.AltAlanlar)
		}
	}
	for _, ad := range []string{"ns1.iana.org", "otherdomain.com", "elliott.ns.cloudflare.com", "partner.com"} {
		if bekle(kesif.Iliskili, ad) == nil {
			t.Errorf("ilişkili alan adı eksik: %s (mevcut: %+v)", ad, kesif.Iliskili)
		}
	}
	if bekle(kesif.IPler, "93.184.216.34") == nil {
		t.Errorf("IP eksik: %+v", kesif.IPler)
	}
}

func TestKesifKaynakEtiketleri(t *testing.T) {
	d := &model.AlanAdiBilgisi{}
	d.DNS.Records.NS = []model.DNSRecord{{Name: "example.com", Type: "NS", Value: "ns1.example.com."}}
	kesif := Kesif(KesifGirdisi{Hedef: "example.com", Kok: "example.com", AlanAdi: d})
	for _, oge := range kesif.AltAlanlar {
		if oge.Deger == "ns1.example.com" {
			if len(oge.Kaynaklar) == 0 || oge.Kaynaklar[0] != "DNS NS" {
				t.Fatalf("kaynak etiketi hatalı: %+v", oge.Kaynaklar)
			}
			return
		}
	}
	t.Fatal("ns1.example.com bulunamadı")
}

func TestKesifBosListeJSON(t *testing.T) {
	kesif := Kesif(KesifGirdisi{Hedef: "example.com", Kok: "example.com"})
	if kesif.AltAlanlar == nil || kesif.Iliskili == nil || kesif.IPler == nil {
		t.Fatal("boş keşif listeleri nil olmamalı (JSON [] üretmeli)")
	}
}
