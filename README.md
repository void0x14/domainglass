# domainglass

**domain.glass** servisinin tüm istihbaratını terminale taşıyan Go aracı. İnsan
güvenlik araştırmacıları ve yapay zeka ajanları için tasarlandı.

Tek ikili dosya, sıfır üçüncü parti bağımlılık. Alan adı, IP, ASN ve TLD
istihbaratını; DNS/RDAP/WHOIS kayıtlarını, TLS sertifika keşfini, tehdit
filtrelerini ve pasif altyapı korelasyonunu tek komutta verir.

---

## Neden domainglass

- **Tek çağrı, tam rapor:** DNS, RDAP, WHOIS, TLS SAN, popülerlik sıralaması,
  DNS filtreleri ve web profili paralel toplanır.
- **Ajan dostu:** Kararlı JSON şeması (`domainglass sema`), makine kataloğu
  (`domainglass yetenek`), alan seçimi (`-alan`), NDJSON akışı ve anlamlı çıkış
  kodları.
- **Boru hattı dostu:** `-subs`, `-ips`, `-related`, `-emails`, `-orgs` bayrakları
  temiz satır çıktısı verir; `httpx`, `nuclei`, `subfinder` ile doğrudan zincirlenir.
- **Bağımsız doğrulama:** `-cozumleyici` ile Quad9, AdGuard, CleanBrowsing,
  Control D ve Cloudflare aile/güvenlik filtrelerine doğrudan DNS sorgusu yapılır.
- **Hız sınırına saygılı:** İstekler arasında ayarlanabilir asgari aralık ve
  üstel geri çekilme; domain.glass kapılarına takılmadan toplu tarama.

---

## Kurulum

### Yapay zeka ajanına kurdur (önerilen)

Ajanına aşağıdaki bloğu olduğu gibi ver. Ajan dosyayı okur, adımları sırayla
uygular, doğrular ve kontrol listesini raporlar. İnsan talimatı okumasına gerek
kalmaz.

```text
https://raw.githubusercontent.com/void0x14/domainglass/main/KURULUM.md adresini oku
ve içindeki adımları sırayla, kendin uygula. Kurulumu doğrulamadan "kuruldu" deme;
6. bölümdeki kontrol listesini tek tek çalıştırıp sonucu raporla.
```

Ajan dosyayı indiremiyorsa (ağ kısıtı), doğrudan oku:

```bash
curl -fsSL https://raw.githubusercontent.com/void0x14/domainglass/main/KURULUM.md
# veya depo klonlandıysa:
cat KURULUM.md
```

Kurulum dosyası; Go ön koşulunu, iki kurulum yolunu (go install / kaynak koddan
derleme), zorunlu doğrulama komutlarını, çıkış kodu sözleşmesini, ajan kullanım
kalıplarını, kısıtları ve sorun giderme adımlarını içerir.

### İnsan olarak kur

İnsanlar için iki yol var: Go ile derleme veya kaynak koddan derleme.

#### Go ile derleme

```bash
go install github.com/void0x14/domainglass/cmd/domainglass@latest
```

#### Kaynak koddan

```bash
git clone https://github.com/void0x14/domainglass.git
cd domainglass
go build -o domainglass ./cmd/domainglass
sudo install -m 0755 domainglass /usr/local/bin/
```

Go 1.21 veya üzeri gerekir. Harici Go bağımlılığı yoktur.

---

## Hızlı başlangıç

```bash
# Alan adı: DNS + RDAP + WHOIS + TLS + filtreler + keşif
domainglass example.com

# IP adresi: BGP/ASN + konum + ters DNS + TLS
domainglass ip 1.1.1.1

# ASN: sahip, prefix sayısı, komşuluk
domainglass asn 13335

# TLD: IANA kaydı
domainglass tld com

# Popülerlik akışı: en çok yükselenler
domainglass akis top-gainers

# Tam JSON raporu
domainglass -json example.com > rapor.json

# Yalnızca alt alan adları (boru hattı)
domainglass -subs tesla.com | httpx -silent

# Toplu tarama
cat hedefler.txt | domainglass toplu -ndjson > sonuclar.ndjson
```

---

## Komutlar

| Komut | Açıklama |
|---|---|
| `domain <ad>` | Alan adı istihbaratı (varsayılan; komut yazmadan da çalışır) |
| `ip <adres>` | IP adresi istihbaratı (varsayılan; komut yazmadan da çalışır) |
| `asn <numara>` | ASN istihbaratı |
| `tld <tld>` | IANA TLD kaydı |
| `akis <ad>` | RSS popülerlik akışı |
| `toplu` | stdin üzerinden çoklu hedef |
| `saglik` | Servis erişimini sınar |
| `sema` | Rapor JSON şeması |
| `yetenek` | Makine kataloğu (JSON) |
| `kimlik` | Bu derlemenin kaynak kimliği (commit) |

Hedef türü otomatik sezilir: `domainglass example.com` ile
`domainglass domain example.com` aynıdır.

---

## Bayraklar

| Bayrak | Açıklama |
|---|---|
| `-json` | Okunabilir JSON çıktısı |
| `-ndjson` | Tek satır JSON (akış için) |
| `-alan <liste>` | JSON çıktısında yalnızca seçilen üst düzey alanlar |
| `-subs` | Yalnızca alt alan adları (satır satır) |
| `-ips` | Yalnızca IP adresleri (satır satır) |
| `-related` | Yalnızca ilişkili alan adları (satır satır) |
| `-emails` | Yalnızca e-postalar (satır satır) |
| `-orgs` | Yalnızca kurum adları (satır satır) |
| `-cozumleyici <idler>` | Bağımsız DNS filtre çözümleyicileri (virgülle ayrılmış) |
| `-timeout <sn>` | İstek zaman aşımı (varsayılan 20) |
| `-hiz <ms>` | İstekler arası asgari aralık (varsayılan 1100) |
| `-renk <kip>` | Renk kipi: `auto`, `always`, `never` |
| `-sessiz` | İlerleme günlüklerini bastır |
| `-girintisiz` | JSON çıktısını girintisiz yaz |
| `-yardim`, `-kimlik`, `-sema`, `-yetenek` | Yardım, derleme kimliği, şema, katalog |

---

## Çıkış kodları

| Kod | Ad | Anlam |
|---|---|---|
| 0 | `basarili` | Sorgu tamamlandı, veri mevcut |
| 1 | `hata` | Parametre, ağ veya beklenmeyen hata |
| 2 | `bulunamadi` | Hedef kayıtlı değil (NXDOMAIN veya RDAP not_found) |
| 3 | `hiz_siniri` | Hız sınırı; hiçbir kaynak veri döndürmedi |
| 4 | `kullanim` | Geçersiz komut satırı kullanımı |

Veri **stdout**, günlük ve uyarılar **stderr** üzerindedir. Boru hatlarında
`2>/dev/null` ile günlükleri ayırabilirsiniz.

---

## Ortam değişkenleri

| Değişken | Etki |
|---|---|
| `DOMAINGLASS_API` | API kök adresi (varsayılan `https://domain.glass`) |
| `DOMAINGLASS_UA` | HTTP User-Agent |
| `DOMAINGLASS_HIZ` | Varsayılan istek aralığı (ms) |
| `DOMAINGLASS_TIMEOUT` | Varsayılan zaman aşımı (sn) |
| `DOMAINGLASS_RENK` | `auto` / `always` / `never` |
| `DOMAINGLASS_COZUMSYICI` | Varsayılan çözümleyici listesi |

---

## Bağımsız DNS çözümleyicileri

`-cozumleyici` bayrağı, verilen adı sekiz farklı DNS filtresine doğrudan sorar.
Kendi DNS yapılandırmanız yanlış olsa bile gerçek engelleme durumunu görürsünüz.

| ID | Sağlayıcı | Filtre rolü |
|---|---|---|
| `cloudflare` | Cloudflare 1.1.1.1 | — |
| `cloudflare-guvenlik` | Cloudflare for Families (Güvenlik) | malware |
| `cloudflare-aile` | Cloudflare for Families (Aile) | adult |
| `google` | Google Public DNS | — |
| `quad9` | Quad9 | malware |
| `adguard` | AdGuard DNS | advertising |
| `cleanbrowsing` | CleanBrowsing Güvenlik | adult |
| `controld` | Control D (Free) | advertising |

```bash
domainglass -cozumleyici quad9,adguard,cleanbrowsing example.com
domainglass -cozumleyici cloudflare-guvenlik supheli-alan.com -ips
```

Filtreler engellenen adları `0.0.0.0` veya `::` adresine yönlendirir; araç bunu
`ENGELLİ` olarak işaretler.

---

## Popülerlik akışları

```bash
domainglass akis trending            # Cisco Umbrella trend
domainglass akis top-gainers         # Cisco Umbrella en çok yükselenler
domainglass akis newly-ranked        # Cisco Umbrella yeni girenler
domainglass akis trending-tranco     # Tranco trend
domainglass akis top-gainers-tranco  # Tranco en çok yükselenler
domainglass akis newly-ranked-tranco # Tranco yeni girenler
```

---

## Örnek boru hatları

```bash
# Alt alan adı keşfi ve canlılık kontrolü
domainglass -subs tesla.com | httpx -silent -status-code

# İlişkili IP'leri topla
domainglass -ips tesla.com | sort -u > ips.txt

# Toplu tarama, yalnızca kayıtlı alanları süz
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null | jq -c 'select(.summary.registered==true) | .target'

# Yalnızca güvenlik bayraklı alanları süz
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null | jq -c 'select(.safety.flagged==true) | .target'

# IP -> ASN -> komşuluk zinciri
domainglass ip 1.1.1.1 -json | jq -r '.ip.asns[0].number' | xargs -I{} domainglass asn {} -json
```

---

## Yapay zeka ajanları için

Ajanlar tahmin etmek zorunda değil; aracın kendi kataloğunu okur:

```bash
domainglass yetenek   # komutlar, bayraklar, çıkış kodları, çözümleyiciler (JSON)
domainglass sema      # rapor JSON şeması (JSON Schema 2020-12)
```

- Ayrıntılı ajan sözleşmesi: [AGENTS.md](AGENTS.md)
- Ajan kullanım kılavuzu: [docs/AI-AJANLARI.md](docs/AI-AJANLARI.md)
- Claude Code skill: [.claude/skills/domainglass/SKILL.md](.claude/skills/domainglass/SKILL.md)

---

## Dokümantasyon

| Belge | İçerik |
|---|---|
| [docs/KULLANIM.md](docs/KULLANIM.md) | Ayrıntılı kullanım, tüm bayraklar, tarifler |
| [docs/AI-AJANLARI.md](docs/AI-AJANLARI.md) | Ajan entegrasyonu, JSON şeması, tarifler |
| [docs/MIMARI.md](docs/MIMARI.md) | Paket yapısı, veri akışı, tasarım kararları |
| [docs/API-KAYNAK.md](docs/API-KAYNAK.md) | domain.glass uç noktaları ve başlık sözleşmesi |
| [docs/GELISTIRME.md](docs/GELISTIRME.md) | Geliştirme, test, katkı |

---

## Sınırlar ve dürüst notlar

- domain.glass resmî bir API yayınlamaz; araç sitenin kendi uç noktalarını
  kullanır. Site sözleşmesi değişirse kaynaklar tek tek düşer ve rapor bunu
  `sources` bölümünde `error`/`throttled` olarak bildirir; sessizce yanlış veri
  üretmez.
- Bazı uç noktalar (TLS, web profili, WHOIS, sınıflandırma, güvenlik) site
  tarafında "eşleşen sayfadan gel" kapısıyla korunur. Araç, isteği site
  sözleşmesine uygun başlıklarla (Referer, Origin, X-Domain-Glass-Action) yapar;
  yine de hız sınırına takılırsa üstel geri çekilmeyle yeniden dener ve sonuçta
  ilgili kaynağı rapor içinde işaretler.
- IP konumu kaba tahmindir (BGP anons verisine dayanır), adres düzeyinde kesin
  konum değildir.
- TLS SAN listeleri ortak sertifikalarda ilgisiz adlar içerebilir; bu ortak
  sahiplik kanıtı değildir.

---

## Lisans

**GNU Affero General Public License v3.0 (AGPL-3.0)**. Ayrıntılar için
[LICENSE](LICENSE) dosyasına bakın.

