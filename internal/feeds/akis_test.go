package feeds

import (
	"strings"
	"testing"
)

const ornekAkis = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Top domain rank gainers</title>
    <link>https://domain.glass/</link>
    <description>Domains with the largest measured one-day gains.</description>
    <lastBuildDate>Mon, 05 Oct 2026 04:31:47 GMT</lastBuildDate>
    <item>
      <title>eventtracking-ap1.hubapi.com — up 973366</title>
      <link>https://domain.glass/eventtracking-ap1.hubapi.com</link>
      <guid isPermaLink="true">https://domain.glass/eventtracking-ap1.hubapi.com</guid>
      <description>Gained 973366 places to rank 7514.</description>
      <pubDate>Mon, 05 Oct 2026 04:31:47 GMT</pubDate>
      <category>rank-gain</category>
    </item>
    <item>
      <title>web.mashov.info — up 952918</title>
      <link>https://domain.glass/web.mashov.info</link>
      <description>Gained 952918 places.</description>
      <category>rank-gain</category>
    </item>
  </channel>
</rss>`

func TestCozumle(t *testing.T) {
	bolum, err := Cozumle([]byte(ornekAkis))
	if err != nil {
		t.Fatalf("Cozumle hata: %v", err)
	}
	if bolum.Baslik != "Top domain rank gainers" {
		t.Errorf("başlık hatalı: %q", bolum.Baslik)
	}
	if bolum.OgeSayisi != 2 {
		t.Fatalf("2 öğe bekleniyordu, %d alındı", bolum.OgeSayisi)
	}
	if bolum.Ogeler[0].Baglanti != "https://domain.glass/eventtracking-ap1.hubapi.com" {
		t.Errorf("bağlantı hatalı: %q", bolum.Ogeler[0].Baglanti)
	}
	if bolum.Guncelleme == "" {
		t.Error("lastBuildDate boş")
	}
}

func TestHostlar(t *testing.T) {
	bolum, err := Cozumle([]byte(ornekAkis))
	if err != nil {
		t.Fatal(err)
	}
	hostlar := Hostlar(bolum)
	if len(hostlar) != 2 {
		t.Fatalf("2 host bekleniyordu: %v", hostlar)
	}
	if hostlar[0] != "eventtracking-ap1.hubapi.com" || hostlar[1] != "web.mashov.info" {
		t.Fatalf("hostlar hatalı: %v", hostlar)
	}
}

func TestHostlarHariciBaglanti(t *testing.T) {
	bolum, _ := Cozumle([]byte(`<rss><channel><item><link>https://example.com/x</link></item><item><link>https://domain.glass/ok.com</link></item></channel></rss>`))
	hostlar := Hostlar(bolum)
	if len(hostlar) != 1 || hostlar[0] != "ok.com" {
		t.Fatalf("yalnızca domain.glass bağlantısı bekleniyordu: %v", hostlar)
	}
}

func TestCozumleBozuk(t *testing.T) {
	if _, err := Cozumle([]byte("bu xml degil")); err == nil {
		t.Fatal("bozuk girdi için hata bekleniyordu")
	}
}

func TestAdlar(t *testing.T) {
	for _, ad := range []string{"trending", "top-gainers", "newly-ranked"} {
		if _, ok := Adlar[ad]; !ok {
			t.Errorf("%s akışı tanımlı değil", ad)
		}
	}
	if !strings.Contains(Adlar["top-gainers"], "yükselen") {
		t.Error("top-gainers açıklaması Türkçe olmalı")
	}
}
