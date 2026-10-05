package out

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/void0x14/domainglass/internal/model"
)

func ornekRapor() *model.Rapor {
	sira := 5098
	degisim := -69
	kalanGun := 82
	kayitli := true
	imzali := true
	return &model.Rapor{
		Hedef:     "example.com",
		Tur:       "domain",
		Arac:      model.AracBilgisi{Ad: "domainglass", Commit: "3dbbad9", Go: "go1.21", Platform: "linux/amd64"},
		Olusturma: "2026-10-05T00:00:00Z",
		Ozet: model.Ozet{
			Kayitli:      &kayitli,
			CiscoSira:    &sira,
			DNSSECImzali: &imzali,
			TLSDurumu:    "available",
		},
		Siralama: &model.SiralamaBolumu{
			Cisco: &model.Siralama{Domain: "example.com", Provider: "cisco", CurrentRank: &sira, Change: &degisim, Listed: true},
		},
		Kayit: &model.KayitBolumu{Durum: "registered", Kaynak: "rdap", Kayitci: "IANA"},
		DNS: &model.DNSBolumuRapor{
			Durum: "NOERROR",
			Kayitlar: map[string][]model.DNSKaydiRapor{
				"A": {{Ad: "example.com", TTL: 88, Deger: "104.20.23.154"}},
			},
		},
		TLS: &model.TLSBolumu{
			Durum: "available", Dogrulandi: true, KalanGun: &kalanGun,
			Konu:  model.KimlikBilgisi{OrtakAdlar: []string{"example.com"}},
			Adlar: []model.TLSAdi{{Ad: "example.com"}, {Ad: "*.example.com", Joker: true}},
		},
		Web: &model.WebBolumu{DurumKodu: 200, DurumMetni: "OK", SonURL: "https://example.com/", Meta: map[string]string{"title": "Example Domain"}},
		Guvenlik: &model.GuvenlikBolumu{
			Kategoriler:  []model.GuvenlikKategorisiRapor{{Kategori: "malware", Etiket: "Zararlı yazılım", Durum: "not_flagged"}},
			Saglayicilar: []model.SaglayiciGuvenlik{{ID: "quad9", Ad: "Quad9", Durum: "ok", Sinyaller: map[string]string{"malware": "not_flagged"}}},
		},
		IP: &model.IPBolumu{Adres: "1.1.1.1", Surum: 4, GenelMi: true, Duyurulmus: true, Prefix: "1.1.1.0/24"},
		Kesif: model.KesifBolumu{
			AltAlanlar: []model.KesifOgesi{{Deger: "cdn.example.com", Kaynaklar: []string{"DNS CNAME"}}},
			Iliskili:   []model.KesifOgesi{},
			IPler:      []model.KesifOgesi{{Deger: "104.20.23.154", Kaynaklar: []string{"DNS A"}}},
			Epostalar:  []model.KesifOgesi{},
			Telefonlar: []model.KesifOgesi{},
			Kurumlar:   []model.KesifOgesi{},
		},
		Siniflandirma:  &model.SiniflandirmaBolumu{EslesmeSayisi: 0},
		ASN:            &model.ASNBolumu{Numara: 13335, Etiket: "AS13335"},
		TLD:            &model.TLDBolumu{TLD: "com"},
		Akis:           &model.AkisBolumu{Baslik: "Test", Ogeler: []model.AkisOgesi{}},
		Cozumleyiciler: []model.CozumleyiciSonucu{{ID: "quad9", Ad: "Quad9", Durum: "ok"}},
		Kaynaklar:      []model.KaynakDurumu{{ID: "domain", Ad: "DNS", Durum: "ok"}},
		Uyarilar:       []string{},
	}
}

func TestJSONCikti(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, ornekRapor(), true); err != nil {
		t.Fatalf("JSON hata: %v", err)
	}
	var genel map[string]any
	if err := json.Unmarshal(buf.Bytes(), &genel); err != nil {
		t.Fatalf("çıktı geçerli JSON değil: %v", err)
	}
	for _, anahtar := range []string{"target", "kind", "tool", "generatedAt", "summary", "discovery", "sources", "warnings"} {
		if _, ok := genel[anahtar]; !ok {
			t.Errorf("zorunlu alan eksik: %s", anahtar)
		}
	}
	if _, ok := genel["hamAlanAdi"]; ok {
		t.Error("ham API alanları JSON'a sızmamalı")
	}
}

func TestNDJSONTekSatir(t *testing.T) {
	var buf bytes.Buffer
	if err := NDJSON(&buf, ornekRapor()); err != nil {
		t.Fatalf("NDJSON hata: %v", err)
	}
	satirlar := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(satirlar) != 1 {
		t.Fatalf("NDJSON tek satır olmalı, %d satır", len(satirlar))
	}
	var genel map[string]any
	if err := json.Unmarshal([]byte(satirlar[0]), &genel); err != nil {
		t.Fatalf("geçersiz JSON: %v", err)
	}
}

func TestAlanSecimi(t *testing.T) {
	secili, bilinmeyen, err := AlanSecimi(ornekRapor(), []string{"target", "discovery"})
	if err != nil {
		t.Fatalf("hata: %v", err)
	}
	if len(bilinmeyen) != 0 {
		t.Fatalf("bilinmeyen alan olmamalı: %v", bilinmeyen)
	}
	if len(secili) != 2 {
		t.Fatalf("2 alan bekleniyordu: %v", secili)
	}
	if secili["target"] != "example.com" {
		t.Errorf("hedef hatalı: %v", secili["target"])
	}
	if _, ok := secili["dns"]; ok {
		t.Error("seçilmeyen alan çıktıda olmamalı")
	}
}

func TestAlanSecimiBilinmeyen(t *testing.T) {
	_, bilinmeyen, err := AlanSecimi(ornekRapor(), []string{"target", "yokboyle"})
	if err != nil {
		t.Fatalf("hata: %v", err)
	}
	if len(bilinmeyen) != 1 || bilinmeyen[0] != "yokboyle" {
		t.Fatalf("bilinmeyen alan raporlanmadı: %v", bilinmeyen)
	}
}

func TestKullanilabilirAlanlarSemaIleUyumlu(t *testing.T) {
	govde, err := json.Marshal(ornekRapor())
	if err != nil {
		t.Fatal(err)
	}
	var genel map[string]any
	if err := json.Unmarshal(govde, &genel); err != nil {
		t.Fatal(err)
	}
	for _, a := range KullanilabilirAlanlar() {
		if _, ok := genel[a]; !ok {
			t.Errorf("KullanilabilirAlanlar listesindeki %q rapor JSON'unda yok", a)
		}
	}
}

func TestInsanCikti(t *testing.T) {
	var buf bytes.Buffer
	Yaz(&buf, ornekRapor(), Renk{Acik: false})
	cikti := buf.String()
	beklenen := []string{
		"DOMAINGLASS",
		"Popülerlik sıralamaları",
		"Cisco Umbrella",
		"Alan adı kaydı",
		"DNS kayıtları",
		"TLS sertifikası",
		"Ana sayfa profili",
		"DNS filtreleri",
		"Keşfedilen varlıklar",
		"cdn.example.com",
	}
	for _, b := range beklenen {
		if !strings.Contains(cikti, b) {
			t.Errorf("insan çıktısında %q eksik", b)
		}
	}
	if strings.Contains(cikti, "\x1b[") {
		t.Error("renk kapalıyken ANSI kaçış dizisi üretilmemeli")
	}
}

func TestRenkAcikCikti(t *testing.T) {
	var buf bytes.Buffer
	Yaz(&buf, ornekRapor(), Renk{Acik: true})
	if !strings.Contains(buf.String(), "\x1b[") {
		t.Error("renk açıkken ANSI dizisi bekleniyordu")
	}
}

func TestOzetSatirlari(t *testing.T) {
	satirlar := OzetSatirlari(ornekRapor())
	if len(satirlar) == 0 {
		t.Fatal("özet satırı üretilmedi")
	}
	birlesik := strings.Join(satirlar, " | ")
	for _, b := range []string{"kayıt durumu: kayıtlı", "Cisco sıra: #5098", "DNSSEC: imzalı"} {
		if !strings.Contains(birlesik, b) {
			t.Errorf("özet satırında %q eksik: %s", b, birlesik)
		}
	}
}
