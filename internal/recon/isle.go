package recon

import (
	"strings"

	"github.com/void0x14/domainglass/internal/model"
)

// Bu dosya, ham API yanıtlarını kararlı rapor şemasına dönüştürür.

// isle, ham yanıtları rapor bölümlerine çevirir.
func isle(r *model.Rapor) {
	if r.HamAlanAdi != nil {
		d := r.HamAlanAdi
		r.DNS = &model.DNSBolumuRapor{
			Durum:    d.DNS.Status,
			Rcode:    d.DNS.Rcode,
			Kayitlar: map[string][]model.DNSKaydiRapor{},
		}
		ekle := func(tip string, kayitlar []model.DNSRecord) {
			if len(kayitlar) == 0 {
				return
			}
			liste := make([]model.DNSKaydiRapor, 0, len(kayitlar))
			for _, k := range kayitlar {
				liste = append(liste, model.DNSKaydiRapor{Ad: k.Name, TTL: k.TTL, Deger: k.Value})
			}
			r.DNS.Kayitlar[tip] = liste
		}
		ekle("A", d.DNS.Records.A)
		ekle("AAAA", d.DNS.Records.AAAA)
		ekle("CNAME", d.DNS.Records.CNAME)
		ekle("MX", d.DNS.Records.MX)
		ekle("NS", d.DNS.Records.NS)
		ekle("TXT", d.DNS.Records.TXT)
		ekle("CAA", d.DNS.Records.CAA)
		ekle("SOA", d.DNS.Records.SOA)

		kayit := &model.KayitBolumu{
			Durum:         "unknown",
			Kaynak:        "rdap",
			Kayitci:       d.RDAP.Registrar.Name,
			KayitciID:     d.RDAP.Registrar.ID,
			KayitciURL:    d.RDAP.Registrar.URL,
			Durumlar:      d.RDAP.Statuses,
			AdSunuculari:  d.RDAP.Nameservers,
			KayitDefteri:  d.RDAP.RegistryHandle,
			WHOISSunucusu: d.RDAP.Port43,
			RDAPKaynagi:   d.RDAP.SourceURL,
		}
		switch d.RDAP.RegistrationStatus {
		case "registered":
			kayit.Durum = "registered"
		case "not_found":
			kayit.Durum = "not_found"
		default:
			if d.DNS.RegistrationInference.Status == "registered" {
				kayit.Durum = "registered"
			}
		}
		for _, ev := range d.RDAP.Events {
			switch ev.Action {
			case "registration":
				kayit.KayitTarihi = ev.Date
			case "last changed", "last update of rdap database":
				if kayit.GuncellemeTarihi == "" {
					kayit.GuncellemeTarihi = ev.Date
				}
			case "expiration":
				kayit.BitisTarihi = ev.Date
			}
		}
		r.Kayit = kayit
	}

	if r.HamWHOIS != nil {
		w := r.HamWHOIS
		if r.Kayit == nil {
			r.Kayit = &model.KayitBolumu{Durum: "unknown"}
		}
		r.Kayit.Kaynak = "whois"
		if r.HamAlanAdi != nil {
			r.Kayit.Kaynak = "rdap+whois"
		}
		if w.Registered != nil {
			if *w.Registered {
				r.Kayit.Durum = "registered"
			} else {
				r.Kayit.Durum = "not_found"
			}
		}
		if w.Registrar.Name != "" {
			r.Kayit.Kayitci = w.Registrar.Name
		}
		if w.Registrar.ID != "" {
			r.Kayit.KayitciID = w.Registrar.ID
		}
		if w.Registrar.URL != "" {
			r.Kayit.KayitciURL = w.Registrar.URL
		}
		if w.Dates.Created != "" {
			r.Kayit.KayitTarihi = w.Dates.Created
		}
		if w.Dates.Updated != "" {
			r.Kayit.GuncellemeTarihi = w.Dates.Updated
		}
		if w.Dates.Expires != "" {
			r.Kayit.BitisTarihi = w.Dates.Expires
		}
		if len(w.Statuses) > 0 {
			r.Kayit.Durumlar = w.Statuses
		}
		if len(w.Nameservers) > 0 {
			r.Kayit.AdSunuculari = w.Nameservers
		}
		if w.RegistryHandle != "" {
			r.Kayit.KayitDefteri = w.RegistryHandle
		}
		if w.WHOISServer != "" {
			r.Kayit.WHOISSunucusu = w.WHOISServer
		}
	}

	if r.HamTLS != nil {
		t := r.HamTLS
		tlsBolum := &model.TLSBolumu{
			Durum:        t.Status,
			Dogrulandi:   t.Authorized,
			SNI:          t.SNI,
			SeffaflikURL: t.CertificateTransparencyURL,
		}
		if t.Certificate != nil {
			c := t.Certificate
			tlsBolum.Konu = kimlik(c.Subject.CommonNames, c.Subject.Organizations, c.Subject.OrganizationalUnits, c.Subject.Localities, c.Subject.States, c.Subject.Countries)
			tlsBolum.Yayimci = kimlik(c.Issuer.CommonNames, c.Issuer.Organizations, c.Issuer.OrganizationalUnits, c.Issuer.Localities, c.Issuer.States, c.Issuer.Countries)
			tlsBolum.GecerlilikBas = c.ValidFrom
			tlsBolum.GecerlilikSon = c.ValidTo
			tlsBolum.KalanGun = c.DaysRemaining
			tlsBolum.ImzaAlgoritmasi = c.SignatureAlgorithm
			tlsBolum.AnahtarTuru = c.KeyType
			tlsBolum.AnahtarBit = c.KeyBits
			tlsBolum.SeriNumarasi = c.SerialNumber
			tlsBolum.ParmakIzi = c.FingerprintSHA256
			tlsBolum.IPAdresleri = c.IPAddresses
			for _, ad := range c.DNSNames {
				tlsBolum.Adlar = append(tlsBolum.Adlar, model.TLSAdi{Ad: ad.Value, Joker: ad.Wildcard, Iliski: ad.Relationship})
			}
		}
		for _, o := range t.Observations {
			tlsBolum.Gozlemler = append(tlsBolum.Gozlemler, model.Gozlem{Seviye: o.Level, Kod: o.Code, Mesaj: o.Message})
		}
		r.TLS = tlsBolum
	}

	if r.HamWeb != nil {
		w := r.HamWeb
		r.Web = &model.WebBolumu{
			SonURL:     w.FinalURL,
			Protokol:   w.Protocol,
			DurumKodu:  w.Status,
			DurumMetni: w.StatusText,
			IcerikTuru: w.ContentType,
			Basliklar:  w.Headers,
			Meta:       w.Metadata,
			BulunanURL: w.DiscoveredURLs,
			Uyarilar:   w.Warnings,
		}
		for _, org := range w.Organizations {
			kb := model.KurumBilgisi{}
			if v, ok := org["name"].(string); ok {
				kb.Ad = v
			}
			if v, ok := org["type"].(string); ok {
				kb.Tur = v
			}
			if v, ok := org["url"].(string); ok {
				kb.Website = v
			}
			if v, ok := org["emails"].([]any); ok {
				for _, e := range v {
					if s, ok := e.(string); ok {
						kb.Epostalar = append(kb.Epostalar, s)
					}
				}
			}
			if v, ok := org["phones"].([]any); ok {
				for _, p := range v {
					if s, ok := p.(string); ok {
						kb.Telefonlar = append(kb.Telefonlar, s)
					}
				}
			}
			r.Web.Kurumlar = append(r.Web.Kurumlar, kb)
		}
	}

	if r.HamGuvenlik != nil {
		g := r.HamGuvenlik
		bolum := &model.GuvenlikBolumu{}
		kategoriler := []struct {
			ID     string
			Etiket string
			K      model.GuvenlikKategorisi
		}{
			{"malware", "Zararlı yazılım ve oltalama", g.Categories.Malware},
			{"adult", "Yetişkin ve aile dışı içerik", g.Categories.Adult},
			{"advertising", "Reklam, analitik ve izleme", g.Categories.Advertising},
		}
		for _, k := range kategoriler {
			kr := model.GuvenlikKategorisiRapor{
				Kategori:    k.ID,
				Etiket:      k.Etiket,
				Durum:       k.K.State,
				Bayraklayan: k.K.FlaggedBy,
				KontrolEden: k.K.CheckedBy,
				Aciklama:    k.K.Explanation,
			}
			if k.K.State == "flagged" {
				bolum.Bayraklandi = true
			}
			bolum.Kategoriler = append(bolum.Kategoriler, kr)
		}
		for _, p := range g.Providers {
			sg := model.SaglayiciGuvenlik{
				ID:        p.ID,
				Ad:        p.Name,
				Durum:     p.Status,
				BelgeURL:  p.DocumentationURL,
				Sinyaller: map[string]string{},
			}
			for ad, sinyal := range p.Signals {
				sg.Sinyaller[ad] = sinyal.State
			}
			bolum.Saglayicilar = append(bolum.Saglayicilar, sg)
		}
		r.Guvenlik = bolum
	}

	if r.HamSiniflandirma != nil {
		sf := r.HamSiniflandirma
		bolum := &model.SiniflandirmaBolumu{
			EslesmeSayisi: len(sf.Matches),
			VeriTarihi:    sf.SourceDataAt,
			Surum:         sf.ReleaseID,
			Kaynak:        sf.Source.Name,
		}
		for _, m := range sf.Matches {
			bolum.Eslesmeler = append(bolum.Eslesmeler, model.SiniflandirmaEslesme{
				Kategori:       m.Category,
				Etiket:         m.Label,
				Aciklama:       m.Description,
				Onem:           m.Severity,
				EslesenDeger:   m.MatchedValue,
				Kapsam:         m.Scope,
				KaynakListeler: m.SourceFeeds,
			})
		}
		r.Siniflandirma = bolum
	}

	if r.HamIP != nil {
		i := r.HamIP
		bolum := &model.IPBolumu{
			Adres:        i.Target.Address,
			Surum:        i.Target.Version,
			Kapsam:       i.Target.Scope,
			GenelMi:      i.Target.IsPublic,
			TersDNS:      i.ReverseDNS.Names,
			TersDNSSorgu: i.ReverseDNS.QueryName,
			Duyurulmus:   i.Routing.Announced,
			Prefix:       i.Routing.Prefix,
			BGPGozlem:    i.Routing.ObservedAt,
			Ulke:         i.Location.CountryCode,
			Sehir:        i.Location.City,
			KonumKapsama: i.Location.CoveredPercentage,
			KayitDefteri: i.Allocation.Registry,
			TahsisAdi:    i.Allocation.Name,
			TahsisTuru:   i.Allocation.Type,
			CIDRler:      i.Allocation.CIDRs,
			RDAPKaynagi:  i.Allocation.SourceURL,
		}
		if i.ReverseDNS.QueryName != "" {
			ileri := i.ReverseDNS.ForwardConfirmed
			bolum.IleriDogrulandi = &ileri
		}
		if i.Location.Latitude != 0 || i.Location.Longitude != 0 {
			enlem := i.Location.Latitude
			boylam := i.Location.Longitude
			bolum.Enlem = &enlem
			bolum.Boylam = &boylam
		}
		if i.Allocation.StartAddress != "" {
			bolum.Aralik = i.Allocation.StartAddress + " - " + i.Allocation.EndAddress
		}
		for _, asn := range i.Routing.ASNs {
			bolum.ASNs = append(bolum.ASNs, model.ASNOzet{Numara: asn.Number, Sahip: asn.Holder})
		}
		r.IP = bolum
	}

	if r.HamASNlar != nil {
		var numaralar []int
		for n := range r.HamASNlar {
			numaralar = append(numaralar, n)
		}
		if len(numaralar) > 0 {
			ilk := r.HamASNlar[numaralar[0]]
			bolum := &model.ASNBolumu{
				Numara:        ilk.Target.Number,
				Etiket:        ilk.Target.Label,
				Sahip:         ilk.Overview.Holder,
				Duyurulmus:    ilk.Overview.Announced,
				Blok:          ilk.Overview.Block.Resource,
				PrefixSayisi:  ilk.Prefixes.Total,
				Prefixler:     ilk.Prefixes.Items,
				PrefixKesildi: ilk.Prefixes.Truncated,
				YukariAkim:    ilk.Neighbours.Counts.Upstream,
				AsagiAkim:     ilk.Neighbours.Counts.Downstream,
				KayitDefteri:  ilk.Registration.Source,
				KayitAdi:      ilk.Registration.Name,
				KaynakURL:     ilk.Registration.SourceURL,
			}
			for _, k := range ilk.Neighbours.Items {
				bolum.Komsular = append(bolum.Komsular, model.ASNKomsu{Numara: k.Number, Iliski: k.Relationship, Guc: k.Power})
			}
			r.ASN = bolum
		}
	}

	if r.HamTLD != nil {
		t := r.HamTLD
		bolum := &model.TLDBolumu{TLD: t.Record.TLD, Adresler: t.Record.Addresses}
		for _, c := range t.Record.Contacts {
			bolum.Iletisimler = append(bolum.Iletisimler, model.TLDIletisim{
				Rol: c.Role, Ad: c.Name, Kurum: c.Organization,
				Telefon: c.Phone, Faks: c.Fax, Eposta: c.Email,
			})
		}
		for _, ns := range t.Record.Nameservers {
			bolum.AdSunuculari = append(bolum.AdSunuculari, model.TLDAdSunucusu{Adres: ns.Host, IPler: ns.Addresses})
		}
		r.TLD = bolum
	}

	// Özet alanlarını doldur.
	if r.Kayit != nil {
		kayitli := r.Kayit.Durum == "registered"
		r.Ozet.Kayitli = &kayitli
	}
	if r.Siralama != nil {
		if r.Siralama.Cisco != nil && r.Siralama.Cisco.Listed {
			r.Ozet.CiscoSira = r.Siralama.Cisco.CurrentRank
		}
		if r.Siralama.Tranco != nil && r.Siralama.Tranco.Listed {
			r.Ozet.TrancoSira = r.Siralama.Tranco.CurrentRank
		}
	}
	if r.TLS != nil {
		r.Ozet.TLSDurumu = r.TLS.Durum
		r.Ozet.TLSSertifikaGun = r.TLS.KalanGun
	}
	if r.HamAlanAdi != nil {
		imzali := r.HamAlanAdi.DNS.DNSSEC.AuthenticatedData
		r.Ozet.DNSSECImzali = &imzali
	}
	if r.Guvenlik != nil {
		for _, k := range r.Guvenlik.Kategoriler {
			if k.Durum == "flagged" {
				r.Ozet.GuvenlikBayrak = append(r.Ozet.GuvenlikBayrak, k.Kategori)
			}
		}
	}
}

func kimlik(ortakAdlar, kurumlar, birimler, yereller, bolgeler, ulkeler []string) model.KimlikBilgisi {
	k := model.KimlikBilgisi{
		OrtakAdlar: ortakAdlar,
		Kurumlar:   kurumlar,
		Birimler:   birimler,
	}
	for _, v := range append(append([]string{}, yereller...), bolgeler...) {
		if strings.TrimSpace(v) != "" {
			k.Konumlar = append(k.Konumlar, v)
		}
	}
	k.Konumlar = append(k.Konumlar, ulkeler...)
	return k
}

// Ham yanıt erişimcileri, Topla sonrası keşif korelasyonu için kullanılır.

// DNSKaynagi, ham alan adı yanıtını döner.
func DNSKaynagi(r *model.Rapor) *model.AlanAdiBilgisi { return r.HamAlanAdi }

// TLSKaynagi, ham TLS yanıtını döner.
func TLSKaynagi(r *model.Rapor) *model.TLSBilgisi { return r.HamTLS }

// WebKaynagi, ham web profili yanıtını döner.
func WebKaynagi(r *model.Rapor) *model.WebProfili { return r.HamWeb }

// AramaKaynagi, ham arama yanıtını döner.
func AramaKaynagi(r *model.Rapor) *model.AramaBilgisi { return r.HamArama }

// IPKaynagi, ham IP yanıtını döner.
func IPKaynagi(r *model.Rapor) *model.IPBilgisi { return r.HamIP }

// WHOISKaynagi, ham WHOIS yanıtını döner.
func WHOISKaynagi(r *model.Rapor) *model.WHOISBilgisi { return r.HamWHOIS }

// Isle, rapor dönüşümünü dışarıya açar.
func Isle(r *model.Rapor) { isle(r) }
