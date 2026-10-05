// Package model, domain.glass API yanıtlarının ham tiplerini ve aracın
// rapor tiplerini barındırır.
//
// Ham tipler ("ham" önekli dosyalarda) API sözleşmesini birebir yansıtır;
// rapor tipleri ise kullanıcıya ve ajanlara sunulan kararlı şemadır.
package model

// DNSRecord, domain.glass DNS kaydı.
type DNSRecord struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	TTL        int         `json:"ttl"`
	Value      string      `json:"value"`
	ValueParts []ValuePart `json:"valueParts,omitempty"`
}

// ValuePart, DNS kaydı değerinin parçalanmış hâlidir.
type ValuePart struct {
	Text       string `json:"text"`
	Target     string `json:"target"`
	TargetKind string `json:"targetKind"`
}

// DNSBolumu, bir alan adının DNS bölümüdür.
type DNSBolumu struct {
	Status  string `json:"status"`
	Rcode   int    `json:"rcode"`
	Records struct {
		A     []DNSRecord `json:"A"`
		AAAA  []DNSRecord `json:"AAAA"`
		CNAME []DNSRecord `json:"CNAME"`
		MX    []DNSRecord `json:"MX"`
		NS    []DNSRecord `json:"NS"`
		TXT   []DNSRecord `json:"TXT"`
		CAA   []DNSRecord `json:"CAA"`
		SOA   []DNSRecord `json:"SOA"`
	} `json:"records"`
	DNSSEC struct {
		Checked           bool `json:"checked"`
		AuthenticatedData bool `json:"authenticatedData"`
	} `json:"dnssec"`
	RegistrationInference struct {
		Status              string   `json:"status"`
		Confidence          string   `json:"confidence"`
		Summary             string   `json:"summary"`
		Caveat              string   `json:"caveat"`
		PositiveRecordTypes []string `json:"positiveRecordTypes"`
		Evidence            []string `json:"evidence"`
	} `json:"registrationInference"`
}

// TumKayitlar, DNS bölümündeki tüm kayıtları tek listede döner.
func (d DNSBolumu) TumKayitlar() []DNSRecord {
	var out []DNSRecord
	for _, grup := range [][]DNSRecord{d.Records.A, d.Records.AAAA, d.Records.CNAME, d.Records.MX, d.Records.NS, d.Records.TXT, d.Records.CAA, d.Records.SOA} {
		out = append(out, grup...)
	}
	return out
}

// Hedef, sorgulanan adın kimliğidir.
type Hedef struct {
	Ascii   string   `json:"ascii"`
	Unicode string   `json:"unicode"`
	Labels  []string `json:"labels"`
	Kind    string   `json:"kind"`
	TLD     string   `json:"tld"`
}

// NameserverDetay, RDAP ad sunucusu ayrıntısı.
type NameserverDetay struct {
	Name string   `json:"name"`
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// RDAPOlay, kayıt yaşam döngüsü olayı.
type RDAPOlay struct {
	Action string `json:"action"`
	Date   string `json:"date"`
}

// RDAPDurumAciklama, kayıt durumu açıklaması.
type RDAPDurumAciklama struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

// RDAPBilgisi, alan adı kayıt verisi.
type RDAPBilgisi struct {
	Available          bool     `json:"available"`
	RegistrationStatus string   `json:"registrationStatus"`
	BootstrapURLs      []string `json:"bootstrapUrls"`
	SourceURL          string   `json:"sourceUrl"`
	RegistryHandle     string   `json:"registryHandle"`
	Port43             string   `json:"port43"`
	Registrar          struct {
		Name   string `json:"name"`
		Handle string `json:"handle"`
		ID     string `json:"id"`
		URL    string `json:"url"`
	} `json:"registrar"`
	Events             []RDAPOlay          `json:"events"`
	Statuses           []string            `json:"statuses"`
	StatusExplanations []RDAPDurumAciklama `json:"statusExplanations"`
	Nameservers        []string            `json:"nameservers"`
	NameserverDetails  []NameserverDetay   `json:"nameserverDetails"`
	DNSSEC             struct {
		DelegationSigned *bool `json:"delegationSigned"`
		ZoneSigned       *bool `json:"zoneSigned"`
		MaxSigLife       *int  `json:"maxSigLife"`
		DSData           []struct {
			KeyTag     int    `json:"keyTag"`
			Algorithm  int    `json:"algorithm"`
			DigestType int    `json:"digestType"`
			Digest     string `json:"digest"`
		} `json:"dsData"`
	} `json:"dnssec"`
}

// IcerikGuvenligi, Cloudflare içerik filtresi sonucudur.
type IcerikGuvenligi struct {
	Checked   bool   `json:"checked"`
	Target    string `json:"target"`
	Malware   Sinyal `json:"malware"`
	Adult     Sinyal `json:"adult"`
	CheckedAt string `json:"checkedAt"`
}

// Sinyal, tek bir filtre kategorisinin durumudur.
type Sinyal struct {
	State       string `json:"state"`
	Explanation string `json:"explanation"`
}

// AlanAdiBilgisi, /api/v1/domain/{hedef} yanıtıdır.
type AlanAdiBilgisi struct {
	Target             Hedef             `json:"target"`
	RegistrationTarget Hedef             `json:"registrationTarget"`
	DNS                DNSBolumu         `json:"dns"`
	RDAP               RDAPBilgisi       `json:"rdap"`
	ContentSafety      *IcerikGuvenligi  `json:"contentSafety"`
	GeneratedAt        string            `json:"generatedAt"`
	MaxAgeSeconds      int               `json:"maxAgeSeconds"`
	Errors             []APIIcerikHatasi `json:"errors"`
}

// APIIcerikHatasi, API içi kısmi hata kaydıdır.
type APIIcerikHatasi struct {
	Source  string `json:"source"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GuvenlikBilgisi, /api/v1/domain/{hedef}/safety yanıtıdır.
type GuvenlikBilgisi struct {
	Target     string `json:"target"`
	Categories struct {
		Malware     GuvenlikKategorisi `json:"malware"`
		Adult       GuvenlikKategorisi `json:"adult"`
		Advertising GuvenlikKategorisi `json:"advertising"`
	} `json:"categories"`
	Providers []struct {
		ID               string                        `json:"id"`
		Name             string                        `json:"name"`
		Status           string                        `json:"status"`
		DocumentationURL string                        `json:"documentationUrl"`
		Signals          map[string]GuvenlikKategorisi `json:"signals"`
	} `json:"providers"`
}

// GuvenlikKategorisi, bir sağlayıcının tek kategori sinyalidir.
type GuvenlikKategorisi struct {
	State       string   `json:"state"`
	FlaggedBy   []string `json:"flaggedBy"`
	CheckedBy   []string `json:"checkedBy"`
	Explanation string   `json:"explanation"`
}

// Siniflandirma, /api/v1/classifications/{hedef} yanıtıdır.
type Siniflandirma struct {
	Target  string `json:"target"`
	Kind    string `json:"kind"`
	Matches []struct {
		Category     string   `json:"category"`
		Label        string   `json:"label"`
		Description  string   `json:"description"`
		Severity     string   `json:"severity"`
		MatchedValue string   `json:"matchedValue"`
		Scope        string   `json:"scope"`
		SourceFeeds  []string `json:"sourceFeeds"`
	} `json:"matches"`
	Flagged            bool   `json:"flagged"`
	SuppressTopResults bool   `json:"suppressTopResults"`
	SourceDataAt       string `json:"sourceDataAt"`
	ReleaseID          string `json:"releaseId"`
	Source             struct {
		Name       string `json:"name"`
		ProjectURL string `json:"projectUrl"`
		LicenseURL string `json:"licenseUrl"`
	} `json:"source"`
}

// Siralama, popülerlik sıralaması yanıtıdır.
type Siralama struct {
	Domain       string `json:"domain"`
	Provider     string `json:"provider"`
	CurrentDate  string `json:"currentDate"`
	PreviousDate string `json:"previousDate"`
	CurrentRank  *int   `json:"currentRank"`
	PreviousRank *int   `json:"previousRank"`
	Change       *int   `json:"change"`
	Listed       bool   `json:"listed"`
	Source       struct {
		Name           string `json:"name"`
		URL            string `json:"url"`
		InformationURL string `json:"informationUrl"`
		LastModified   string `json:"lastModified"`
	} `json:"source"`
}

// TLSBilgisi, /api/v1/tls/{tür}/{hedef} yanıtıdır.
type TLSBilgisi struct {
	Target      string `json:"target"`
	TargetKind  string `json:"targetKind"`
	Port        int    `json:"port"`
	SNI         string `json:"sni"`
	Status      string `json:"status"`
	Authorized  bool   `json:"authorized"`
	Certificate *struct {
		Subject struct {
			CommonNames         []string `json:"commonNames"`
			Organizations       []string `json:"organizations"`
			OrganizationalUnits []string `json:"organizationalUnits"`
			Localities          []string `json:"localities"`
			States              []string `json:"states"`
			Countries           []string `json:"countries"`
		} `json:"subject"`
		Issuer struct {
			CommonNames         []string `json:"commonNames"`
			Organizations       []string `json:"organizations"`
			OrganizationalUnits []string `json:"organizationalUnits"`
			Localities          []string `json:"localities"`
			States              []string `json:"states"`
			Countries           []string `json:"countries"`
		} `json:"issuer"`
		ValidFrom          string `json:"validFrom"`
		ValidTo            string `json:"validTo"`
		DaysRemaining      *int   `json:"daysRemaining"`
		SignatureAlgorithm string `json:"signatureAlgorithm"`
		KeyType            string `json:"keyType"`
		KeyBits            int    `json:"keyBits"`
		KeyCurve           string `json:"keyCurve"`
		SerialNumber       string `json:"serialNumber"`
		FingerprintSHA256  string `json:"fingerprintSha256"`
		DNSNames           []struct {
			Value        string `json:"value"`
			Unicode      string `json:"unicode"`
			Wildcard     bool   `json:"wildcard"`
			Relationship string `json:"relationship"`
		} `json:"dnsNames"`
		IPAddresses     []string `json:"ipAddresses"`
		OmittedSANCount int      `json:"omittedSanCount"`
	} `json:"certificate"`
	Observations []struct {
		Level   string `json:"level"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"observations"`
	CertificateTransparencyURL string `json:"certificateTransparencyUrl"`
}

// WebProfili, /api/v1/web-profile/{hedef} yanıtıdır.
type WebProfili struct {
	Target          string            `json:"target"`
	TargetKind      string            `json:"targetKind"`
	RequestedURL    string            `json:"requestedUrl"`
	FinalURL        string            `json:"finalUrl"`
	Protocol        string            `json:"protocol"`
	Status          int               `json:"status"`
	StatusText      string            `json:"statusText"`
	ContentType     string            `json:"contentType"`
	HTTPSPortState  string            `json:"httpsPortState"`
	TLSFallbackUsed bool              `json:"tlsFallbackUsed"`
	Headers         map[string]string `json:"headers"`
	Metadata        map[string]string `json:"metadata"`
	DiscoveredURLs  []string          `json:"discoveredUrls"`
	Organizations   []map[string]any  `json:"organizations"`
	Contacts        map[string]any    `json:"contacts"`
	Warnings        []string          `json:"warnings"`
}

// WHOISBilgisi, /api/v1/whois/{hedef} POST yanıtıdır.
type WHOISBilgisi struct {
	Target     Hedef `json:"target"`
	Registered *bool `json:"registered"`
	Registrar  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"registrar"`
	Dates struct {
		Created string `json:"created"`
		Updated string `json:"updated"`
		Expires string `json:"expires"`
	} `json:"dates"`
	Statuses          []string          `json:"statuses"`
	Nameservers       []string          `json:"nameservers"`
	NameserverDetails []NameserverDetay `json:"nameserverDetails"`
	DNSSEC            *bool             `json:"dnssec"`
	RegistryHandle    string            `json:"registryHandle"`
	WHOISServer       string            `json:"whoisServer"`
}

// AramaSonucu, tek bir web arama sonucudur.
type AramaSonucu struct {
	Title             string `json:"title"`
	URL               string `json:"url"`
	Host              string `json:"host"`
	RegistrableDomain string `json:"registrableDomain"`
	DisplayURL        string `json:"displayUrl"`
	Snippet           string `json:"snippet"`
	ImageURL          string `json:"imageUrl"`
	Kind              string `json:"kind"`
	ASNNumber         *int   `json:"asnNumber"`
}

// AramaBilgisi, /api/v1/search yanıtıdır.
type AramaBilgisi struct {
	Query       string        `json:"query"`
	Results     []AramaSonucu `json:"results"`
	Suggestions []string      `json:"suggestions"`
}

// IPBilgisi, /api/v1/ip/{ip} yanıtıdır.
type IPBilgisi struct {
	Target struct {
		Address     string `json:"address"`
		Version     int    `json:"version"`
		ReverseName string `json:"reverseName"`
		Scope       string `json:"scope"`
		IsPublic    bool   `json:"isPublic"`
	} `json:"target"`
	ReverseDNS struct {
		QueryName           string   `json:"queryName"`
		Names               []string `json:"names"`
		Domains             []Hedef  `json:"domains"`
		ForwardConfirmed    bool     `json:"forwardConfirmed"`
		DNSSECAuthenticated bool     `json:"dnssecAuthenticated"`
	} `json:"reverseDns"`
	Routing struct {
		Announced bool   `json:"announced"`
		Prefix    string `json:"prefix"`
		ASNs      []struct {
			Number int    `json:"number"`
			Holder string `json:"holder"`
		} `json:"asns"`
		ObservedAt string `json:"observedAt"`
	} `json:"routing"`
	Location struct {
		CountryCode       string  `json:"countryCode"`
		City              string  `json:"city"`
		Latitude          float64 `json:"latitude"`
		Longitude         float64 `json:"longitude"`
		CoveredPercentage float64 `json:"coveredPercentage"`
		ObservedAt        string  `json:"observedAt"`
	} `json:"location"`
	Allocation struct {
		Registry     string     `json:"registry"`
		Handle       string     `json:"handle"`
		Name         string     `json:"name"`
		Type         string     `json:"type"`
		CountryCode  string     `json:"countryCode"`
		Country      string     `json:"country"`
		StartAddress string     `json:"startAddress"`
		EndAddress   string     `json:"endAddress"`
		CIDRs        []string   `json:"cidrs"`
		Statuses     []string   `json:"statuses"`
		Port43       string     `json:"port43"`
		SourceURL    string     `json:"sourceUrl"`
		Events       []RDAPOlay `json:"events"`
	} `json:"allocation"`
	Sources []struct {
		ID          string `json:"id"`
		Label       string `json:"label"`
		URL         string `json:"url"`
		Description string `json:"description"`
	} `json:"sources"`
	Errors []APIIcerikHatasi `json:"errors"`
}

// ASNBilgisi, /api/v1/asn/{asn} yanıtıdır.
type ASNBilgisi struct {
	Target struct {
		Number int    `json:"number"`
		Label  string `json:"label"`
		ASDot  string `json:"asdot"`
	} `json:"target"`
	Overview struct {
		Holder    string `json:"holder"`
		Announced bool   `json:"announced"`
		Block     struct {
			Resource    string `json:"resource"`
			Description string `json:"description"`
		} `json:"block"`
		ObservedAt string `json:"observedAt"`
	} `json:"overview"`
	Prefixes struct {
		Total        int      `json:"total"`
		Items        []string `json:"items"`
		Truncated    bool     `json:"truncated"`
		ObservedFrom string   `json:"observedFrom"`
		ObservedTo   string   `json:"observedTo"`
	} `json:"prefixes"`
	Neighbours struct {
		Counts struct {
			Upstream   int `json:"upstream"`
			Downstream int `json:"downstream"`
			Uncertain  int `json:"uncertain"`
			Unique     int `json:"unique"`
		} `json:"counts"`
		Items []struct {
			Number       int    `json:"number"`
			Relationship string `json:"relationship"`
			Power        int    `json:"power"`
		} `json:"items"`
	} `json:"neighbours"`
	Peering *struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		Website           string `json:"website"`
		PolicyURL         string `json:"policyUrl"`
		IRRASSet          string `json:"irrAsSet"`
		NetworkType       string `json:"networkType"`
		Scope             string `json:"scope"`
		TrafficRatio      string `json:"trafficRatio"`
		IPv4Prefixes      int    `json:"ipv4Prefixes"`
		IPv6Prefixes      int    `json:"ipv6Prefixes"`
		InternetExchanges int    `json:"internetExchanges"`
		Facilities        int    `json:"facilities"`
		SupportsIPv6      bool   `json:"supportsIpv6"`
		Policy            struct {
			General       string `json:"general"`
			Locations     string `json:"locations"`
			Contracts     string `json:"contracts"`
			RatioRequired bool   `json:"ratioRequired"`
		} `json:"policy"`
	} `json:"peering"`
	Registration struct {
		Source      string     `json:"source"`
		Handle      string     `json:"handle"`
		Name        string     `json:"name"`
		Statuses    []string   `json:"statuses"`
		Events      []RDAPOlay `json:"events"`
		WHOISServer string     `json:"whoisServer"`
		SourceURL   string     `json:"sourceUrl"`
	} `json:"registration"`
	Errors []APIIcerikHatasi `json:"errors"`
}

// TLDGosterge, IANA TLD bilgisi (tld-iana uç noktası).
type TLDGosterge struct {
	Record struct {
		TLD       string   `json:"tld"`
		Addresses []string `json:"addresses"`
		Contacts  []struct {
			Role         string `json:"role"`
			Name         string `json:"name"`
			Organization string `json:"organization"`
			Phone        string `json:"phone"`
			Fax          string `json:"fax"`
			Email        string `json:"email"`
		} `json:"contacts"`
		Nameservers []struct {
			Host      string   `json:"host"`
			Addresses []string `json:"addresses"`
		} `json:"nameservers"`
	} `json:"record"`
}
