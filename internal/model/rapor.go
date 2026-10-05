package model

// Rapor, aracın kullanıcıya ve ajanlara sunduğu kararlı şemadır.
// JSON anahtarları İngilizcedir (makine sözleşmesi); değerler ve açıklamalar Türkçedir.
type Rapor struct {
	Hedef          string               `json:"target"`
	Tur            string               `json:"kind"`
	Arac           AracBilgisi          `json:"tool"`
	Olusturma      string               `json:"generatedAt"`
	Ozet           Ozet                 `json:"summary"`
	Siralama       *SiralamaBolumu      `json:"rankings,omitempty"`
	Kayit          *KayitBolumu         `json:"registration,omitempty"`
	DNS            *DNSBolumuRapor      `json:"dns,omitempty"`
	TLS            *TLSBolumu           `json:"tls,omitempty"`
	Web            *WebBolumu           `json:"web,omitempty"`
	Guvenlik       *GuvenlikBolumu      `json:"safety,omitempty"`
	Siniflandirma  *SiniflandirmaBolumu `json:"classifications,omitempty"`
	IP             *IPBolumu            `json:"ip,omitempty"`
	ASN            *ASNBolumu           `json:"asn,omitempty"`
	TLD            *TLDBolumu           `json:"tld,omitempty"`
	Akis           *AkisBolumu          `json:"feed,omitempty"`
	Cozumleyiciler []CozumleyiciSonucu  `json:"resolvers,omitempty"`
	Kesif          KesifBolumu          `json:"discovery"`

	// Ham API yanıtları; keşif korelasyonu ve dönüşüm için tutulur, JSON'a yazılmaz.
	HamAlanAdi       *AlanAdiBilgisi     `json:"-"`
	HamGuvenlik      *GuvenlikBilgisi    `json:"-"`
	HamSiniflandirma *Siniflandirma      `json:"-"`
	HamTLS           *TLSBilgisi         `json:"-"`
	HamWeb           *WebProfili         `json:"-"`
	HamArama         *AramaBilgisi       `json:"-"`
	HamWHOIS         *WHOISBilgisi       `json:"-"`
	HamIP            *IPBilgisi          `json:"-"`
	HamASNlar        map[int]*ASNBilgisi `json:"-"`
	HamTLD           *TLDGosterge        `json:"-"`
	Kaynaklar        []KaynakDurumu      `json:"sources"`
	Uyarilar         []string            `json:"warnings"`
}

// AracBilgisi, üretici aracın kimliğidir.
type AracBilgisi struct {
	Ad    string `json:"name"`
	Surum string `json:"version"`
}

// Ozet, raporun makine tarafından hızlı tüketilen özetidir.
type Ozet struct {
	Kayitli         *bool    `json:"registered,omitempty"`
	CiscoSira       *int     `json:"ciscoRank,omitempty"`
	TrancoSira      *int     `json:"trancoRank,omitempty"`
	DNSSECImzali    *bool    `json:"dnssecSigned,omitempty"`
	TLSDurumu       string   `json:"tlsStatus,omitempty"`
	TLSSertifikaGun *int     `json:"tlsDaysRemaining,omitempty"`
	GuvenlikBayrak  []string `json:"safetyFlags,omitempty"`
	AltAlanSayisi   int      `json:"subdomainCount"`
	IliskiliSayisi  int      `json:"relatedCount"`
	IPAdresSayisi   int      `json:"ipCount"`
}

// KaynakDurumu, tek bir veri kaynağının sonucudur.
type KaynakDurumu struct {
	ID        string `json:"id"`
	Ad        string `json:"name"`
	Durum     string `json:"status"` // ok | error | throttled | skipped | empty
	Hata      string `json:"error,omitempty"`
	GecikmeMs int64  `json:"latencyMs"`
	Onbellek  bool   `json:"cacheHit,omitempty"`
}

// SiralamaBolumu, popülerlik sıralamalarıdır.
type SiralamaBolumu struct {
	Cisco  *Siralama `json:"cisco,omitempty"`
	Tranco *Siralama `json:"tranco,omitempty"`
}

// KayitBolumu, alan adı kayıt bilgisidir (RDAP + WHOIS birleşik).
type KayitBolumu struct {
	Durum            string   `json:"status"` // registered | not_found | unknown
	Kaynak           string   `json:"source"` // rdap | whois | rdap+whois
	Kayitci          string   `json:"registrar,omitempty"`
	KayitciID        string   `json:"registrarId,omitempty"`
	KayitciURL       string   `json:"registrarUrl,omitempty"`
	KayitTarihi      string   `json:"created,omitempty"`
	GuncellemeTarihi string   `json:"updated,omitempty"`
	BitisTarihi      string   `json:"expires,omitempty"`
	Durumlar         []string `json:"statuses,omitempty"`
	AdSunuculari     []string `json:"nameservers,omitempty"`
	KayitDefteri     string   `json:"registryHandle,omitempty"`
	WHOISSunucusu    string   `json:"whoisServer,omitempty"`
	RDAPKaynagi      string   `json:"rdapSourceUrl,omitempty"`
}

// DNSBolumuRapor, DNS kayıtlarının gruplanmış hâlidir.
type DNSBolumuRapor struct {
	Durum    string                     `json:"status"`
	Rcode    int                        `json:"rcode"`
	Kayitlar map[string][]DNSKaydiRapor `json:"records"`
}

// DNSKaydiRapor, tek bir DNS kaydıdır.
type DNSKaydiRapor struct {
	Ad    string `json:"name"`
	TTL   int    `json:"ttl"`
	Deger string `json:"value"`
}

// TLSBolumu, sertifika istihbaratıdır.
type TLSBolumu struct {
	Durum           string        `json:"status"`
	Dogrulandi      bool          `json:"authorized"`
	SNI             string        `json:"sni,omitempty"`
	Ucnokta         string        `json:"endpoint,omitempty"`
	Konu            KimlikBilgisi `json:"subject"`
	Yayimci         KimlikBilgisi `json:"issuer"`
	GecerlilikBas   string        `json:"validFrom,omitempty"`
	GecerlilikSon   string        `json:"validTo,omitempty"`
	KalanGun        *int          `json:"daysRemaining,omitempty"`
	ImzaAlgoritmasi string        `json:"signatureAlgorithm,omitempty"`
	AnahtarTuru     string        `json:"keyType,omitempty"`
	AnahtarBit      int           `json:"keyBits,omitempty"`
	SeriNumarasi    string        `json:"serialNumber,omitempty"`
	ParmakIzi       string        `json:"fingerprintSha256,omitempty"`
	Adlar           []TLSAdi      `json:"dnsNames"`
	IPAdresleri     []string      `json:"ipAddresses,omitempty"`
	Gozlemler       []Gozlem      `json:"observations,omitempty"`
	SeffaflikURL    string        `json:"certificateTransparencyUrl,omitempty"`
}

// KimlikBilgisi, sertifika konu/yayımcı kimliğidir.
type KimlikBilgisi struct {
	OrtakAdlar []string `json:"commonNames,omitempty"`
	Kurumlar   []string `json:"organizations,omitempty"`
	Birimler   []string `json:"organizationalUnits,omitempty"`
	Konumlar   []string `json:"locations,omitempty"`
}

// TLSAdi, sertifikadaki bir DNS adıdır.
type TLSAdi struct {
	Ad     string `json:"value"`
	Joker  bool   `json:"wildcard,omitempty"`
	Iliski string `json:"relationship,omitempty"`
}

// Gozlem, TLS analiz notudur.
type Gozlem struct {
	Seviye string `json:"level"`
	Kod    string `json:"code"`
	Mesaj  string `json:"message"`
}

// WebBolumu, ana sayfa profilidir.
type WebBolumu struct {
	SonURL     string            `json:"finalUrl,omitempty"`
	Protokol   string            `json:"protocol,omitempty"`
	DurumKodu  int               `json:"status,omitempty"`
	DurumMetni string            `json:"statusText,omitempty"`
	IcerikTuru string            `json:"contentType,omitempty"`
	Basliklar  map[string]string `json:"headers,omitempty"`
	Meta       map[string]string `json:"metadata,omitempty"`
	BulunanURL []string          `json:"discoveredUrls,omitempty"`
	Kurumlar   []KurumBilgisi    `json:"organizations,omitempty"`
	Epostalar  []string          `json:"emails,omitempty"`
	Telefonlar []string          `json:"phones,omitempty"`
	Uyarilar   []string          `json:"warnings,omitempty"`
}

// KurumBilgisi, ana sayfada bulunan kuruluş kaydıdır.
type KurumBilgisi struct {
	Ad         string   `json:"name,omitempty"`
	Tur        string   `json:"type,omitempty"`
	Website    string   `json:"url,omitempty"`
	Epostalar  []string `json:"emails,omitempty"`
	Telefonlar []string `json:"phones,omitempty"`
}

// GuvenlikBolumu, DNS filtre sağlayıcılarının sonuçlarıdır.
type GuvenlikBolumu struct {
	Kategoriler  []GuvenlikKategorisiRapor `json:"categories"`
	Saglayicilar []SaglayiciGuvenlik       `json:"providers"`
	Bayraklandi  bool                      `json:"flagged"`
}

// GuvenlikKategorisiRapor, toplu kategori sonucudur.
type GuvenlikKategorisiRapor struct {
	Kategori    string   `json:"category"`
	Etiket      string   `json:"label"`
	Durum       string   `json:"state"`
	Bayraklayan []string `json:"flaggedBy,omitempty"`
	KontrolEden []string `json:"checkedBy,omitempty"`
	Aciklama    string   `json:"explanation,omitempty"`
}

// SaglayiciGuvenlik, tek bir DNS filtresinin kategori sonuçlarıdır.
type SaglayiciGuvenlik struct {
	ID        string            `json:"id"`
	Ad        string            `json:"name"`
	Durum     string            `json:"status"`
	BelgeURL  string            `json:"documentationUrl,omitempty"`
	Sinyaller map[string]string `json:"signals"`
}

// SiniflandirmaBolumu, HaGeZi liste sınıflandırmasıdır.
type SiniflandirmaBolumu struct {
	EslesmeSayisi int                    `json:"matchCount"`
	Eslesmeler    []SiniflandirmaEslesme `json:"matches"`
	VeriTarihi    string                 `json:"sourceDataAt,omitempty"`
	Surum         string                 `json:"releaseId,omitempty"`
	Kaynak        string                 `json:"source,omitempty"`
}

// SiniflandirmaEslesme, tek bir liste eşleşmesidir.
type SiniflandirmaEslesme struct {
	Kategori       string   `json:"category"`
	Etiket         string   `json:"label"`
	Aciklama       string   `json:"description,omitempty"`
	Onem           string   `json:"severity,omitempty"`
	EslesenDeger   string   `json:"matchedValue,omitempty"`
	Kapsam         string   `json:"scope,omitempty"`
	KaynakListeler []string `json:"sourceFeeds,omitempty"`
}

// IPBolumu, IP adresi istihbaratıdır.
type IPBolumu struct {
	Adres           string    `json:"address"`
	Surum           int       `json:"version"`
	Kapsam          string    `json:"scope,omitempty"`
	GenelMi         bool      `json:"isPublic"`
	TersDNS         []string  `json:"reverseDns,omitempty"`
	TersDNSSorgu    string    `json:"reverseDnsQuery,omitempty"`
	IleriDogrulandi *bool     `json:"forwardConfirmed,omitempty"`
	Duyurulmus      bool      `json:"announced"`
	Prefix          string    `json:"prefix,omitempty"`
	ASNs            []ASNOzet `json:"asns,omitempty"`
	BGPGozlem       string    `json:"bgpObservedAt,omitempty"`
	Ulke            string    `json:"countryCode,omitempty"`
	Sehir           string    `json:"city,omitempty"`
	Enlem           *float64  `json:"latitude,omitempty"`
	Boylam          *float64  `json:"longitude,omitempty"`
	KonumKapsama    float64   `json:"locationCoveredPercentage,omitempty"`
	KayitDefteri    string    `json:"registry,omitempty"`
	TahsisAdi       string    `json:"allocationName,omitempty"`
	TahsisTuru      string    `json:"allocationType,omitempty"`
	Aralik          string    `json:"range,omitempty"`
	CIDRler         []string  `json:"cidrs,omitempty"`
	RDAPKaynagi     string    `json:"rdapSourceUrl,omitempty"`
}

// ASNOzet, ASN kimliğidir.
type ASNOzet struct {
	Numara int    `json:"number"`
	Sahip  string `json:"holder,omitempty"`
}

// ASNBolumu, ASN istihbaratıdır.
type ASNBolumu struct {
	Numara        int        `json:"number"`
	Etiket        string     `json:"label"`
	Sahip         string     `json:"holder,omitempty"`
	Duyurulmus    bool       `json:"announced"`
	Blok          string     `json:"block,omitempty"`
	PrefixSayisi  int        `json:"prefixCount"`
	Prefixler     []string   `json:"prefixes,omitempty"`
	PrefixKesildi bool       `json:"prefixesTruncated,omitempty"`
	YukariAkim    int        `json:"upstreamCount"`
	AsagiAkim     int        `json:"downstreamCount"`
	Komsular      []ASNKomsu `json:"neighbours,omitempty"`
	KayitDefteri  string     `json:"registry,omitempty"`
	KayitAdi      string     `json:"registryName,omitempty"`
	KaynakURL     string     `json:"rdapSourceUrl,omitempty"`
}

// ASNKomsu, BGP komşu ASN kaydıdır.
type ASNKomsu struct {
	Numara int    `json:"number"`
	Iliski string `json:"relationship"`
	Guc    int    `json:"power,omitempty"`
}

// TLDBolumu, üst düzey alan adı bilgisidir.
type TLDBolumu struct {
	TLD          string          `json:"tld"`
	Adresler     []string        `json:"addresses,omitempty"`
	Iletisimler  []TLDIletisim   `json:"contacts,omitempty"`
	AdSunuculari []TLDAdSunucusu `json:"nameservers,omitempty"`
}

// TLDIletisim, IANA TLD iletişim kaydıdır.
type TLDIletisim struct {
	Rol     string `json:"role,omitempty"`
	Ad      string `json:"name,omitempty"`
	Kurum   string `json:"organization,omitempty"`
	Telefon string `json:"phone,omitempty"`
	Faks    string `json:"fax,omitempty"`
	Eposta  string `json:"email,omitempty"`
}

// TLDAdSunucusu, TLD ad sunucusudur.
type TLDAdSunucusu struct {
	Adres string   `json:"host"`
	IPler []string `json:"addresses,omitempty"`
}

// AkisBolumu, RSS akışıdır.
type AkisBolumu struct {
	Baslik     string      `json:"title,omitempty"`
	Aciklama   string      `json:"description,omitempty"`
	Guncelleme string      `json:"lastBuildDate,omitempty"`
	OgeSayisi  int         `json:"itemCount"`
	Ogeler     []AkisOgesi `json:"items"`
}

// AkisOgesi, tek bir RSS öğesidir.
type AkisOgesi struct {
	Baslik      string `json:"title"`
	Baglanti    string `json:"link"`
	Aciklama    string `json:"description,omitempty"`
	YayimTarihi string `json:"pubDate,omitempty"`
	Kategori    string `json:"category,omitempty"`
}

// CozumleyiciSonucu, bir DNS çözümleyicisinin sonucudur.
type CozumleyiciSonucu struct {
	ID       string   `json:"id"`
	Ad       string   `json:"name"`
	Rol      string   `json:"role,omitempty"`
	Durum    string   `json:"status"` // ok | error
	Engelli  bool     `json:"blocked"`
	Hata     string   `json:"error,omitempty"`
	Kayitlar []string `json:"records,omitempty"`
}

// KesifBolumu, pasif keşif korelasyonudur.
type KesifBolumu struct {
	AltAlanlar []KesifOgesi `json:"subdomains"`
	Iliskili   []KesifOgesi `json:"related"`
	IPler      []KesifOgesi `json:"ips"`
	Epostalar  []KesifOgesi `json:"emails"`
	Telefonlar []KesifOgesi `json:"phones"`
	Kurumlar   []KesifOgesi `json:"organizations"`
}

// KesifOgesi, keşfedilen bir varlık ve onu ortaya çıkaran kaynaklardır.
type KesifOgesi struct {
	Deger     string   `json:"value"`
	Kaynaklar []string `json:"sources"`
}
