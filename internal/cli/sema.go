package cli

// Sema, rapor JSON şemasını (JSON Schema 2020-12) döner.
//
// Şema bilinçli olarak "yapısal iskelet" düzeyindedir: üst düzey alanlar,
// keşif bölümü ve kaynak/uyarı dizileri kesin; iç içe bölümler ek alanlara
// açıktır. Ajanlar bu şemayı `domainglass sema` ile alıp
// doğrulama ve kod üretimi için kullanabilir.
var Sema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/void0x14/domainglass/schema/rapor.json",
  "title": "DomainGlassRapor",
  "description": "domainglass rapor şeması. JSON anahtarları İngilizcedir; değerler ve açıklamalar Türkçedir.",
  "type": "object",
  "required": ["target", "kind", "tool", "generatedAt", "summary", "discovery", "sources", "warnings"],
  "properties": {
    "target": { "type": "string", "description": "Sorgulanan alan adı, IP, ASN veya akış adı." },
    "kind": { "type": "string", "enum": ["domain", "ip", "asn", "tld", "feed"] },
    "tool": {
      "type": "object",
      "required": ["name"],
      "properties": {
        "name": { "const": "domainglass" },
        "commit": { "type": "string", "description": "İkiliyi üreten commit hash'i (7 hane)." },
        "dirty": { "type": "boolean", "description": "Derleme sırasında çalışma ağacı kirli miydi." },
        "go": { "type": "string" },
        "platform": { "type": "string" }
      }
    },
    "generatedAt": { "type": "string", "description": "RFC3339 UTC üretim zamanı." },
    "summary": {
      "type": "object",
      "description": "Makine tarafından hızlı tüketilen özet.",
      "properties": {
        "registered": { "type": "boolean" },
        "ciscoRank": { "type": ["integer", "null"] },
        "trancoRank": { "type": ["integer", "null"] },
        "dnssecSigned": { "type": "boolean" },
        "tlsStatus": { "type": "string" },
        "tlsDaysRemaining": { "type": ["integer", "null"] },
        "safetyFlags": { "type": "array", "items": { "type": "string" } },
        "subdomainCount": { "type": "integer" },
        "relatedCount": { "type": "integer" },
        "ipCount": { "type": "integer" }
      }
    },
    "rankings": {
      "type": "object",
      "properties": {
        "cisco": { "$ref": "#/$defs/ranking" },
        "tranco": { "$ref": "#/$defs/ranking" }
      }
    },
    "registration": {
      "type": "object",
      "properties": {
        "status": { "type": "string", "enum": ["registered", "not_found", "unknown"] },
        "source": { "type": "string" },
        "registrar": { "type": "string" },
        "registrarId": { "type": "string" },
        "created": { "type": "string" },
        "updated": { "type": "string" },
        "expires": { "type": "string" },
        "statuses": { "type": "array", "items": { "type": "string" } },
        "nameservers": { "type": "array", "items": { "type": "string" } }
      }
    },
    "dns": {
      "type": "object",
      "properties": {
        "status": { "type": "string", "description": "NOERROR, NXDOMAIN, SERVFAIL…" },
        "rcode": { "type": "integer" },
        "records": {
          "type": "object",
          "additionalProperties": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["name", "ttl", "value"],
              "properties": {
                "name": { "type": "string" },
                "ttl": { "type": "integer" },
                "value": { "type": "string" }
              }
            }
          }
        }
      }
    },
    "tls": {
      "type": "object",
      "properties": {
        "status": { "type": "string", "enum": ["available", "unavailable"] },
        "authorized": { "type": "boolean" },
        "sni": { "type": "string" },
        "subject": { "$ref": "#/$defs/identity" },
        "issuer": { "$ref": "#/$defs/identity" },
        "validFrom": { "type": "string" },
        "validTo": { "type": "string" },
        "daysRemaining": { "type": ["integer", "null"] },
        "dnsNames": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["value"],
            "properties": {
              "value": { "type": "string" },
              "wildcard": { "type": "boolean" },
              "relationship": { "type": "string" }
            }
          }
        },
        "ipAddresses": { "type": "array", "items": { "type": "string" } }
      }
    },
    "web": {
      "type": "object",
      "properties": {
        "finalUrl": { "type": "string" },
        "status": { "type": "integer" },
        "headers": { "type": "object", "additionalProperties": { "type": "string" } },
        "metadata": { "type": "object", "additionalProperties": { "type": "string" } }
      }
    },
    "safety": {
      "type": "object",
      "properties": {
        "flagged": { "type": "boolean" },
        "categories": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["category", "state"],
            "properties": {
              "category": { "type": "string", "enum": ["malware", "adult", "advertising"] },
              "label": { "type": "string" },
              "state": { "type": "string", "enum": ["flagged", "not_flagged", "unknown"] },
              "flaggedBy": { "type": "array", "items": { "type": "string" } },
              "checkedBy": { "type": "array", "items": { "type": "string" } }
            }
          }
        },
        "providers": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["id", "name", "signals"],
            "properties": {
              "id": { "type": "string" },
              "name": { "type": "string" },
              "status": { "type": "string" },
              "signals": { "type": "object", "additionalProperties": { "type": "string" } }
            }
          }
        }
      }
    },
    "classifications": {
      "type": "object",
      "properties": {
        "matchCount": { "type": "integer" },
        "matches": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["category", "label"],
            "properties": {
              "category": { "type": "string" },
              "label": { "type": "string" },
              "severity": { "type": "string" },
              "matchedValue": { "type": "string" },
              "scope": { "type": "string" }
            }
          }
        }
      }
    },
    "ip": {
      "type": "object",
      "properties": {
        "address": { "type": "string" },
        "version": { "type": "integer" },
        "isPublic": { "type": "boolean" },
        "reverseDns": { "type": "array", "items": { "type": "string" } },
        "announced": { "type": "boolean" },
        "prefix": { "type": "string" },
        "asns": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["number"],
            "properties": {
              "number": { "type": "integer" },
              "holder": { "type": "string" }
            }
          }
        },
        "countryCode": { "type": "string" },
        "city": { "type": "string" },
        "registry": { "type": "string" },
        "cidrs": { "type": "array", "items": { "type": "string" } }
      }
    },
    "asn": {
      "type": "object",
      "properties": {
        "number": { "type": "integer" },
        "label": { "type": "string" },
        "holder": { "type": "string" },
        "announced": { "type": "boolean" },
        "prefixCount": { "type": "integer" },
        "prefixes": { "type": "array", "items": { "type": "string" } },
        "upstreamCount": { "type": "integer" },
        "downstreamCount": { "type": "integer" },
        "neighbours": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "number": { "type": "integer" },
              "relationship": { "type": "string" }
            }
          }
        }
      }
    },
    "tld": {
      "type": "object",
      "properties": {
        "tld": { "type": "string" },
        "addresses": { "type": "array", "items": { "type": "string" } },
        "nameservers": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["host"],
            "properties": {
              "host": { "type": "string" },
              "addresses": { "type": "array", "items": { "type": "string" } }
            }
          }
        }
      }
    },
    "feed": {
      "type": "object",
      "properties": {
        "title": { "type": "string" },
        "lastBuildDate": { "type": "string" },
        "itemCount": { "type": "integer" },
        "items": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["title", "link"],
            "properties": {
              "title": { "type": "string" },
              "link": { "type": "string" },
              "description": { "type": "string" },
              "pubDate": { "type": "string" },
              "category": { "type": "string" }
            }
          }
        }
      }
    },
    "resolvers": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "name", "status", "blocked"],
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "role": { "type": "string" },
          "status": { "type": "string", "enum": ["ok", "error"] },
          "blocked": { "type": "boolean" },
          "records": { "type": "array", "items": { "type": "string" } }
        }
      }
    },
    "discovery": {
      "type": "object",
      "required": ["subdomains", "related", "ips"],
      "properties": {
        "subdomains": { "type": "array", "items": { "$ref": "#/$defs/discovered" } },
        "related": { "type": "array", "items": { "$ref": "#/$defs/discovered" } },
        "ips": { "type": "array", "items": { "$ref": "#/$defs/discovered" } },
        "emails": { "type": "array", "items": { "$ref": "#/$defs/discovered" } },
        "phones": { "type": "array", "items": { "$ref": "#/$defs/discovered" } },
        "organizations": { "type": "array", "items": { "$ref": "#/$defs/discovered" } }
      }
    },
    "sources": {
      "type": "array",
      "description": "Her veri kaynağının durumu. Kısmi hatalarda rapor yine üretilir.",
      "items": {
        "type": "object",
        "required": ["id", "name", "status", "latencyMs"],
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "status": { "type": "string", "enum": ["ok", "empty", "throttled", "error", "skipped"] },
          "error": { "type": "string" },
          "latencyMs": { "type": "integer" }
        }
      }
    },
    "warnings": { "type": "array", "items": { "type": "string" } }
  },
  "$defs": {
    "identity": {
      "type": "object",
      "properties": {
        "commonNames": { "type": "array", "items": { "type": "string" } },
        "organizations": { "type": "array", "items": { "type": "string" } },
        "organizationalUnits": { "type": "array", "items": { "type": "string" } },
        "locations": { "type": "array", "items": { "type": "string" } }
      }
    },
    "ranking": {
      "type": "object",
      "required": ["domain", "provider", "listed"],
      "properties": {
        "domain": { "type": "string" },
        "provider": { "type": "string", "enum": ["cisco", "tranco"] },
        "currentDate": { "type": "string" },
        "previousDate": { "type": "string" },
        "currentRank": { "type": ["integer", "null"] },
        "previousRank": { "type": ["integer", "null"] },
        "change": { "type": ["integer", "null"] },
        "listed": { "type": "boolean" }
      }
    },
    "discovered": {
      "type": "object",
      "required": ["value", "sources"],
      "properties": {
        "value": { "type": "string" },
        "sources": { "type": "array", "items": { "type": "string" } }
      }
    }
  }
}`
