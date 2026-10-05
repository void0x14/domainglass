package cli

import (
	"fmt"
	"sort"
	"strings"
)

// Bayrak, desteklenen bir komut satırı bayrağıdır.
type Bayrak struct {
	Adlar    []string
	Degerli  bool
	Aciklama string
}

// Bayraklar, desteklenen tüm bayrakların tanımıdır.
var Bayraklar = []Bayrak{
	{Adlar: []string{"json"}, Aciklama: "Raporu okunabilir JSON olarak yazdır"},
	{Adlar: []string{"ndjson"}, Aciklama: "Raporu tek satır JSON (NDJSON) olarak yazdır"},
	{Adlar: []string{"alan", "fields"}, Degerli: true, Aciklama: "JSON çıktısında yalnızca virgülle ayrılmış üst düzey alanları döndür"},
	{Adlar: []string{"subs", "subdomains"}, Aciklama: "Yalnızca keşfedilen alt alan adlarını satır satır yazdır"},
	{Adlar: []string{"ips"}, Aciklama: "Yalnızca keşfedilen IP adreslerini satır satır yazdır"},
	{Adlar: []string{"related"}, Aciklama: "Yalnızca ilişkili kök alan adlarını satır satır yazdır"},
	{Adlar: []string{"emails"}, Aciklama: "Yalnızca keşfedilen e-posta adreslerini satır satır yazdır"},
	{Adlar: []string{"orgs", "organizations"}, Aciklama: "Yalnızca keşfedilen kurum adlarını satır satır yazdır"},
	{Adlar: []string{"cozumleyici", "resolvers"}, Degerli: true, Aciklama: "Bağımsız DNS filtre çözümleyicilerini çalıştır (virgülle ayrılmış ID)"},
	{Adlar: []string{"timeout"}, Degerli: true, Aciklama: "İstek zaman aşımı saniye cinsinden (varsayılan 20)"},
	{Adlar: []string{"hiz", "rate"}, Degerli: true, Aciklama: "İstekler arası asgari aralık milisaniye cinsinden (varsayılan 1100)"},
	{Adlar: []string{"renk", "color"}, Degerli: true, Aciklama: "Renk kipi: auto | always | never"},
	{Adlar: []string{"sessiz", "quiet"}, Aciklama: "İlerleme ve uyarı günlüklerini stderr üzerinde bastır"},
	{Adlar: []string{"yardim", "help"}, Aciklama: "Kullanım bilgisini yazdır"},
	{Adlar: []string{"surum", "version"}, Aciklama: "Sürüm bilgisini yazdır"},
	{Adlar: []string{"sema", "schema"}, Aciklama: "Rapor JSON şemasını yazdır"},
	{Adlar: []string{"yetenek", "capabilities"}, Aciklama: "Makineler için komut/bayrak/çıkış kodu kataloğunu JSON olarak yazdır"},
	{Adlar: []string{"girintisiz"}, Aciklama: "JSON çıktısını girintisiz yazdır"},
}

// Komut, tek bir alt komut tanımıdır.
type Komut struct {
	Ad    string `json:"name"`
	Ozet  string `json:"summary"`
	Ornek string `json:"example"`
}

// Komutlar, desteklenen alt komutların tanımıdır.
var Komutlar = []Komut{
	{Ad: "domain", Ozet: "Alan adı istihbaratı (DNS, RDAP, WHOIS, TLS, web profili, sıralamalar, filtreler)", Ornek: "domainglass domain example.com"},
	{Ad: "ip", Ozet: "IP adresi istihbaratı (BGP/ASN, konum, ters DNS, TLS)", Ornek: "domainglass ip 1.1.1.1"},
	{Ad: "asn", Ozet: "ASN istihbaratı (sahip, prefix, komşuluk)", Ornek: "domainglass asn 13335"},
	{Ad: "tld", Ozet: "IANA TLD kaydı (adres, iletişim, ad sunucuları)", Ornek: "domainglass tld com"},
	{Ad: "akis", Ozet: "Popülerlik RSS akışları (trending, top-gainers, newly-ranked)", Ornek: "domainglass akis top-gainers"},
	{Ad: "toplu", Ozet: "stdin üzerinden çoklu hedef tarama", Ornek: "cat hedefler.txt | domainglass toplu -subs"},
	{Ad: "saglik", Ozet: "domain.glass servis erişimini sınar", Ornek: "domainglass saglik"},
	{Ad: "sema", Ozet: "Rapor JSON şemasını yazdırır", Ornek: "domainglass sema"},
	{Ad: "yetenek", Ozet: "Makine kataloğunu yazdırır", Ornek: "domainglass yetenek"},
	{Ad: "surum", Ozet: "Sürüm bilgisini yazdırır", Ornek: "domainglass surum"},
	{Ad: "yardim", Ozet: "Bu kullanım bilgisini yazdırır", Ornek: "domainglass yardim"},
}

// CikisKodu, tek bir çıkış kodu tanımıdır.
type CikisKodu struct {
	Kod      int    `json:"code"`
	Ad       string `json:"name"`
	Aciklama string `json:"description"`
}

// CikisKodlari, süreç çıkış kodlarının sözleşmesidir.
var CikisKodlari = []CikisKodu{
	{Kod: 0, Ad: "basarili", Aciklama: "Sorgu tamamlandı; veri mevcut."},
	{Kod: 1, Ad: "hata", Aciklama: "Parametre, ağ veya beklenmeyen hata."},
	{Kod: 2, Ad: "bulunamadi", Aciklama: "Hedef kayıtlı değil (NXDOMAIN veya RDAP not_found)."},
	{Kod: 3, Ad: "hiz_siniri", Aciklama: "domain.glass hız sınırı; hiçbir kaynak veri döndürmedi."},
	{Kod: 4, Ad: "kullanim", Aciklama: "Geçersiz komut satırı kullanımı."},
}

// Kullanim, insan okunur Türkçe yardım metnini döner.
func Kullanim() string {
	var b strings.Builder
	b.WriteString("domainglass — domain.glass istihbaratını terminale taşıyan araç\n\n")
	b.WriteString("KULLANIM\n")
	b.WriteString("  domainglass <hedef> [bayraklar]         Alan adı veya IP (tür otomatik sezilir)\n")
	b.WriteString("  domainglass <komut> [argüman] [bayrak]\n")
	b.WriteString("  <stdin> | domainglass toplu [bayraklar]\n\n")
	b.WriteString("KOMUTLAR\n")
	for _, k := range Komutlar {
		fmt.Fprintf(&b, "  %-10s %s\n", k.Ad, k.Ozet)
	}
	b.WriteString("\nBAYRAKLAR\n")
	for _, f := range Bayraklar {
		ad := "-" + strings.Join(f.Adlar, " | -")
		if f.Degerli {
			ad += " <deger>"
		}
		fmt.Fprintf(&b, "  %-44s %s\n", ad, f.Aciklama)
	}
	b.WriteString("\nÇIKIŞ KODLARI\n")
	for _, c := range CikisKodlari {
		fmt.Fprintf(&b, "  %d  %-14s %s\n", c.Kod, c.Ad, c.Aciklama)
	}
	b.WriteString("\nÖRNEKLER\n")
	ornekler := []string{
		"domainglass example.com",
		"domainglass ip 1.1.1.1",
		"domainglass asn 13335",
		"domainglass -json example.com > rapor.json",
		"domainglass -alan summary,discovery -json example.com",
		"domainglass -subs tesla.com | httpx -silent",
		"domainglass -cozumleyici quad9,adguard example.com",
		"cat hedefler.txt | domainglass toplu -ndjson",
		"domainglass akis top-gainers -subs",
	}
	for _, o := range ornekler {
		fmt.Fprintf(&b, "  %s\n", o)
	}
	b.WriteString("\nAI AJANLARI İÇİN\n")
	fmt.Fprintf(&b, "  domainglass yetenek      # makine kataloğu (JSON)\n")
	fmt.Fprintf(&b, "  domainglass sema         # rapor JSON şeması\n")
	fmt.Fprintf(&b, "  AGENTS.md ve docs/AI-AJANLARI.md dosyalarına bakın.\n")
	return b.String()
}

// YetenekKatalogu, makineler için tam kataloğu JSON'a uygun biçimde döner.
func YetenekKatalogu(surum string) map[string]any {
	bayraklar := make([]map[string]any, 0, len(Bayraklar))
	for _, f := range Bayraklar {
		bayraklar = append(bayraklar, map[string]any{
			"names":       f.Adlar,
			"takesValue":  f.Degerli,
			"description": f.Aciklama,
		})
	}
	komutlar := make([]map[string]any, 0, len(Komutlar))
	for _, k := range Komutlar {
		komutlar = append(komutlar, map[string]any{
			"name":    k.Ad,
			"summary": k.Ozet,
			"example": k.Ornek,
		})
	}
	cikislar := make([]map[string]any, 0, len(CikisKodlari))
	for _, c := range CikisKodlari {
		cikislar = append(cikislar, map[string]any{
			"code":        c.Kod,
			"name":        c.Ad,
			"description": c.Aciklama,
		})
	}
	cozumleyiciler := CozumleyiciKatalogu()
	sort.Slice(cozumleyiciler, func(i, j int) bool {
		return cozumleyiciler[i]["id"].(string) < cozumleyiciler[j]["id"].(string)
	})
	return map[string]any{
		"tool":            "domainglass",
		"version":         surum,
		"description":     "domain.glass istihbarat aracı; insan ve AI ajanları için.",
		"commands":        komutlar,
		"flags":           bayraklar,
		"exitCodes":       cikislar,
		"resolvers":       cozumleyiciler,
		"feedNames":       AkisAdlari(),
		"jsonSchemaCmd":   "domainglass sema",
		"pipelineHint":    "Satır çıktısı için -subs, -ips, -related, -emails, -orgs kullanın.",
		"determinismNote": "JSON anahtarları İngilizcedir; hata günlükleri stderr, veri stdout üzerindedir.",
	}
}
