package recon

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/void0x14/domainglass/internal/build"
	"github.com/void0x14/domainglass/internal/dg"
	"github.com/void0x14/domainglass/internal/dns"
	"github.com/void0x14/domainglass/internal/model"
)

// Secenekler, toplama davranışını belirler.
type Secenekler struct {
	Cozumleyici       *dns.Cozumleyici
	CozumleyiciID     string
	CozumleyiciFiltre []string
	WHOISZorla        bool
	AkisAdi           string
	KaynakSecimi      []string
	Derin             bool
	EsZamanli         int
}

// Topla, hedef türüne göre ilgili tüm kaynakları paralel toplar.
func Topla(ctx context.Context, c *dg.Client, hedef string, s Secenekler) *model.Rapor {
	rapor := &model.Rapor{
		Hedef:     hedef,
		Arac:      build.AracBilgisi("domainglass"),
		Olusturma: time.Now().UTC().Format(time.RFC3339),
		Uyarilar:  []string{},
		Kaynaklar: []model.KaynakDurumu{},
	}
	if IPMi(hedef) {
		rapor.Tur = "ip"
		ipTopla(ctx, c, hedef, rapor, s)
	} else {
		rapor.Tur = "domain"
		alanAdiTopla(ctx, c, hedef, rapor, s)
	}
	rapor.Kesif = Kesif(KesifGirdisi{
		Hedef:   hedef,
		Kok:     strings.ToLower(strings.TrimSpace(hedef)),
		AlanAdi: DNSKaynagi(rapor),
		TLS:     TLSKaynagi(rapor),
		Web:     WebKaynagi(rapor),
		Arama:   AramaKaynagi(rapor),
		IP:      IPKaynagi(rapor),
		WHOIS:   WHOISKaynagi(rapor),
	})
	rapor.Ozet.AltAlanSayisi = len(rapor.Kesif.AltAlanlar)
	rapor.Ozet.IliskiliSayisi = len(rapor.Kesif.Iliskili)
	rapor.Ozet.IPAdresSayisi = len(rapor.Kesif.IPler)
	return rapor
}

// isDurumu, tek bir kaynağın sonucunu rapora işler.
func isDurumu(rapor *model.Rapor, id, ad string, baslangic time.Time, err error) {
	d := model.KaynakDurumu{ID: id, Ad: ad, GecikmeMs: time.Since(baslangic).Milliseconds()}
	switch {
	case err == nil:
		d.Durum = "ok"
	case dg.YokMu(err):
		d.Durum = "empty"
	case dg.GeciciMi(err):
		d.Durum = "throttled"
		d.Hata = err.Error()
	default:
		d.Durum = "error"
		d.Hata = err.Error()
	}
	rapor.Kaynaklar = append(rapor.Kaynaklar, d)
}

func alanAdiTopla(ctx context.Context, c *dg.Client, hedef string, rapor *model.Rapor, s Secenekler) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	calistir := func(id, ad string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			baslangic := time.Now()
			err := fn()
			mu.Lock()
			isDurumu(rapor, id, ad, baslangic, err)
			mu.Unlock()
		}()
	}

	calistir("domain", "DNS, RDAP ve içerik güvenliği", func() error {
		d, err := c.AlanAdi(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamAlanAdi = d
		mu.Unlock()
		return nil
	})

	calistir("cisco", "Cisco Umbrella sıralaması", func() error {
		sir, err := c.Siralama(ctx, hedef, "cisco")
		if err != nil {
			return err
		}
		mu.Lock()
		if rapor.Siralama == nil {
			rapor.Siralama = &model.SiralamaBolumu{}
		}
		rapor.Siralama.Cisco = sir
		mu.Unlock()
		return nil
	})

	calistir("tranco", "Tranco sıralaması", func() error {
		sir, err := c.Siralama(ctx, hedef, "tranco")
		if err != nil {
			return err
		}
		mu.Lock()
		if rapor.Siralama == nil {
			rapor.Siralama = &model.SiralamaBolumu{}
		}
		rapor.Siralama.Tranco = sir
		mu.Unlock()
		return nil
	})

	calistir("safety", "DNS filtre sağlayıcıları", func() error {
		g, err := c.Guvenlik(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamGuvenlik = g
		mu.Unlock()
		return nil
	})

	calistir("classifications", "HaGeZi liste sınıflandırması", func() error {
		sf, err := c.Siniflandirma(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamSiniflandirma = sf
		mu.Unlock()
		return nil
	})

	calistir("tls", "TLS sertifikası (port 443)", func() error {
		t, err := c.TLS(ctx, "domain", hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamTLS = t
		mu.Unlock()
		return nil
	})

	calistir("web-profile", "Ana sayfa profili", func() error {
		w, err := c.WebProfili(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamWeb = w
		mu.Unlock()
		return nil
	})

	calistir("search", "Web arama sonuçları", func() error {
		a, err := c.Arama(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamArama = a
		mu.Unlock()
		return nil
	})

	calistir("whois", "WHOIS kaydı", func() error {
		w, err := c.WHOIS(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamWHOIS = w
		mu.Unlock()
		return nil
	})

	calistir("tld-iana", "IANA TLD kaydı", func() error {
		tld := hedef
		if i := strings.LastIndex(hedef, "."); i >= 0 && i < len(hedef)-1 {
			tld = hedef[i+1:]
		}
		g, err := c.TLDiana(ctx, tld)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamTLD = g
		mu.Unlock()
		return nil
	})

	wg.Wait()
	isle(rapor)
	cozumleyiciTopla(ctx, rapor, s)
}

func ipTopla(ctx context.Context, c *dg.Client, hedef string, rapor *model.Rapor, s Secenekler) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	calistir := func(id, ad string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			baslangic := time.Now()
			err := fn()
			mu.Lock()
			isDurumu(rapor, id, ad, baslangic, err)
			mu.Unlock()
		}()
	}

	calistir("ip", "IP yönlendirme ve konum", func() error {
		i, err := c.IP(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamIP = i
		mu.Unlock()
		return nil
	})

	calistir("tls-ip", "TLS sertifikası (port 443)", func() error {
		t, err := c.TLS(ctx, "ip", hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamTLS = t
		mu.Unlock()
		return nil
	})

	calistir("classifications", "HaGeZi liste sınıflandırması", func() error {
		sf, err := c.Siniflandirma(ctx, hedef)
		if err != nil {
			return err
		}
		mu.Lock()
		rapor.HamSiniflandirma = sf
		mu.Unlock()
		return nil
	})

	wg.Wait()
	if rapor.HamIP != nil {
		for _, asn := range rapor.HamIP.Routing.ASNs {
			numara := asn.Number
			calistir("asn-"+itoa(numara), "ASN "+itoa(numara)+" komşuluk verisi", func() error {
				a, err := c.ASN(ctx, numara)
				if err != nil {
					return err
				}
				mu.Lock()
				if rapor.HamASNlar == nil {
					rapor.HamASNlar = map[int]*model.ASNBilgisi{}
				}
				rapor.HamASNlar[numara] = a
				mu.Unlock()
				return nil
			})
		}
		wg.Wait()
	}
	isle(rapor)
	cozumleyiciTopla(ctx, rapor, s)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negatif := n < 0
	if negatif {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if negatif {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// cozumleyiciTopla, istenen DNS filtre çözümleyicilerini çalıştırır.
func cozumleyiciTopla(ctx context.Context, rapor *model.Rapor, s Secenekler) {
	if s.Cozumleyici == nil || len(s.CozumleyiciFiltre) == 0 {
		return
	}
	hedef := rapor.Hedef
	var sonuclar []model.CozumleyiciSonucu
	for _, id := range s.CozumleyiciFiltre {
		sag, ok := dns.SaglayiciGetir(id)
		if !ok {
			continue
		}
		kayitlar := map[string][]dns.Record{}
		var err error
		if IPMi(hedef) {
			resp, e := s.Cozumleyici.Cozumle(ctx, id, hedef, dns.TypePTR)
			err = e
			if e == nil {
				kayitlar["PTR"] = resp.Records
			}
		} else {
			kayitlar, err = s.Cozumleyici.CozumleTum(ctx, id, hedef)
		}
		sonuc := model.CozumleyiciSonucu{ID: id, Ad: sag.Ad, Rol: sag.FiltreRolu}
		if err != nil {
			sonuc.Durum = "error"
			sonuc.Hata = err.Error()
		} else {
			sonuc.Durum = "ok"
			var degerler []string
			for _, tip := range dns.AllTypes {
				for _, r := range kayitlar[dns.TypeName(tip)] {
					degerler = append(degerler, dns.TypeName(tip)+" "+r.Value)
					if dns.Engelli(&dns.Response{Records: []dns.Record{r}}) {
						sonuc.Engelli = true
					}
				}
			}
			sort.Strings(degerler)
			sonuc.Kayitlar = degerler
		}
		sonuclar = append(sonuclar, sonuc)
	}
	rapor.Cozumleyiciler = sonuclar
}
