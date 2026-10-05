// Package out, raporun insan ve makine biçimlerini üretir.
//
// Kural: JSON anahtarları İngilizcedir (makine sözleşmesi); insan çıktısı
// tamamen Türkçedir. Renk yalnızca uçbirim (TTY) algılandığında uygulanır.
package out

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/void0x14/domainglass/internal/model"
)

// Renk, ANSI renk kodlarını tutar.
type Renk struct {
	Acik bool
}

func (r Renk) kod(k string) string {
	if !r.Acik {
		return ""
	}
	return k
}

func (r Renk) Baslik(s string) string { return r.kod("\x1b[1;36m") + s + r.kod("\x1b[0m") }
func (r Renk) Etiket(s string) string { return r.kod("\x1b[1;34m") + s + r.kod("\x1b[0m") }
func (r Renk) Iyi(s string) string    { return r.kod("\x1b[32m") + s + r.kod("\x1b[0m") }
func (r Renk) Kotu(s string) string   { return r.kod("\x1b[31m") + s + r.kod("\x1b[0m") }
func (r Renk) Uyari(s string) string  { return r.kod("\x1b[33m") + s + r.kod("\x1b[0m") }
func (r Renk) Solgun(s string) string { return r.kod("\x1b[2m") + s + r.kod("\x1b[0m") }

// JSON, raporu okunabilir JSON olarak yazar.
func JSON(w io.Writer, rapor *model.Rapor, girinti bool) error {
	enc := json.NewEncoder(w)
	if girinti {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)
	return enc.Encode(rapor)
}

// NDJSON, raporu tek satır JSON olarak yazar (toplu akış için).
func NDJSON(w io.Writer, rapor *model.Rapor) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(rapor)
}

// AlanSecimi, JSON çıktısını istenen üst düzey alanlarla sınırlar.
// Boş seçim tüm raporu döner. Bilinmeyen alan adları hataya yazılır.
func AlanSecimi(rapor *model.Rapor, alanlar []string) (map[string]any, []string, error) {
	govde, err := json.Marshal(rapor)
	if err != nil {
		return nil, nil, err
	}
	var tum map[string]any
	if err := json.Unmarshal(govde, &tum); err != nil {
		return nil, nil, err
	}
	if len(alanlar) == 0 {
		return tum, nil, nil
	}
	secili := map[string]any{}
	var bilinmeyen []string
	for _, a := range alanlar {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		v, ok := tum[a]
		if !ok {
			bilinmeyen = append(bilinmeyen, a)
			continue
		}
		secili[a] = v
	}
	return secili, bilinmeyen, nil
}

// KullanilabilirAlanlar, seçilebilir üst düzey JSON alanlarını döner.
func KullanilabilirAlanlar() []string {
	return []string{
		"target", "kind", "tool", "generatedAt", "summary", "rankings", "registration",
		"dns", "tls", "web", "safety", "classifications", "ip", "asn", "tld", "feed",
		"resolvers", "discovery", "sources", "warnings",
	}
}

// Yaz, raporu insan okunur Türkçe biçimde yazar.
func Yaz(w io.Writer, rapor *model.Rapor, renk Renk) {
	baslik(w, rapor, renk)
	if rapor.Siralama != nil {
		siralamaYaz(w, rapor.Siralama, renk)
	}
	if rapor.Kayit != nil {
		kayitYaz(w, rapor.Kayit, renk)
	}
	if rapor.DNS != nil {
		dnsYaz(w, rapor.DNS, renk)
	}
	if rapor.TLS != nil {
		tlsYaz(w, rapor.TLS, renk)
	}
	if rapor.Web != nil {
		webYaz(w, rapor.Web, renk)
	}
	if rapor.Guvenlik != nil {
		guvenlikYaz(w, rapor.Guvenlik, renk)
	}
	if rapor.Siniflandirma != nil && rapor.Siniflandirma.EslesmeSayisi > 0 {
		siniflandirmaYaz(w, rapor.Siniflandirma, renk)
	}
	if rapor.IP != nil {
		ipYaz(w, rapor.IP, renk)
	}
	if rapor.ASN != nil {
		asnYaz(w, rapor.ASN, renk)
	}
	if rapor.TLD != nil {
		tldYaz(w, rapor.TLD, renk)
	}
	if rapor.Akis != nil {
		akisYaz(w, rapor.Akis, renk)
	}
	if len(rapor.Cozumleyiciler) > 0 {
		cozumleyiciYaz(w, rapor.Cozumleyiciler, renk)
	}
	kesifYaz(w, &rapor.Kesif, renk)
	if len(rapor.Uyarilar) > 0 {
		fmt.Fprintf(w, "\n%s\n", renk.Uyari("[!] Uyarılar"))
		for _, u := range rapor.Uyarilar {
			fmt.Fprintf(w, "    - %s\n", u)
		}
	}
	if hatali := basarisizKaynaklar(rapor.Kaynaklar); len(hatali) > 0 {
		fmt.Fprintf(w, "\n%s\n", renk.Uyari("[!] Ulaşılamayan kaynaklar"))
		for _, k := range hatali {
			fmt.Fprintf(w, "    - %s (%s): %s\n", k.Ad, k.Durum, k.Hata)
		}
	}
}

func basarisizKaynaklar(kaynaklar []model.KaynakDurumu) []model.KaynakDurumu {
	var out []model.KaynakDurumu
	for _, k := range kaynaklar {
		if k.Durum == "error" || k.Durum == "throttled" {
			out = append(out, k)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ad < out[j].Ad })
	return out
}

func baslik(w io.Writer, rapor *model.Rapor, renk Renk) {
	cizgi := strings.Repeat("─", 74)
	fmt.Fprintf(w, "%s\n", renk.Solgun(cizgi))
	tur := "Alan adı"
	switch rapor.Tur {
	case "ip":
		tur = "IP adresi"
	case "asn":
		tur = "ASN"
	case "tld":
		tur = "TLD"
	case "feed":
		tur = "Akış"
	}
	fmt.Fprintf(w, " %s  %s\n", renk.Baslik("DOMAINGLASS"), renk.Solgun(aracKimligi(rapor.Arac)))
	fmt.Fprintf(w, " %s: %s  (%s)\n", renk.Etiket("Hedef"), rapor.Hedef, tur)
	if rapor.Olusturma != "" {
		fmt.Fprintf(w, " %s: %s\n", renk.Etiket("Oluşturma"), rapor.Olusturma)
	}
	fmt.Fprintf(w, "%s\n", renk.Solgun(cizgi))
}

func bolum(w io.Writer, renk Renk, baslik string) {
	fmt.Fprintf(w, "\n%s\n", renk.Baslik("[+] "+baslik))
}

func satir(w io.Writer, renk Renk, etiket, deger string) {
	fmt.Fprintf(w, "    %s %s\n", renk.Etiket(fmt.Sprintf("%-16s", etiket+":")), deger)
}

func siralamaYaz(w io.Writer, s *model.SiralamaBolumu, renk Renk) {
	bolum(w, renk, "Popülerlik sıralamaları")
	satir(w, renk, "Cisco Umbrella", siralamaMetni(s.Cisco, renk))
	satir(w, renk, "Tranco", siralamaMetni(s.Tranco, renk))
}

func siralamaMetni(s *model.Siralama, renk Renk) string {
	if s == nil {
		return renk.Solgun("veri yok")
	}
	if !s.Listed || s.CurrentRank == nil {
		return renk.Solgun("ilk 1 milyon içinde değil")
	}
	metin := fmt.Sprintf("#%d", *s.CurrentRank)
	if s.Change != nil {
		switch {
		case *s.Change > 0:
			metin += " " + renk.Iyi(fmt.Sprintf("(▲ %d)", *s.Change))
		case *s.Change < 0:
			metin += " " + renk.Kotu(fmt.Sprintf("(▼ %d)", -*s.Change))
		default:
			metin += " (değişim yok)"
		}
	}
	if s.CurrentDate != "" {
		metin += " " + renk.Solgun("· "+s.CurrentDate)
	}
	return metin
}

func kayitYaz(w io.Writer, k *model.KayitBolumu, renk Renk) {
	bolum(w, renk, "Alan adı kaydı")
	durum := k.Durum
	switch k.Durum {
	case "registered":
		durum = renk.Iyi("kayıtlı")
	case "not_found":
		durum = renk.Kotu("kayıtlı değil")
	default:
		durum = renk.Solgun("bilinmiyor")
	}
	satir(w, renk, "Durum", durum)
	satir(w, renk, "Kaynak", k.Kaynak)
	if k.Kayitci != "" {
		kayitci := k.Kayitci
		if k.KayitciID != "" {
			kayitci += " (IANA ID: " + k.KayitciID + ")"
		}
		satir(w, renk, "Kayıtçı", kayitci)
	}
	if k.KayitTarihi != "" {
		satir(w, renk, "Kayıt tarihi", k.KayitTarihi)
	}
	if k.GuncellemeTarihi != "" {
		satir(w, renk, "Güncelleme", k.GuncellemeTarihi)
	}
	if k.BitisTarihi != "" {
		satir(w, renk, "Bitiş tarihi", k.BitisTarihi)
	}
	if len(k.Durumlar) > 0 {
		satir(w, renk, "Durumlar", strings.Join(k.Durumlar, ", "))
	}
	if len(k.AdSunuculari) > 0 {
		satir(w, renk, "Ad sunucuları", strings.Join(k.AdSunuculari, ", "))
	}
	if k.WHOISSunucusu != "" {
		satir(w, renk, "WHOIS sunucusu", k.WHOISSunucusu)
	}
}

func dnsYaz(w io.Writer, d *model.DNSBolumuRapor, renk Renk) {
	bolum(w, renk, "DNS kayıtları")
	satir(w, renk, "Durum", d.Durum)
	tipSirasi := []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "CAA", "SOA"}
	var yazilan int
	for _, tip := range tipSirasi {
		kayitlar, ok := d.Kayitlar[tip]
		if !ok {
			continue
		}
		for _, k := range kayitlar {
			fmt.Fprintf(w, "    %s %-6s %s\n", renk.Etiket(fmt.Sprintf("%-6s", tip)), renk.Solgun(fmt.Sprintf("TTL %d", k.TTL)), k.Deger)
			yazilan++
		}
	}
	if yazilan == 0 {
		fmt.Fprintf(w, "    %s\n", renk.Solgun("kayıt döndürülmedi"))
	}
}

func tlsYaz(w io.Writer, t *model.TLSBolumu, renk Renk) {
	bolum(w, renk, "TLS sertifikası")
	durum := t.Durum
	if t.Durum == "available" && t.Dogrulandi {
		durum = renk.Iyi("geçerli")
	} else if t.Durum == "available" {
		durum = renk.Uyari("geçerli değil")
	}
	satir(w, renk, "Durum", durum)
	if t.SNI != "" {
		satir(w, renk, "SNI", t.SNI)
	}
	if len(t.Konu.OrtakAdlar) > 0 {
		satir(w, renk, "Konu", strings.Join(t.Konu.OrtakAdlar, ", "))
	}
	if len(t.Yayimci.Kurumlar) > 0 {
		satir(w, renk, "Yayımcı", strings.Join(t.Yayimci.Kurumlar, ", "))
	}
	if t.GecerlilikBas != "" && t.GecerlilikSon != "" {
		gecerlilik := t.GecerlilikBas + " → " + t.GecerlilikSon
		if t.KalanGun != nil {
			gecerlilik += fmt.Sprintf(" (%d gün kaldı)", *t.KalanGun)
		}
		satir(w, renk, "Geçerlilik", gecerlilik)
	}
	if t.ImzaAlgoritmasi != "" || t.AnahtarTuru != "" {
		anahtar := t.AnahtarTuru
		if t.AnahtarBit > 0 {
			anahtar += fmt.Sprintf(" %d bit", t.AnahtarBit)
		}
		if t.ImzaAlgoritmasi != "" {
			anahtar += " · " + t.ImzaAlgoritmasi
		}
		satir(w, renk, "Anahtar", anahtar)
	}
	if t.ParmakIzi != "" {
		satir(w, renk, "SHA-256", t.ParmakIzi)
	}
	if len(t.Adlar) > 0 {
		var adlar []string
		for _, a := range t.Adlar {
			adlar = append(adlar, a.Ad)
		}
		if len(adlar) > 12 {
			satir(w, renk, fmt.Sprintf("SAN (%d)", len(adlar)), strings.Join(adlar[:12], ", ")+fmt.Sprintf(" … (+%d)", len(adlar)-12))
		} else {
			satir(w, renk, fmt.Sprintf("SAN (%d)", len(adlar)), strings.Join(adlar, ", "))
		}
	}
	if len(t.IPAdresleri) > 0 {
		satir(w, renk, "IP adresleri", strings.Join(t.IPAdresleri, ", "))
	}
	for _, g := range t.Gozlemler {
		renkli := g.Mesaj
		if g.Seviye == "error" || g.Seviye == "warning" {
			renkli = renk.Uyari(g.Mesaj)
		}
		fmt.Fprintf(w, "    %s %s\n", renk.Solgun("·"), renkli)
	}
}

func webYaz(w io.Writer, web *model.WebBolumu, renk Renk) {
	bolum(w, renk, "Ana sayfa profili")
	if web.DurumKodu > 0 {
		satir(w, renk, "HTTP", fmt.Sprintf("%d %s", web.DurumKodu, web.DurumMetni))
	}
	if web.SonURL != "" {
		satir(w, renk, "Son URL", web.SonURL)
	}
	if v, ok := web.Meta["title"]; ok && v != "" {
		satir(w, renk, "Başlık", v)
	}
	if v, ok := web.Basliklar["server"]; ok && v != "" {
		satir(w, renk, "Sunucu", v)
	}
	for _, k := range web.Kurumlar {
		ad := k.Ad
		if ad == "" {
			ad = k.Tur
		}
		satir(w, renk, "Kurum", ad)
	}
}

func guvenlikYaz(w io.Writer, g *model.GuvenlikBolumu, renk Renk) {
	bolum(w, renk, "DNS filtreleri (içerik güvenliği)")
	for _, k := range g.Kategoriler {
		durum := k.Durum
		switch k.Durum {
		case "flagged":
			durum = renk.Kotu("BAYRAKLI")
		case "not_flagged":
			durum = renk.Iyi("temiz")
		default:
			durum = renk.Solgun(k.Durum)
		}
		satir(w, renk, k.Etiket, durum)
	}
	if len(g.Saglayicilar) > 0 {
		fmt.Fprintf(w, "    %s\n", renk.Solgun("sağlayıcı bazında:"))
		for _, p := range g.Saglayicilar {
			var parcalar []string
			for _, kat := range []string{"malware", "adult", "advertising"} {
				if s, ok := p.Sinyaller[kat]; ok {
					parcalar = append(parcalar, kat+"="+s)
				}
			}
			fmt.Fprintf(w, "      - %-34s %s\n", p.Ad, strings.Join(parcalar, " "))
		}
	}
}

func siniflandirmaYaz(w io.Writer, s *model.SiniflandirmaBolumu, renk Renk) {
	bolum(w, renk, fmt.Sprintf("Liste sınıflandırması (%d eşleşme)", s.EslesmeSayisi))
	for _, m := range s.Eslesmeler {
		severity := m.Onem
		switch m.Onem {
		case "danger":
			severity = renk.Kotu(m.Onem)
		case "warning":
			severity = renk.Uyari(m.Onem)
		default:
			severity = renk.Solgun(m.Onem)
		}
		fmt.Fprintf(w, "    %s [%s] %s\n", renk.Etiket("•"), severity, m.Etiket)
		if m.Aciklama != "" {
			fmt.Fprintf(w, "      %s\n", m.Aciklama)
		}
		if m.EslesenDeger != "" {
			fmt.Fprintf(w, "      %s\n", renk.Solgun(m.Kapsam+": "+m.EslesenDeger))
		}
	}
}

func ipYaz(w io.Writer, ip *model.IPBolumu, renk Renk) {
	bolum(w, renk, "IP yönlendirme ve konum")
	satir(w, renk, "Adres", fmt.Sprintf("%s (IPv%d, kapsam: %s)", ip.Adres, ip.Surum, ip.Kapsam))
	satir(w, renk, "Duyurulmuş", fmt.Sprintf("%v", ip.Duyurulmus))
	if ip.Prefix != "" {
		satir(w, renk, "Prefix", ip.Prefix)
	}
	for _, asn := range ip.ASNs {
		satir(w, renk, "Origin ASN", fmt.Sprintf("AS%d — %s", asn.Numara, asn.Sahip))
	}
	if ip.Ulke != "" || ip.Sehir != "" {
		konum := strings.Trim(strings.Join([]string{ip.Sehir, ip.Ulke}, ", "), ", ")
		if ip.Enlem != nil && ip.Boylam != nil {
			konum += fmt.Sprintf(" (%.2f, %.2f)", *ip.Enlem, *ip.Boylam)
		}
		satir(w, renk, "Konum", konum)
	}
	if len(ip.TersDNS) > 0 {
		satir(w, renk, "Ters DNS", strings.Join(ip.TersDNS, ", "))
	}
	if ip.KayitDefteri != "" {
		satir(w, renk, "Kayıt defteri", fmt.Sprintf("%s · %s", ip.KayitDefteri, ip.TahsisAdi))
	}
	if ip.Aralik != "" {
		satir(w, renk, "Aralık", ip.Aralik)
	}
}

func asnYaz(w io.Writer, a *model.ASNBolumu, renk Renk) {
	bolum(w, renk, "ASN istihbaratı")
	satir(w, renk, "ASN", a.Etiket)
	if a.Sahip != "" {
		satir(w, renk, "Sahip", a.Sahip)
	}
	satir(w, renk, "Prefix sayısı", fmt.Sprintf("%d", a.PrefixSayisi))
	satir(w, renk, "Komşular", fmt.Sprintf("yukarı akım %d · aşağı akım %d", a.YukariAkim, a.AsagiAkim))
	if len(a.Komsular) > 0 {
		var liste []string
		for i, k := range a.Komsular {
			if i >= 10 {
				break
			}
			liste = append(liste, fmt.Sprintf("AS%d (%s)", k.Numara, k.Iliski))
		}
		satir(w, renk, "Öne çıkanlar", strings.Join(liste, ", "))
	}
}

func tldYaz(w io.Writer, t *model.TLDBolumu, renk Renk) {
	bolum(w, renk, "IANA TLD kaydı")
	satir(w, renk, "TLD", "."+t.TLD)
	if len(t.Adresler) > 0 {
		satir(w, renk, "Adres", strings.Join(t.Adresler, ", "))
	}
	for _, c := range t.Iletisimler {
		if c.Rol != "administrative" && c.Rol != "technical" {
			continue
		}
		parcalar := []string{}
		if c.Ad != "" {
			parcalar = append(parcalar, c.Ad)
		}
		if c.Kurum != "" {
			parcalar = append(parcalar, c.Kurum)
		}
		if c.Eposta != "" {
			parcalar = append(parcalar, c.Eposta)
		}
		satir(w, renk, c.Rol, strings.Join(parcalar, " · "))
	}
	if len(t.AdSunuculari) > 0 {
		var liste []string
		for i, ns := range t.AdSunuculari {
			if i >= 6 {
				liste = append(liste, fmt.Sprintf("… (+%d)", len(t.AdSunuculari)-6))
				break
			}
			liste = append(liste, ns.Adres)
		}
		satir(w, renk, "Ad sunucuları", strings.Join(liste, ", "))
	}
}

func akisYaz(w io.Writer, a *model.AkisBolumu, renk Renk) {
	bolum(w, renk, "Akış")
	if a.Baslik != "" {
		satir(w, renk, "Başlık", a.Baslik)
	}
	if a.Guncelleme != "" {
		satir(w, renk, "Güncelleme", a.Guncelleme)
	}
	satir(w, renk, "Öğe sayısı", fmt.Sprintf("%d", a.OgeSayisi))
	for i, o := range a.Ogeler {
		if i >= 25 {
			fmt.Fprintf(w, "    %s\n", renk.Solgun(fmt.Sprintf("… (+%d öğe daha)", a.OgeSayisi-25)))
			break
		}
		fmt.Fprintf(w, "    %s %s\n", renk.Solgun(fmt.Sprintf("%2d.", i+1)), o.Baslik)
	}
}

func cozumleyiciYaz(w io.Writer, sonuclar []model.CozumleyiciSonucu, renk Renk) {
	bolum(w, renk, "Bağımsız DNS çözümleyicileri")
	for _, s := range sonuclar {
		durum := s.Durum
		if s.Engelli {
			durum = renk.Kotu("ENGELLİ")
		} else if s.Durum == "ok" {
			durum = renk.Iyi("serbest")
		}
		rol := ""
		if s.Rol != "" {
			rol = renk.Solgun(" [" + s.Rol + "]")
		}
		fmt.Fprintf(w, "    %s %-46s %s%s\n", renk.Etiket("•"), s.Ad, durum, rol)
		if s.Hata != "" {
			fmt.Fprintf(w, "      %s\n", renk.Solgun(s.Hata))
		}
	}
}

func kesifYaz(w io.Writer, k *model.KesifBolumu, renk Renk) {
	if len(k.AltAlanlar) == 0 && len(k.Iliskili) == 0 && len(k.IPler) == 0 && len(k.Epostalar) == 0 && len(k.Telefonlar) == 0 {
		return
	}
	bolum(w, renk, "Keşfedilen varlıklar")
	yazListe := func(baslik string, liste []model.KesifOgesi, sinir int) {
		if len(liste) == 0 {
			return
		}
		fmt.Fprintf(w, "    %s %s\n", renk.Etiket(baslik+":"), renk.Solgun(fmt.Sprintf("(%d)", len(liste))))
		for i, o := range liste {
			if i >= sinir {
				fmt.Fprintf(w, "      %s\n", renk.Solgun(fmt.Sprintf("… (+%d daha)", len(liste)-sinir)))
				break
			}
			fmt.Fprintf(w, "      - %-46s %s\n", o.Deger, renk.Solgun(strings.Join(o.Kaynaklar, ", ")))
		}
	}
	yazListe("Alt alan adları", k.AltAlanlar, 40)
	yazListe("İlişkili alan adları", k.Iliskili, 20)
	yazListe("IP adresleri", k.IPler, 20)
	yazListe("E-postalar", k.Epostalar, 10)
	yazListe("Telefonlar", k.Telefonlar, 10)
	yazListe("Kurumlar", k.Kurumlar, 10)
}

// OzetSatirlari, rapor özetini kısa metin satırları olarak döner.
func OzetSatirlari(rapor *model.Rapor) []string {
	var out []string
	if rapor.Ozet.Kayitli != nil {
		durum := "bilinmiyor"
		if *rapor.Ozet.Kayitli {
			durum = "kayıtlı"
		} else {
			durum = "kayıtlı değil"
		}
		out = append(out, "kayıt durumu: "+durum)
	}
	if rapor.Ozet.CiscoSira != nil {
		out = append(out, fmt.Sprintf("Cisco sıra: #%d", *rapor.Ozet.CiscoSira))
	}
	if rapor.Ozet.TrancoSira != nil {
		out = append(out, fmt.Sprintf("Tranco sıra: #%d", *rapor.Ozet.TrancoSira))
	}
	if rapor.Ozet.DNSSECImzali != nil {
		if *rapor.Ozet.DNSSECImzali {
			out = append(out, "DNSSEC: imzalı")
		} else {
			out = append(out, "DNSSEC: imzasız")
		}
	}
	if rapor.Ozet.TLSDurumu != "" {
		out = append(out, "TLS: "+rapor.Ozet.TLSDurumu)
	}
	if len(rapor.Ozet.GuvenlikBayrak) > 0 {
		out = append(out, "güvenlik bayrağı: "+strings.Join(rapor.Ozet.GuvenlikBayrak, ", "))
	}
	return out
}

// aracKimligi, rapor başlığında gösterilecek araç kimliğini üretir.
// Sürüm numarası yoktur; ikiliyi ayırt eden şey commit hash'idir.
func aracKimligi(a model.AracBilgisi) string {
	if a.Commit == "" {
		return "bilinmeyen derleme"
	}
	kimlik := a.Commit
	if a.Degisti {
		kimlik += "+ (yerel değişikliklerle)"
	}
	return kimlik
}
