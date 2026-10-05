package models

type DNSRecord struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	TTL        int    `json:"ttl"`
	Value      string `json:"value"`
	ValueParts []struct {
		Text       string `json:"text"`
		Target     string `json:"target"`
		TargetKind string `json:"targetKind"`
	} `json:"valueParts"`
}

type DomainInfo struct {
	Target struct {
		Ascii   string   `json:"ascii"`
		Unicode string   `json:"unicode"`
		Labels  []string `json:"labels"`
		Kind    string   `json:"kind"`
		TLD     string   `json:"tld"`
	} `json:"target"`
	RegistrationTarget struct {
		Ascii   string `json:"ascii"`
		Unicode string `json:"unicode"`
		Kind    string `json:"kind"`
		TLD     string `json:"tld"`
	} `json:"registrationTarget"`
	DNS struct {
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
			Status     string   `json:"status"`
			Confidence string   `json:"confidence"`
			Summary    string   `json:"summary"`
			Evidence   []string `json:"evidence"`
		} `json:"registrationInference"`
	} `json:"dns"`
	RDAP struct {
		Available          bool     `json:"available"`
		RegistrationStatus string   `json:"registrationStatus"`
		SourceURL          string   `json:"sourceUrl"`
		RegistryHandle     string   `json:"registryHandle"`
		Nameservers        []string `json:"nameservers"`
		Registrar          struct {
			Name   string `json:"name"`
			Handle string `json:"handle"`
			ID     string `json:"id"`
			URL    string `json:"url"`
		} `json:"registrar"`
		Events []struct {
			Action string `json:"action"`
			Date   string `json:"date"`
		} `json:"events"`
		Statuses []string `json:"statuses"`
	} `json:"rdap"`
	ContentSafety struct {
		Checked bool `json:"checked"`
		Malware struct {
			State       string `json:"state"`
			Explanation string `json:"explanation"`
		} `json:"malware"`
		Adult struct {
			State       string `json:"state"`
			Explanation string `json:"explanation"`
		} `json:"adult"`
	} `json:"contentSafety"`
}

type RankingInfo struct {
	Domain       string `json:"domain"`
	Provider     string `json:"provider"`
	CurrentDate  string `json:"currentDate"`
	PreviousDate string `json:"previousDate"`
	CurrentRank  int    `json:"currentRank"`
	PreviousRank int    `json:"previousRank"`
	Change       int    `json:"change"`
	Listed       bool   `json:"listed"`
	Source       struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"source"`
}

type SafetyInfo struct {
	Target     string `json:"target"`
	Categories struct {
		Malware struct {
			State       string   `json:"state"`
			FlaggedBy   []string `json:"flaggedBy"`
			CheckedBy   []string `json:"checkedBy"`
			Explanation string   `json:"explanation"`
		} `json:"malware"`
		Adult struct {
			State       string   `json:"state"`
			FlaggedBy   []string `json:"flaggedBy"`
			CheckedBy   []string `json:"checkedBy"`
			Explanation string   `json:"explanation"`
		} `json:"adult"`
		Advertising struct {
			State       string   `json:"state"`
			FlaggedBy   []string `json:"flaggedBy"`
			CheckedBy   []string `json:"checkedBy"`
			Explanation string   `json:"explanation"`
		} `json:"advertising"`
	} `json:"categories"`
	Providers []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Status  string `json:"status"`
		Signals map[string]struct {
			State       string `json:"state"`
			Explanation string `json:"explanation"`
		} `json:"signals"`
	} `json:"providers"`
}

type TLSInfo struct {
	Target      string `json:"target"`
	TargetKind  string `json:"targetKind"`
	Port        int    `json:"port"`
	SNI         string `json:"sni"`
	Status      string `json:"status"`
	Authorized  bool   `json:"authorized"`
	Certificate struct {
		Subject struct {
			CommonNames   []string `json:"commonNames"`
			Organizations []string `json:"organizations"`
		} `json:"subject"`
		Issuer struct {
			Organizations []string `json:"organizations"`
		} `json:"issuer"`
		ValidFrom          string   `json:"validFrom"`
		ValidTo            string   `json:"validTo"`
		DaysRemaining      int      `json:"daysRemaining"`
		SignatureAlgorithm string   `json:"signatureAlgorithm"`
		KeyType            string   `json:"keyType"`
		KeyBits            int      `json:"keyBits"`
		DNSNames           []struct {
			Value        string `json:"value"`
			Unicode      string `json:"unicode"`
			Wildcard     bool   `json:"wildcard"`
			Relationship string `json:"relationship"`
		} `json:"dnsNames"`
		IPAddresses []string `json:"ipAddresses"`
	} `json:"certificate"`
	Observations []struct {
		Level   string `json:"level"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"observations"`
	CertificateTransparencyURL string `json:"certificateTransparencyUrl"`
}

type WebProfileInfo struct {
	Target         string            `json:"target"`
	RequestedURL   string            `json:"requestedUrl"`
	FinalURL       string            `json:"finalUrl"`
	Protocol       string            `json:"protocol"`
	Status         int               `json:"status"`
	StatusText     string            `json:"statusText"`
	Headers        map[string]string `json:"headers"`
	Metadata       map[string]string `json:"metadata"`
	DiscoveredURLs []string          `json:"discoveredUrls"`
	Organizations  []map[string]any  `json:"organizations"`
	Contacts       map[string]any    `json:"contacts"`
}

type SearchInfo struct {
	Query   string `json:"query"`
	Results []struct {
		Title             string `json:"title"`
		URL               string `json:"url"`
		Host              string `json:"host"`
		RegistrableDomain string `json:"registrableDomain"`
		DisplayURL        string `json:"displayUrl"`
		Snippet           string `json:"snippet"`
	} `json:"results"`
	Suggestions []string `json:"suggestions"`
}

type IPInfo struct {
	Target struct {
		Address     string `json:"address"`
		Version     int    `json:"version"`
		ReverseName string `json:"reverseName"`
		Scope       string `json:"scope"`
		IsPublic    bool   `json:"isPublic"`
	} `json:"target"`
	ReverseDNS struct {
		QueryName            string   `json:"queryName"`
		Names                []string `json:"names"`
		ForwardConfirmed     bool     `json:"forwardConfirmed"`
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
		CountryCode string  `json:"countryCode"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	} `json:"location"`
	Allocation struct {
		Registry     string   `json:"registry"`
		Handle       string   `json:"handle"`
		Name         string   `json:"name"`
		Type         string   `json:"type"`
		CountryCode  string   `json:"countryCode"`
		Country      string   `json:"country"`
		StartAddress string   `json:"startAddress"`
		EndAddress   string   `json:"endAddress"`
		CIDRs        []string `json:"cidrs"`
		Statuses     []string `json:"statuses"`
		Port43       string   `json:"port43"`
		SourceURL    string   `json:"sourceUrl"`
	} `json:"allocation"`
}

type DirectTLSResult struct {
	SubjectCN     string   `json:"subjectCN"`
	IssuerOrg     string   `json:"issuerOrg"`
	SANs          []string `json:"sans"`
	ValidFrom     string   `json:"validFrom"`
	ValidTo       string   `json:"validTo"`
	DaysRemaining int      `json:"daysRemaining"`
}

type DiscoveredHosts struct {
	Subdomains []string `json:"subdomains"`
	Related    []string `json:"related"`
	IPs        []string `json:"ips"`
}

type TargetReport struct {
	Target          string           `json:"target"`
	Kind            string           `json:"kind"`
	Domain          *DomainInfo      `json:"domain,omitempty"`
	CiscoRank       *RankingInfo     `json:"ciscoRank,omitempty"`
	TrancoRank      *RankingInfo     `json:"trancoRank,omitempty"`
	Safety          *SafetyInfo      `json:"safety,omitempty"`
	TLS             *TLSInfo         `json:"tls,omitempty"`
	DirectTLS       *DirectTLSResult `json:"directTls,omitempty"`
	WebProfile      *WebProfileInfo  `json:"webProfile,omitempty"`
	IP              *IPInfo          `json:"ip,omitempty"`
	Search          *SearchInfo      `json:"search,omitempty"`
	DiscoveredHosts DiscoveredHosts  `json:"discoveredHosts"`
}
