// Package recon, domain.glass verilerini toplar ve pasif keşif korelasyonunu üretir.
package recon

import (
	"net"
	"net/url"
	"sort"
	"strings"

	"github.com/void0x14/domainglass/internal/model"
)

// IPMi, değerin IP adresi olup olmadığını bildirir.
func IPMi(s string) bool {
	return net.ParseIP(strings.TrimSpace(s)) != nil
}

// HostTemizle, alan adı/URL/port içeren girdileri normalize eder.
func HostTemizle(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "*.")
	s = strings.TrimSuffix(s, ".")
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		if u, err := url.Parse(s); err == nil && u.Hostname() != "" {
			s = u.Hostname()
		}
	}
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	s = strings.Trim(s, "[]")
	return s
}

// kesifToplayici, keşif öğelerini kaynak etiketiyle toplar.
type kesifToplayici struct {
	altAlan    map[string]map[string]bool
	iliskili   map[string]map[string]bool
	ipler      map[string]map[string]bool
	epostalar  map[string]map[string]bool
	telefonlar map[string]map[string]bool
	kurumlar   map[string]map[string]bool
}

func yeniKesifToplayici() *kesifToplayici {
	return &kesifToplayici{
		altAlan:    map[string]map[string]bool{},
		iliskili:   map[string]map[string]bool{},
		ipler:      map[string]map[string]bool{},
		epostalar:  map[string]map[string]bool{},
		telefonlar: map[string]map[string]bool{},
		kurumlar:   map[string]map[string]bool{},
	}
}

func ekle(m map[string]map[string]bool, deger, kaynak string) {
	if deger == "" {
		return
	}
	if m[deger] == nil {
		m[deger] = map[string]bool{}
	}
	m[deger][kaynak] = true
}

func (k *kesifToplayici) host(deger, kaynak, kok string) {
	h := HostTemizle(deger)
	if h == "" || h == kok {
		return
	}
	if IPMi(h) {
		ekle(k.ipler, h, kaynak)
		return
	}
	if !strings.Contains(h, ".") {
		return
	}
	if strings.HasSuffix(h, "."+kok) {
		ekle(k.altAlan, h, kaynak)
		return
	}
	ekle(k.iliskili, h, kaynak)
}

// metindenHostlar, serbest metinden URL ve host adaylarını çıkarır.
func (k *kesifToplayici) metindenHostlar(metin, kaynak, kok string) {
	if metin == "" {
		return
	}
	for _, parca := range strings.FieldsFunc(metin, func(r rune) bool {
		return r == ' ' || r == ',' || r == ';' || r == '"' || r == '<' || r == '>' || r == '(' || r == ')'
	}) {
		parca = strings.Trim(parca, "\x60'\"")
		if parca == "" {
			continue
		}
		if strings.HasPrefix(parca, "http://") || strings.HasPrefix(parca, "https://") {
			if u, err := url.Parse(parca); err == nil {
				k.host(u.Hostname(), kaynak, kok)
			}
			continue
		}
		if strings.ContainsAny(parca, "/?#@") {
			continue
		}
		if strings.Contains(parca, ".") || IPMi(parca) {
			k.host(parca, kaynak, kok)
		}
	}
}

func (k *kesifToplayici) disari(m map[string]map[string]bool) []model.KesifOgesi {
	if len(m) == 0 {
		return []model.KesifOgesi{}
	}
	anahtarlar := make([]string, 0, len(m))
	for a := range m {
		anahtarlar = append(anahtarlar, a)
	}
	sort.Strings(anahtarlar)
	out := make([]model.KesifOgesi, 0, len(anahtarlar))
	for _, a := range anahtarlar {
		kaynaklar := make([]string, 0, len(m[a]))
		for s := range m[a] {
			kaynaklar = append(kaynaklar, s)
		}
		sort.Strings(kaynaklar)
		out = append(out, model.KesifOgesi{Deger: a, Kaynaklar: kaynaklar})
	}
	return out
}

// KesifGirdisi, korelasyon için gereken ham verilerdir.
type KesifGirdisi struct {
	Hedef   string
	Kok     string
	AlanAdi *model.AlanAdiBilgisi
	TLS     *model.TLSBilgisi
	Web     *model.WebProfili
	Arama   *model.AramaBilgisi
	IP      *model.IPBilgisi
	WHOIS   *model.WHOISBilgisi
}

// Kesif, verilen kaynaklardan alt alan adı, ilişkili alan adı, IP, e-posta,
// telefon ve kurum korelasyonunu üretir.
func Kesif(g KesifGirdisi) model.KesifBolumu {
	k := yeniKesifToplayici()
	kok := g.Kok
	if kok == "" {
		kok = HostTemizle(g.Hedef)
	}

	if d := g.AlanAdi; d != nil {
		for _, kayit := range d.DNS.TumKayitlar() {
			kaynak := "DNS " + kayit.Type
			for _, vp := range kayit.ValueParts {
				if vp.Target != "" {
					k.host(vp.Target, kaynak, kok)
				}
			}
			k.metindenHostlar(kayit.Value, kaynak, kok)
		}
		for _, ns := range d.RDAP.Nameservers {
			k.host(ns, "RDAP ad sunucusu", kok)
		}
		for _, nd := range d.RDAP.NameserverDetails {
			k.host(nd.Name, "RDAP ad sunucusu", kok)
			for _, ip := range nd.IPv4 {
				k.host(ip, "RDAP ad sunucusu", kok)
			}
			for _, ip := range nd.IPv6 {
				k.host(ip, "RDAP ad sunucusu", kok)
			}
		}
	}

	if t := g.TLS; t != nil && t.Certificate != nil {
		for _, ad := range t.Certificate.DNSNames {
			k.host(ad.Value, "TLS sertifikası", kok)
		}
		for _, ip := range t.Certificate.IPAddresses {
			k.host(ip, "TLS sertifikası", kok)
		}
	}

	if w := g.Web; w != nil {
		k.metindenHostlar(w.RequestedURL, "Ana sayfa URL", kok)
		k.metindenHostlar(w.FinalURL, "Ana sayfa URL", kok)
		for _, u := range w.DiscoveredURLs {
			k.metindenHostlar(u, "HTML ve yönlendirmeler", kok)
		}
		for ad, deger := range w.Headers {
			k.metindenHostlar(deger, "HTTP "+ad, kok)
		}
		for _, deger := range w.Metadata {
			k.metindenHostlar(deger, "Sayfa meta verisi", kok)
		}
		kurumAdlari, epostalar, telefonlar := webIletisim(w)
		for _, e := range epostalar {
			ekle(k.epostalar, e, "Ana sayfa iletişim")
		}
		for _, t := range telefonlar {
			ekle(k.telefonlar, t, "Ana sayfa iletişim")
		}
		for _, ku := range kurumAdlari {
			ekle(k.kurumlar, ku, "Ana sayfa JSON-LD")
		}
	}

	if a := g.Arama; a != nil {
		for _, r := range a.Results {
			k.host(r.Host, "Arama sonuçları", kok)
			k.metindenHostlar(r.URL, "Arama sonuçları", kok)
		}
	}

	if ip := g.IP; ip != nil {
		for _, ad := range ip.ReverseDNS.Names {
			k.host(ad, "Ters DNS", kok)
		}
	}

	if who := g.WHOIS; who != nil {
		for _, ns := range who.Nameservers {
			k.host(ns, "WHOIS ad sunucusu", kok)
		}
		for _, nd := range who.NameserverDetails {
			k.host(nd.Name, "WHOIS ad sunucusu", kok)
		}
	}

	return model.KesifBolumu{
		AltAlanlar: k.disari(k.altAlan),
		Iliskili:   k.disari(k.iliskili),
		IPler:      k.disari(k.ipler),
		Epostalar:  k.disari(k.epostalar),
		Telefonlar: k.disari(k.telefonlar),
		Kurumlar:   k.disari(k.kurumlar),
	}
}

// webIletisim, web profilinden kurum adlarını, e-postaları ve telefonları çıkarır.
func webIletisim(w *model.WebProfili) (kurumlar, epostalar, telefonlar []string) {
	gorulenEposta := map[string]bool{}
	gorulenTel := map[string]bool{}
	for _, org := range w.Organizations {
		if ad, ok := org["name"].(string); ok && ad != "" {
			kurumlar = append(kurumlar, ad)
		}
		if em, ok := org["emails"].([]any); ok {
			for _, e := range em {
				if s, ok := e.(string); ok && s != "" && !gorulenEposta[s] {
					gorulenEposta[s] = true
					epostalar = append(epostalar, s)
				}
			}
		}
		if ph, ok := org["phones"].([]any); ok {
			for _, p := range ph {
				if s, ok := p.(string); ok && s != "" && !gorulenTel[s] {
					gorulenTel[s] = true
					telefonlar = append(telefonlar, s)
				}
			}
		}
	}
	if w.Contacts != nil {
		if em, ok := w.Contacts["emails"].([]any); ok {
			for _, e := range em {
				if s, ok := e.(string); ok && s != "" && !gorulenEposta[s] {
					gorulenEposta[s] = true
					epostalar = append(epostalar, s)
				}
			}
		}
		if ph, ok := w.Contacts["phones"].([]any); ok {
			for _, p := range ph {
				if s, ok := p.(string); ok && s != "" && !gorulenTel[s] {
					gorulenTel[s] = true
					telefonlar = append(telefonlar, s)
				}
			}
		}
	}
	sort.Strings(kurumlar)
	sort.Strings(epostalar)
	sort.Strings(telefonlar)
	return kurumlar, epostalar, telefonlar
}
