// Package feeds, domain.glass RSS akışlarını ayrıştırır.
//
// Akışlar, Cisco Umbrella ve Tranco popülerlik listelerindeki günlük
// hareketleri (yükselenler, yeni girenler, trend) yayınlar. Bu uç noktalar
// hız sınırı kapısına takılmaz; toplu keşif için uygundur.
package feeds

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/void0x14/domainglass/internal/model"
)

// Adlar, desteklenen akış adlarıdır.
var Adlar = map[string]string{
	"trending":            "Cisco Umbrella — trend olan alan adları",
	"trending-tranco":     "Tranco — trend olan alan adları",
	"top-gainers":         "Cisco Umbrella — en çok yükselenler",
	"top-gainers-tranco":  "Tranco — en çok yükselenler",
	"newly-ranked":        "Cisco Umbrella — listeye yeni girenler",
	"newly-ranked-tranco": "Tranco — listeye yeni girenler",
}

type rss struct {
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		LastBuild   string `xml:"lastBuildDate"`
		Items       []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Category    string `xml:"category"`
		} `xml:"item"`
	} `xml:"channel"`
}

// Cozumle, RSS iletimini rapor bölümüne çevirir.
func Cozumle(veri []byte) (*model.AkisBolumu, error) {
	var r rss
	dec := xml.NewDecoder(strings.NewReader(string(veri)))
	dec.Strict = false
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("RSS akışı ayrıştırılamadı: %w", err)
	}
	bolum := &model.AkisBolumu{
		Baslik:     strings.TrimSpace(r.Channel.Title),
		Aciklama:   strings.TrimSpace(r.Channel.Description),
		Guncelleme: strings.TrimSpace(r.Channel.LastBuild),
		Ogeler:     make([]model.AkisOgesi, 0, len(r.Channel.Items)),
	}
	for _, it := range r.Channel.Items {
		bolum.Ogeler = append(bolum.Ogeler, model.AkisOgesi{
			Baslik:      strings.TrimSpace(it.Title),
			Baglanti:    strings.TrimSpace(it.Link),
			Aciklama:    strings.TrimSpace(it.Description),
			YayimTarihi: strings.TrimSpace(it.PubDate),
			Kategori:    strings.TrimSpace(it.Category),
		})
	}
	bolum.OgeSayisi = len(bolum.Ogeler)
	return bolum, nil
}

// Hostlar, akış öğelerinden alan adı ana bilgisayarlarını çıkarır.
// domain.glass bağlantıları https://domain.glass/<host> biçimindedir.
func Hostlar(bolum *model.AkisBolumu) []string {
	gorulen := make(map[string]bool)
	var out []string
	for _, it := range bolum.Ogeler {
		h := baglantidanHost(it.Baglanti)
		if h == "" || gorulen[h] {
			continue
		}
		gorulen[h] = true
		out = append(out, h)
	}
	return out
}

func baglantidanHost(baglanti string) string {
	baglanti = strings.TrimSpace(baglanti)
	if baglanti == "" {
		return ""
	}
	const onek = "https://domain.glass/"
	if !strings.HasPrefix(baglanti, onek) {
		return ""
	}
	h := strings.TrimPrefix(baglanti, onek)
	h = strings.Trim(h, "/")
	if h == "" || strings.ContainsAny(h, "?#") {
		return ""
	}
	return strings.ToLower(h)
}

// Oku, okuyucudan RSS akışı ayrıştırır.
func Oku(rd io.Reader) (*model.AkisBolumu, error) {
	veri, err := io.ReadAll(io.LimitReader(rd, 32<<20))
	if err != nil {
		return nil, err
	}
	return Cozumle(veri)
}
