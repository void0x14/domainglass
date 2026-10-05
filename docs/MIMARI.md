# Mimari

## Genel bakış

`domainglass`, domain.glass sitesinin kendi uç noktalarını kullanan bir Go
CLI aracıdır. Tek bir ikili dosya üretir; harici Go bağımlılığı yoktur.

```text
cmd/domainglass/main.go   ->  ince kabuk; cli.Calistir çağırır
internal/cli/             ->  bayrak çözümleme, komut yönlendirme, çıktı, çıkış kodları
internal/dg/              ->  HTTP taşıma: başlıklar, hız sınırı, geri çekilme, tipli hata
internal/dns/             ->  RFC 1035 kodlayıcı/çözücü + DoH istemcisi
internal/feeds/           ->  RSS akış çözümleyici
internal/model/           ->  ham API tipleri ve kararlı rapor şeması
internal/recon/           ->  paralel toplama ve pasif keşif korelasyonu
internal/out/             ->  insan / JSON / NDJSON çıktı üreticileri
```

## Katmanlar ve sorumluluklar

### internal/dg — taşıma

Sitenin beklediği başlıkları uygular:

| Başlık | Değer |
|---|---|
| User-Agent | Gerçekçi Chrome UA |
| Accept | application/json |
| Sec-Fetch-Site | same-origin |
| Sec-Fetch-Mode | cors |
| Sec-Fetch-Dest | empty |
| X-Domain-Glass-Action | Uç noktaya özgü eylem |
| Referer / Origin | Eşleşen sayfa URL'i |

Bunlara ek olarak:

- **Hız sınırı:** İstekler mutex ile sıraya sokulur ve asgari aralık uygulanır.
- **Geri çekilme:** Geçici hatalarda üstel gecikme + jitter ile 3 deneme.
- **Tipli hata:** Hata{Tur, HTTPKod, Kod, Mesaj} — çağıran taraf sınıfa göre karar verir.

Hata sınıfları:

| Sınıf | Kaynak | Yeniden dener |
|---|---|---|
| HataGecici | 429, 403 kapı kodları, 5xx | Evet |
| HataYok | 404 | Hayır |
| HataGecersiz | 400, 405, 422 | Hayır |
| HataErisim | Diğer 4xx | Hayır |
| HataAg | Ağ/zaman aşımı | Evet |
| HataBicim | JSON çözümleme | Hayır |

### internal/dns — bağımsız DNS

İki katman:

1. **Tel formatı** (wire.go): RFC 1035 sorgu kodlama ve yanıt çözme. Sıkıştırma
   işaretçileri, MX/SOA/TXT/CAA RDATA çözümü dahil.
2. **Çözümleyici** (resolver.go): Sekiz sağlayıcıya sorgu.

| Sağlayıcı | Taşıma |
|---|---|
| sistem | net.Resolver |
| cloudflare, google, cloudflare-* | JSON DoH |
| quad9, adguard, cleanbrowsing, controld | İkili DoH (application/dns-message) |

İkili DoH gerekir çünkü Quad9, AdGuard, CleanBrowsing ve Control D JSON tabanlı
DoH sunmaz; yalnızca application/dns-message kabul ederler. CleanBrowsing uç
noktası doh.cleanbrowsing.org/doh/security-filter/ yolundadır.

### internal/recon — toplama ve korelasyon

Topla fonksiyonu hedef türüne göre ilgili kaynakları **paralel** çalıştırır:

- Alan adı için 10 kaynak: domain, cisco, tranco, safety, classifications, tls,
  web-profile, search, whois, tld-iana
- IP için: ip, tls-ip, classifications, ardından ASN komşuluk sorguları

Her kaynak kendi goroutine'inde çalışır; sonuçlar mutex ile rapora yazılır. Tek
bir kaynak hata verse bile rapor üretilir ve hata sources içinde işaretlenir.

Kesif fonksiyonu şu kaynaklardan korelasyon yapar: DNS kayıtları, RDAP ad
sunucuları, TLS SAN listesi, web profili URL'leri ve başlıkları, arama sonuçları,
ters DNS, WHOIS ad sunucuları.

### internal/model — tipler

İki grup:

- **Ham tipler** (ham.go): API sözleşmesini birebir yansıtır.
- **Rapor tipleri** (rapor.go): Kullanıcıya ve ajanlara sunulan kararlı şema.

Ham tipler json:"-" ile rapora yazılmaz; yalnızca dönüşüm ve korelasyon için
tutulur.

### internal/out — çıktı

- Yaz: Türkçe insan çıktısı, TTY algılamalı renk.
- JSON: Girintili JSON.
- NDJSON: Tek satır JSON.
- AlanSecimi: Üst düzey alan süzme.

JSON anahtarları İngilizcedir (makine sözleşmesi); insan çıktısı tamamen
Türkçedir.

## Veri akışı

```text
argv -> cli.Cozumle -> cli.Calistir
                         |
                         +- domain/ip -> recon.Topla
                         |                 |
                         |                 +- dg.Client.* (paralel, hız sınırlı)
                         |                 +- recon.isle (ham -> rapor)
                         |                 +- recon.Kesif (korelasyon)
                         |                 +- dns.Cozumleyici.* (isteğe bağlı)
                         +- asn  -> dg.ASN
                         +- tld  -> dg.TLDiana
                         +- akis -> dg.Akis -> feeds.Cozumle
                         +- toplu -> satır satır recon.Topla
                         v
                    out.{Yaz,JSON,NDJSON,AlanSecimi} -> stdout
                    hata/uyarı -> stderr
```

## Tasarım kararları

### Neden internal/ ?

Paketler uygulamaya özeldir; dışarıya kararlı bir Go API'si sunma taahhüdü
yoktur. Ajanlar CLI üzerinden JSON sözleşmesini kullanır.

### Neden harici bağımlılık yok?

Tek ikili dosya, çapraz derleme kolaylığı ve tedarik zinciri riskinin sıfıra
yakın olması. DNS kodlayıcı, RSS çözümleyici ve hız sınırlayıcı standart
kütüphaneyle yazıldı.

### Neden rapor şeması İngilizce anahtarlı?

JSON anahtarları bir makine sözleşmesidir; ekosistem araçları (jq, httpx, ajan
SDK'ları) İngilizce anahtar bekler. İnsan çıktısı ve açıklamalar Türkçedir.

### Neden kısmi başarı?

domain.glass uç noktalarının bir kısmı hız sınırına takılabilir. Tüm raporu
başarısız saymak yerine, elde edilen veriyi sunup eksik kaynakları işaretlemek
daha kullanışlıdır. sources bölümü bu sözleşmeyi taşır.

## Test stratejisi

| Paket | Kapsam |
|---|---|
| internal/dns | Sorgu kodlama (hex karşılaştırma), yanıt çözme, NXDOMAIN, engel tespiti |
| internal/dg | Başlık sözleşmesi, POST WHOIS, 429/403 geri çekilme, kalıcı hata, hız sınırı |
| internal/feeds | RSS çözümleme, host çıkarma, bozuk XML |
| internal/recon | Host normalizasyonu, keşif korelasyonu, kaynak etiketleri |
| internal/out | JSON/NDJSON geçerliliği, alan seçimi, insan çıktısı, renk |
| internal/cli | Bayrak çözümleme, katalog, şema, çıkış kodları, çakışma denetimleri |

Ağ gerektiren uçtan uca test DOMAINGLASS_CANLI_TEST=1 ile çalışır.

