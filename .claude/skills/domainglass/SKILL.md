---
name: domainglass
description: Use when a task needs domain, subdomain, IP, ASN, TLD or DNS intelligence (passive recon, certificate names, registration data, popularity ranks, blocklist status). Covers running the domainglass CLI and parsing its JSON or NDJSON output.
---

# domainglass

`domainglass`, domain.glass istihbaratını terminale taşıyan tek ikili Go aracıdır.
Alan adı, IP, ASN, TLD ve popülerlik akışlarını tek komutta toplar.

## Önce kataloğu oku

Komutları tahmin etme. Araç kendini tanımlar:

```bash
domainglass yetenek   # komutlar, bayraklar, çıkış kodları, çözümleyiciler (JSON)
domainglass sema      # rapor JSON şeması
```

## Temel kullanım

```bash
domainglass -json -sessiz <hedef>          # tam rapor (JSON)
domainglass -subs <hedef>                  # yalnızca alt alan adları
domainglass -ips <hedef>                   # yalnızca IP adresleri
domainglass -related <hedef>               # yalnızca ilişkili alan adları
domainglass ip <adres> -json               # IP: BGP/ASN, konum, ters DNS
domainglass asn <numara> -json             # ASN: sahip, prefix, komşuluk
domainglass tld <tld> -json                # IANA TLD kaydı
domainglass akis top-gainers -subs         # popülerlik akışı
cat hedefler.txt | domainglass toplu -ndjson  # toplu tarama
```

## Çıkış kodları

| Kod | Ad | Davranış |
|---|---|---|
| 0 | basarili | Sonucu kullan |
| 1 | hata | Hatayı bildir |
| 2 | bulunamadi | Kayıtlı değil - geçerli bulgu, hata değil |
| 3 | hiz_siniri | -hiz değerini artır, tekrar dene |
| 4 | kullanim | Komutu düzelt |

Veri stdout'ta, günlükler stderr'dedir. Boru hattında `2>/dev/null` kullan.

## Rapor yapısı

Zorunlu üst düzey alanlar: target, kind, tool, generatedAt, summary, discovery,
sources, warnings.

- `summary`: hızlı karar alanları (registered, ranks, tlsDaysRemaining, safetyFlags).
- `sources`: her kaynağın durumu (ok/empty/throttled/error/skipped).
  Bir kaynak düştüyse ilgili alan eksik olabilir; eksikliği temiz sayma.
- `discovery`: {value, sources} öğeleri; ipucu, sahiplik kanıtı değil.

## Engel durumu

```bash
domainglass -json -sessiz -cozumleyici quad9,adguard,cleanbrowsing <hedef> \
  | jq -c '.resolvers[] | {id, name, blocked}'
```

Filtre sağlayıcıları engellenen adı 0.0.0.0 / :: adresine yönlendirir; araç
bunu blocked=true olarak işaretler.

## Hız sınırı

Varsayılan 1100 ms aralık güvenlidir. Kod 3 alırsan -hiz değerini artır.
Toplu taramada eşzamanlı süreç sayısını 1-2 ile sınırla.

## Kaçınılması gerekenler

- -json ile -subs/-ips/-related birlikte kullanılamaz (kod 4).
- -json ile -ndjson birlikte kullanılamaz.
- -alan tek başına kullanılamaz.
- -hiz 0 kullanma; kaynaklar düşer.

## Ayrıntı

- AGENTS.md - ajan sözleşmesi
- docs/AI-AJANLARI.md - ayrıntılı kılavuz ve tarifler

