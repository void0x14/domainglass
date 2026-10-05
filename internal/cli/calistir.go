package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/void0x14/domainglass/internal/build"
	"github.com/void0x14/domainglass/internal/dg"
	"github.com/void0x14/domainglass/internal/dns"
	"github.com/void0x14/domainglass/internal/feeds"
	"github.com/void0x14/domainglass/internal/model"
	"github.com/void0x14/domainglass/internal/out"
	"github.com/void0x14/domainglass/internal/recon"
)

// Çıkış kodları sözleşmesi.
const (
	CikisBasarili   = 0
	CikisHata       = 1
	CikisBulunamadi = 2
	CikisHizSiniri  = 3
	CikisKullanim   = 4
)

// Argumanlar, çözümlenmiş komut satırı durumudur.
type Argumanlar struct {
	Komut    string
	Degerler []string

	JSON        bool
	NDJSON      bool
	Alan        string
	Subs        bool
	IPs         bool
	Related     bool
	Emails      bool
	Orgs        bool
	Cozumleyici string
	Timeout     int
	HizMs       int
	Renk        string
	Sessiz      bool
	Girintisiz  bool
}

type bayrakTanimi struct {
	adlar   []string
	degerli bool
	ata     func(a *Argumanlar, deger string) error
}

func bayrakTanimlari() []bayrakTanimi {
	return []bayrakTanimi{
		{adlar: []string{"json"}, ata: func(a *Argumanlar, _ string) error { a.JSON = true; return nil }},
		{adlar: []string{"ndjson"}, ata: func(a *Argumanlar, _ string) error { a.NDJSON = true; return nil }},
		{adlar: []string{"alan", "fields"}, degerli: true, ata: func(a *Argumanlar, v string) error { a.Alan = v; return nil }},
		{adlar: []string{"subs", "subdomains"}, ata: func(a *Argumanlar, _ string) error { a.Subs = true; return nil }},
		{adlar: []string{"ips"}, ata: func(a *Argumanlar, _ string) error { a.IPs = true; return nil }},
		{adlar: []string{"related"}, ata: func(a *Argumanlar, _ string) error { a.Related = true; return nil }},
		{adlar: []string{"emails"}, ata: func(a *Argumanlar, _ string) error { a.Emails = true; return nil }},
		{adlar: []string{"orgs", "organizations"}, ata: func(a *Argumanlar, _ string) error { a.Orgs = true; return nil }},
		{adlar: []string{"cozumleyici", "resolvers"}, degerli: true, ata: func(a *Argumanlar, v string) error { a.Cozumleyici = v; return nil }},
		{adlar: []string{"timeout"}, degerli: true, ata: func(a *Argumanlar, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				return fmt.Errorf("-timeout pozitif tam sayı olmalı: %q", v)
			}
			a.Timeout = n
			return nil
		}},
		{adlar: []string{"hiz", "rate"}, degerli: true, ata: func(a *Argumanlar, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return fmt.Errorf("-hiz negatif olmayan tam sayı olmalı: %q", v)
			}
			a.HizMs = n
			return nil
		}},
		{adlar: []string{"renk", "color"}, degerli: true, ata: func(a *Argumanlar, v string) error {
			switch v {
			case "auto", "always", "never":
				a.Renk = v
				return nil
			default:
				return fmt.Errorf("-renk auto|always|never olmalı: %q", v)
			}
		}},
		{adlar: []string{"sessiz", "quiet"}, ata: func(a *Argumanlar, _ string) error { a.Sessiz = true; return nil }},
		{adlar: []string{"girintisiz"}, ata: func(a *Argumanlar, _ string) error { a.Girintisiz = true; return nil }},
		{adlar: []string{"yardim", "help", "h"}, ata: func(a *Argumanlar, _ string) error { a.Komut = "yardim"; return nil }},
		{adlar: []string{"kimlik", "identity"}, ata: func(a *Argumanlar, _ string) error { a.Komut = "kimlik"; return nil }},
		{adlar: []string{"sema", "schema"}, ata: func(a *Argumanlar, _ string) error { a.Komut = "sema"; return nil }},
		{adlar: []string{"yetenek", "capabilities"}, ata: func(a *Argumanlar, _ string) error { a.Komut = "yetenek"; return nil }},
	}
}

// KomutMu, verilen sözcüğün bilinen bir alt komut olup olmadığını bildirir.
func KomutMu(s string) bool {
	for _, k := range Komutlar {
		if k.Ad == s {
			return true
		}
	}
	return false
}

// Cozumle, komut satırı argümanlarını çözümler.
func Cozumle(argv []string) (*Argumanlar, error) {
	a := &Argumanlar{Timeout: 20, HizMs: int(dg.VarsayilanAsgariAralik / time.Millisecond), Renk: "auto"}
	if v := strings.TrimSpace(os.Getenv("DOMAINGLASS_HIZ")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			a.HizMs = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("DOMAINGLASS_TIMEOUT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			a.Timeout = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("DOMAINGLASS_RENK")); v == "always" || v == "never" || v == "auto" {
		a.Renk = v
	}
	if v := strings.TrimSpace(os.Getenv("DOMAINGLASS_COZUMSYICI")); v != "" {
		a.Cozumleyici = v
	}
	tanimlar := bayrakTanimlari()
	bul := func(ad string) *bayrakTanimi {
		for i := range tanimlar {
			for _, x := range tanimlar[i].adlar {
				if x == ad {
					return &tanimlar[i]
				}
			}
		}
		return nil
	}

	i := 0
	for i < len(argv) {
		arg := argv[i]
		if arg == "--" {
			a.Degerler = append(a.Degerler, argv[i+1:]...)
			break
		}
		if len(arg) > 1 && arg[0] == '-' {
			ad := strings.TrimLeft(arg, "-")
			var deger string
			var degerVar bool
			if idx := strings.IndexByte(ad, '='); idx >= 0 {
				deger = ad[idx+1:]
				degerVar = true
				ad = ad[:idx]
			}
			t := bul(ad)
			if t == nil {
				return nil, fmt.Errorf("bilinmeyen bayrak: %s", arg)
			}
			if t.degerli && !degerVar {
				if i+1 >= len(argv) {
					return nil, fmt.Errorf("-%s bir değer bekliyor", ad)
				}
				i++
				deger = argv[i]
			}
			if err := t.ata(a, deger); err != nil {
				return nil, err
			}
			i++
			continue
		}
		a.Degerler = append(a.Degerler, arg)
		i++
	}

	if a.Komut == "" && len(a.Degerler) > 0 && KomutMu(a.Degerler[0]) {
		a.Komut = a.Degerler[0]
		a.Degerler = a.Degerler[1:]
	}
	return a, nil
}

// satirKipi, bayrakların satır satır boru hattı çıktısı isteyip istemediğini bildirir.
func (a *Argumanlar) satirKipi() bool {
	return a.Subs || a.IPs || a.Related || a.Emails || a.Orgs
}

// Calistir, süreci çalıştırır ve çıkış kodunu döner.
func Calistir(argv []string, stdout, stderr *os.File) int {
	a, err := Cozumle(argv)
	if err != nil {
		fmt.Fprintf(stderr, "domainglass: %v\n\n%s", err, Kullanim())
		return CikisKullanim
	}

	switch a.Komut {
	case "yardim":
		fmt.Fprint(stdout, Kullanim())
		return CikisBasarili
	case "kimlik":
		k := build.Oku()
		fmt.Fprintln(stdout, k.Tam())
		return CikisBasarili
	case "sema":
		fmt.Fprintln(stdout, Sema)
		return CikisBasarili
	case "yetenek":
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(YetenekKatalogu()); err != nil {
			fmt.Fprintf(stderr, "domainglass: katalog yazılamadı: %v\n", err)
			return CikisHata
		}
		return CikisBasarili
	}

	if a.JSON && a.NDJSON {
		fmt.Fprintf(stderr, "domainglass: -json ve -ndjson birlikte kullanılamaz\n")
		return CikisKullanim
	}
	if a.Alan != "" && !a.JSON && !a.NDJSON {
		fmt.Fprintf(stderr, "domainglass: -alan yalnızca -json veya -ndjson ile kullanılır\n")
		return CikisKullanim
	}
	if a.satirKipi() && (a.JSON || a.NDJSON) {
		fmt.Fprintf(stderr, "domainglass: satır kipi bayrakları (-subs/-ips/-related/-emails/-orgs) JSON ile birlikte kullanılamaz\n")
		return CikisKullanim
	}

	renk := out.Renk{Acik: renkAcikMi(a.Renk, stdout)}
	zamanAsimi := time.Duration(a.Timeout) * time.Second
	istemciSecenekleri := []dg.Secenek{
		dg.ZamanAsimi(zamanAsimi),
		dg.AsgariAralik(time.Duration(a.HizMs) * time.Millisecond),
	}
	if taban := strings.TrimSpace(os.Getenv("DOMAINGLASS_API")); taban != "" {
		istemciSecenekleri = append(istemciSecenekleri, dg.TabanURL(taban))
	}
	if ua := strings.TrimSpace(os.Getenv("DOMAINGLASS_UA")); ua != "" {
		istemciSecenekleri = append(istemciSecenekleri, dg.KullaniciAraci(ua))
	}
	istemci := dg.Yeni(istemciSecenekleri...)

	var cozumleyici *dns.Cozumleyici
	var cozumleyiciFiltre []string
	if a.Cozumleyici != "" {
		cozumleyici = dns.YeniCozumleyici(zamanAsimi)
		for _, id := range strings.Split(a.Cozumleyici, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := dns.SaglayiciGetir(id); !ok {
				fmt.Fprintf(stderr, "domainglass: bilinmeyen çözümleyici: %s\n", id)
				return CikisKullanim
			}
			cozumleyiciFiltre = append(cozumleyiciFiltre, id)
		}
	}

	secenekler := recon.Secenekler{
		Cozumleyici:       cozumleyici,
		CozumleyiciFiltre: cozumleyiciFiltre,
	}

	ctx, iptal := context.WithCancel(context.Background())
	defer iptal()

	switch a.Komut {
	case "saglik":
		if err := istemci.SaglikKontrolu(ctx); err != nil {
			fmt.Fprintf(stderr, "domainglass: servis erişilemiyor: %v\n", err)
			return cikisKoduHatasindan(err)
		}
		fmt.Fprintln(stdout, "domain.glass servisi erişilebilir.")
		return CikisBasarili

	case "akis":
		if len(a.Degerler) == 0 {
			fmt.Fprintf(stderr, "domainglass: akis komutu bir akış adı bekliyor (ör. top-gainers)\n")
			return CikisKullanim
		}
		return akisCalistir(ctx, istemci, a, a.Degerler[0], stdout, stderr, renk)

	case "toplu":
		return topluCalistir(ctx, istemci, a, secenekler, stdout, stderr, renk)

	case "asn":
		if len(a.Degerler) == 0 {
			fmt.Fprintf(stderr, "domainglass: asn komutu bir ASN numarası bekliyor (ör. 13335)\n")
			return CikisKullanim
		}
		return asnCalistir(ctx, istemci, a, a.Degerler[0], stdout, stderr, renk)

	case "tld":
		if len(a.Degerler) == 0 {
			fmt.Fprintf(stderr, "domainglass: tld komutu bir TLD bekliyor (ör. com)\n")
			return CikisKullanim
		}
		return tldCalistir(ctx, istemci, a, a.Degerler[0], stdout, stderr, renk)

	case "domain", "ip":
		if len(a.Degerler) == 0 {
			fmt.Fprintf(stderr, "domainglass: %s komutu bir hedef bekliyor\n", a.Komut)
			return CikisKullanim
		}
		return hedefCalistir(ctx, istemci, a, a.Degerler[0], secenekler, stdout, stderr, renk)

	case "":
		if len(a.Degerler) == 0 {
			fmt.Fprint(stderr, Kullanim())
			return CikisKullanim
		}
		if len(a.Degerler) > 1 {
			fmt.Fprintf(stderr, "domainglass: tek seferde tek hedef; toplu tarama için 'toplu' komutunu kullanın\n")
			return CikisKullanim
		}
		if !GecerliHedef(a.Degerler[0]) {
			fmt.Fprintf(stderr, "domainglass: %q ne bilinen bir komut ne de geçerli bir hedef.\n\n%s", a.Degerler[0], Kullanim())
			return CikisKullanim
		}
		return hedefCalistir(ctx, istemci, a, a.Degerler[0], secenekler, stdout, stderr, renk)

	default:
		fmt.Fprintf(stderr, "domainglass: bilinmeyen komut: %s\n", a.Komut)
		return CikisKullanim
	}
}

func hedefCalistir(ctx context.Context, c *dg.Client, a *Argumanlar, hedef string, s recon.Secenekler, stdout, stderr *os.File, renk out.Renk) int {
	hedef = strings.TrimSpace(hedef)
	if hedef == "" {
		fmt.Fprintf(stderr, "domainglass: boş hedef\n")
		return CikisKullanim
	}
	if !a.Sessiz {
		fmt.Fprintf(stderr, "domainglass: %s taranıyor…\n", hedef)
	}
	rapor := recon.Topla(ctx, c, hedef, s)
	rapor.Arac = build.AracBilgisi("domainglass")
	if err := ciktiVer(stdout, stderr, rapor, a, renk); err != nil {
		fmt.Fprintf(stderr, "domainglass: çıktı yazılamadı: %v\n", err)
		return CikisHata
	}
	return cikisKodu(rapor)
}

func akisCalistir(ctx context.Context, c *dg.Client, a *Argumanlar, ad string, stdout, stderr *os.File, renk out.Renk) int {
	ad = strings.ToLower(strings.TrimSpace(ad))
	if _, ok := feeds.Adlar[ad]; !ok {
		fmt.Fprintf(stderr, "domainglass: bilinmeyen akış: %s\n", ad)
		var adlar []string
		for k := range feeds.Adlar {
			adlar = append(adlar, k)
		}
		sortStrings(adlar)
		fmt.Fprintf(stderr, "geçerli akışlar: %s\n", strings.Join(adlar, ", "))
		return CikisKullanim
	}
	ham, err := c.Akis(ctx, ad)
	if err != nil {
		fmt.Fprintf(stderr, "domainglass: akış alınamadı: %v\n", err)
		return cikisKoduHatasindan(err)
	}
	bolum, err := feeds.Cozumle(ham)
	if err != nil {
		fmt.Fprintf(stderr, "domainglass: %v\n", err)
		return CikisHata
	}
	rapor := &model.Rapor{
		Hedef:     "akis:" + ad,
		Tur:       "feed",
		Arac:      build.AracBilgisi("domainglass"),
		Olusturma: time.Now().UTC().Format(time.RFC3339),
		Akis:      bolum,
		Kesif: model.KesifBolumu{
			AltAlanlar: []model.KesifOgesi{},
			Iliskili:   []model.KesifOgesi{},
			IPler:      []model.KesifOgesi{},
			Epostalar:  []model.KesifOgesi{},
			Telefonlar: []model.KesifOgesi{},
			Kurumlar:   []model.KesifOgesi{},
		},
		Kaynaklar: []model.KaynakDurumu{},
		Uyarilar:  []string{},
	}
	if a.Subs || a.Related {
		for _, h := range feeds.Hostlar(bolum) {
			rapor.Kesif.AltAlanlar = append(rapor.Kesif.AltAlanlar, model.KesifOgesi{Deger: h, Kaynaklar: []string{"RSS akışı"}})
		}
	}
	rapor.Ozet.AltAlanSayisi = len(rapor.Kesif.AltAlanlar)
	if err := ciktiVer(stdout, stderr, rapor, a, renk); err != nil {
		fmt.Fprintf(stderr, "domainglass: çıktı yazılamadı: %v\n", err)
		return CikisHata
	}
	return CikisBasarili
}

func asnCalistir(ctx context.Context, c *dg.Client, a *Argumanlar, girdi string, stdout, stderr *os.File, renk out.Renk) int {
	girdi = strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(girdi), "AS"))
	numara, err := strconv.Atoi(girdi)
	if err != nil || numara < 0 || numara > 4294967295 {
		fmt.Fprintf(stderr, "domainglass: geçersiz ASN: %s\n", girdi)
		return CikisKullanim
	}
	asn, err := c.ASN(ctx, numara)
	if err != nil {
		fmt.Fprintf(stderr, "domainglass: ASN alınamadı: %v\n", err)
		return cikisKoduHatasindan(err)
	}
	rapor := &model.Rapor{
		Hedef:     asn.Target.Label,
		Tur:       "asn",
		Arac:      build.AracBilgisi("domainglass"),
		Olusturma: time.Now().UTC().Format(time.RFC3339),
		ASN: &model.ASNBolumu{
			Numara:        asn.Target.Number,
			Etiket:        asn.Target.Label,
			Sahip:         asn.Overview.Holder,
			Duyurulmus:    asn.Overview.Announced,
			Blok:          asn.Overview.Block.Resource,
			PrefixSayisi:  asn.Prefixes.Total,
			Prefixler:     asn.Prefixes.Items,
			PrefixKesildi: asn.Prefixes.Truncated,
			YukariAkim:    asn.Neighbours.Counts.Upstream,
			AsagiAkim:     asn.Neighbours.Counts.Downstream,
			KayitDefteri:  asn.Registration.Source,
			KayitAdi:      asn.Registration.Name,
			KaynakURL:     asn.Registration.SourceURL,
		},
		Kesif: model.KesifBolumu{
			AltAlanlar: []model.KesifOgesi{},
			Iliskili:   []model.KesifOgesi{},
			IPler:      []model.KesifOgesi{},
			Epostalar:  []model.KesifOgesi{},
			Telefonlar: []model.KesifOgesi{},
			Kurumlar:   []model.KesifOgesi{},
		},
		Kaynaklar: []model.KaynakDurumu{},
		Uyarilar:  []string{},
	}
	for _, k := range asn.Neighbours.Items {
		rapor.ASN.Komsular = append(rapor.ASN.Komsular, model.ASNKomsu{Numara: k.Number, Iliski: k.Relationship, Guc: k.Power})
	}
	if a.Subs || a.IPs || a.Related {
		for _, p := range asn.Prefixes.Items {
			rapor.Kesif.IPler = append(rapor.Kesif.IPler, model.KesifOgesi{Deger: p, Kaynaklar: []string{"ASN prefix"}})
		}
		rapor.Ozet.IPAdresSayisi = len(rapor.Kesif.IPler)
	}
	if err := ciktiVer(stdout, stderr, rapor, a, renk); err != nil {
		fmt.Fprintf(stderr, "domainglass: çıktı yazılamadı: %v\n", err)
		return CikisHata
	}
	return CikisBasarili
}

func tldCalistir(ctx context.Context, c *dg.Client, a *Argumanlar, tld string, stdout, stderr *os.File, renk out.Renk) int {
	tld = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tld)), ".")
	if tld == "" {
		fmt.Fprintf(stderr, "domainglass: boş TLD\n")
		return CikisKullanim
	}
	g, err := c.TLDiana(ctx, tld)
	if err != nil {
		fmt.Fprintf(stderr, "domainglass: TLD kaydı alınamadı: %v\n", err)
		return cikisKoduHatasindan(err)
	}
	bolum := &model.TLDBolumu{TLD: g.Record.TLD, Adresler: g.Record.Addresses}
	for _, cnt := range g.Record.Contacts {
		bolum.Iletisimler = append(bolum.Iletisimler, model.TLDIletisim{
			Rol: cnt.Role, Ad: cnt.Name, Kurum: cnt.Organization,
			Telefon: cnt.Phone, Faks: cnt.Fax, Eposta: cnt.Email,
		})
	}
	for _, ns := range g.Record.Nameservers {
		bolum.AdSunuculari = append(bolum.AdSunuculari, model.TLDAdSunucusu{Adres: ns.Host, IPler: ns.Addresses})
	}
	rapor := &model.Rapor{
		Hedef:     "." + tld,
		Tur:       "tld",
		Arac:      build.AracBilgisi("domainglass"),
		Olusturma: time.Now().UTC().Format(time.RFC3339),
		TLD:       bolum,
		Kesif: model.KesifBolumu{
			AltAlanlar: []model.KesifOgesi{},
			Iliskili:   []model.KesifOgesi{},
			IPler:      []model.KesifOgesi{},
			Epostalar:  []model.KesifOgesi{},
			Telefonlar: []model.KesifOgesi{},
			Kurumlar:   []model.KesifOgesi{},
		},
		Kaynaklar: []model.KaynakDurumu{},
		Uyarilar:  []string{},
	}
	if a.IPs || a.Subs {
		for _, ns := range bolum.AdSunuculari {
			for _, ip := range ns.IPler {
				rapor.Kesif.IPler = append(rapor.Kesif.IPler, model.KesifOgesi{Deger: ip, Kaynaklar: []string{"IANA ad sunucusu"}})
			}
		}
		rapor.Ozet.IPAdresSayisi = len(rapor.Kesif.IPler)
	}
	if err := ciktiVer(stdout, stderr, rapor, a, renk); err != nil {
		fmt.Fprintf(stderr, "domainglass: çıktı yazılamadı: %v\n", err)
		return CikisHata
	}
	return CikisBasarili
}

func topluCalistir(ctx context.Context, c *dg.Client, a *Argumanlar, s recon.Secenekler, stdout, stderr *os.File, renk out.Renk) int {
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		fmt.Fprintf(stderr, "domainglass: toplu kip için stdin üzerinden hedef listesi bekleniyor\n")
		return CikisKullanim
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	enc := json.NewEncoder(stdout)
	enc.SetEscapeHTML(false)
	enKotu := CikisBasarili
	sayac := 0
	for scanner.Scan() {
		hedef := strings.TrimSpace(scanner.Text())
		if hedef == "" || strings.HasPrefix(hedef, "#") {
			continue
		}
		sayac++
		rapor := recon.Topla(ctx, c, hedef, s)
		rapor.Arac = build.AracBilgisi("domainglass")
		if a.NDJSON {
			if err := enc.Encode(rapor); err != nil {
				fmt.Fprintf(stderr, "domainglass: %s yazılamadı: %v\n", hedef, err)
				enKotu = CikisHata
				continue
			}
		} else if a.JSON {
			if err := out.JSON(stdout, rapor, !a.Girintisiz); err != nil {
				fmt.Fprintf(stderr, "domainglass: %s yazılamadı: %v\n", hedef, err)
				enKotu = CikisHata
				continue
			}
		} else if a.satirKipi() {
			satirlariYaz(stdout, rapor, a)
		} else {
			out.Yaz(stdout, rapor, renk)
		}
		kod := cikisKodu(rapor)
		if kod != CikisBasarili && enKotu == CikisBasarili {
			enKotu = kod
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(stderr, "domainglass: stdin okunamadı: %v\n", err)
		return CikisHata
	}
	if sayac == 0 {
		fmt.Fprintf(stderr, "domainglass: stdin üzerinde hedef bulunamadı\n")
		return CikisKullanim
	}
	return enKotu
}

func ciktiVer(stdout, stderr *os.File, rapor *model.Rapor, a *Argumanlar, renk out.Renk) error {
	switch {
	case a.satirKipi():
		satirlariYaz(stdout, rapor, a)
		return nil
	case a.NDJSON:
		if a.Alan != "" {
			secili, bilinmeyen, err := out.AlanSecimi(rapor, strings.Split(a.Alan, ","))
			if err != nil {
				return err
			}
			bilinmeyenAlanUyarisi(stderr, bilinmeyen)
			enc := json.NewEncoder(stdout)
			enc.SetEscapeHTML(false)
			return enc.Encode(secili)
		}
		return out.NDJSON(stdout, rapor)
	case a.JSON:
		if a.Alan != "" {
			secili, bilinmeyen, err := out.AlanSecimi(rapor, strings.Split(a.Alan, ","))
			if err != nil {
				return err
			}
			bilinmeyenAlanUyarisi(stderr, bilinmeyen)
			enc := json.NewEncoder(stdout)
			enc.SetEscapeHTML(false)
			if !a.Girintisiz {
				enc.SetIndent("", "  ")
			}
			return enc.Encode(secili)
		}
		return out.JSON(stdout, rapor, !a.Girintisiz)
	default:
		out.Yaz(stdout, rapor, renk)
		return nil
	}
}

// bilinmeyenAlanUyarisi, seçilen ancak raporda bulunmayan alanları stderr'e bildirir.
// Sessizce yok saymak, ajanların var olmayan veriyi bekleyip yanılmasına yol açardı.
func bilinmeyenAlanUyarisi(stderr *os.File, alanlar []string) {
	if len(alanlar) == 0 {
		return
	}
	fmt.Fprintf(stderr, "domainglass: uyarı: raporda bulunmayan alanlar yok sayıldı: %s"+"\n", strings.Join(alanlar, ", "))
	fmt.Fprintf(stderr, "geçerli alanlar: %s"+"\n", strings.Join(out.KullanilabilirAlanlar(), ", "))
}

func satirlariYaz(stdout *os.File, rapor *model.Rapor, a *Argumanlar) {
	yaz := func(liste []model.KesifOgesi) {
		for _, o := range liste {
			fmt.Fprintln(stdout, o.Deger)
		}
	}
	switch {
	case a.Subs:
		yaz(rapor.Kesif.AltAlanlar)
	case a.IPs:
		yaz(rapor.Kesif.IPler)
	case a.Related:
		yaz(rapor.Kesif.Iliskili)
	case a.Emails:
		yaz(rapor.Kesif.Epostalar)
	case a.Orgs:
		yaz(rapor.Kesif.Kurumlar)
	}
}

func cikisKodu(rapor *model.Rapor) int {
	kaynakVar := false
	hizSiniri := false
	for _, k := range rapor.Kaynaklar {
		switch k.Durum {
		case "ok", "empty":
			kaynakVar = true
		case "throttled":
			hizSiniri = true
		}
	}
	// Hiçbir kaynak veri döndürmediyse bu başarı değildir. Ajanların boş raporu
	// gerçek sonuç sanmasını engellemek için hata kodu döneriz.
	if !kaynakVar {
		if hizSiniri {
			return CikisHizSiniri
		}
		if len(rapor.Kaynaklar) > 0 {
			return CikisHata
		}
	}
	if rapor.Kayit != nil && rapor.Kayit.Durum == "not_found" {
		if rapor.DNS != nil && rapor.DNS.Durum == "NXDOMAIN" {
			return CikisBulunamadi
		}
		if rapor.Tur == "domain" {
			return CikisBulunamadi
		}
	}
	return CikisBasarili
}

func cikisKoduHatasindan(err error) int {
	if err == nil {
		return CikisBasarili
	}
	if dg.GeciciMi(err) {
		return CikisHizSiniri
	}
	if dg.YokMu(err) {
		return CikisBulunamadi
	}
	var h *dg.Hata
	if errors.As(err, &h) && h.Tur == dg.HataGecersiz {
		return CikisKullanim
	}
	return CikisHata
}

func renkAcikMi(kip string, stdout *os.File) bool {
	switch kip {
	case "always":
		return true
	case "never":
		return false
	default:
		fi, err := stdout.Stat()
		if err != nil {
			return false
		}
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
}

// CozumleyiciKatalogu, çözümleyici listesini makine kataloğu biçiminde döner.
func CozumleyiciKatalogu() []map[string]any {
	var out []map[string]any
	for _, s := range dns.Saglayicilar {
		out = append(out, map[string]any{
			"id":          s.ID,
			"name":        s.Ad,
			"address":     s.Adres,
			"description": s.Aciklama,
			"role":        s.FiltreRolu,
		})
	}
	return out
}

// AkisAdlari, desteklenen RSS akışlarının adlarını döner.
func AkisAdlari() []string {
	var out []string
	for ad := range feeds.Adlar {
		out = append(out, ad)
	}
	sortStrings(out)
	return out
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// GecerliHedef, bir sözcüğün alan adı veya IP adresi olarak kabul edilebilir
// olduğunu bildirir. Yazım hatalarını erken yakalamak için kullanılır.
func GecerliHedef(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if net.ParseIP(s) != nil {
		return true
	}
	if s == "localhost" {
		return true
	}
	if strings.ContainsAny(s, " 	/?#@") {
		return false
	}
	return strings.Contains(s, ".")
}
