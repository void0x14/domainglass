# Kullanım Kılavuzu

`domainglass` aracının tam kullanım başvurusu.

## İçindekiler

1. [Temel kavramlar](#temel-kavramlar)
2. [Komutlar](#komutlar)
3. [Bayraklar](#bayraklar)
4. [Çıktı biçimleri](#çıktı-biçimleri)
5. [Hedef türleri](#hedef-türleri)
6. [Bağımsız çözümleyiciler](#bağımsız-çözümleyiciler)
7. [Toplu tarama](#toplu-tarama)
8. [Boru hattı tarifleri](#boru-hattı-tarifleri)
9. [Hız sınırı ve dayanıklılık](#hız-sınırı-ve-dayanıklılık)
10. [Sorun giderme](#sorun-giderme)

---

## Temel kavramlar

### Veri ve günlük ayrımı

- **stdout** yalnızca veri taşır (JSON, NDJSON, satır listesi).
- **stderr** ilerleme ve hata mesajları taşır.

Bu ayrım sayesinde boru hatlarında veriyi bozmadan günlükleri ayırabilirsiniz:

```bash
domainglass -subs example.com 2>/dev/null | wc -l
```

### Çıkış kodları

| Kod | Ad | Anlam |
|---|---|---|
| 0 | `basarili` | Sorgu tamamlandı, veri mevcut |
| 1 | `hata` | Parametre, ağ veya beklenmeyen hata |
| 2 | `bulunamadi` | Hedef kayıtlı değil |
| 3 | `hiz_siniri` | Hiçbir kaynak veri döndürmedi, hız sınırı |
| 4 | `kullanim` | Geçersiz komut satırı kullanımı |

### Kaynak durumu

JSON raporundaki `sources` dizisi her veri kaynağının sonucunu bildirir:

| Durum | Anlam |
|---|---|
| `ok` | Kaynak başarılı |
| `empty` | Kaynak yanıt verdi, kayıt yok |
| `throttled` | Hız sınırı |
| `error` | Kaynak başarısız; `error` alanı nedenini taşır |
| `skipped` | Kaynak bu hedef için geçerli değil |

---

## Komutlar

### `domain <alan adı>`

Alan adı istihbaratı. Komut yazmadan da varsayılan olarak çalışır.

Topladığı kaynaklar:

- DNS kayıtları (A, AAAA, CNAME, MX, NS, TXT, CAA, SOA) ve DNSSEC durumu
- RDAP kayıt verisi (kayıtçı, tarihler, durumlar, ad sunucuları)
- WHOIS verisi (POST ile talep üzerine)
- Cisco Umbrella ve Tranco popülerlik sıralamaları
- TLS sertifikası ve SAN listesi
- Ana sayfa web profili (HTTP durumu, başlıklar, meta veriler)
- DNS filtre sağlayıcıları (Cloudflare, Quad9, AdGuard, CleanBrowsing, Control D)
- HaGeZi liste sınıflandırması
- IANA TLD kaydı
- Pasif keşif korelasyonu

### `ip <adres>`

IP adresi istihbaratı. Komut yazmadan da hedef IP ise otomatik çalışır.

- BGP anons durumu, prefix, Origin ASN ve organizasyon
- Coğrafi konum tahmini (BGP verisine dayalı)
- Ters DNS ve ileri doğrulama
- RIR RDAP tahsis bilgisi
- ASN komşuluk verisi (yukarı/aşağı akım)
- TLS sertifikası

### `asn <numara>`

ASN istihbaratı: sahip, prefix sayısı, komşuluk özeti ve öne çıkan komşular.

### `tld <tld>`

IANA TLD kaydı: kayıt defteri adresi, iletişim kişileri, yetkili ad sunucuları.

### `akis <ad>`

Popülerlik RSS akışları. Geçerli adlar:

| Ad | Kaynak | Açıklama |
|---|---|---|
| `trending` | Cisco Umbrella | Trend olan alan adları |
| `top-gainers` | Cisco Umbrella | En çok yükselenler |
| `newly-ranked` | Cisco Umbrella | Listeye yeni girenler |
| `trending-tranco` | Tranco | Trend olan alan adları |
| `top-gainers-tranco` | Tranco | En çok yükselenler |
| `newly-ranked-tranco` | Tranco | Listeye yeni girenler |

### `toplu`

stdin üzerinden çoklu hedef. Boş satırlar ve `#` ile başlayanlar atlanır.

### `saglik`

domain.glass servis erişimini sınar. Başarılıysa kod 0, değilse hata kodu döner.

### `sema`, `yetenek`, `kimlik`, `yardim`

Çevrimdışı çalışan bilgi komutları. Ajanlar için önerilen giriş noktası
`yetenek` komutudur.

---

## Bayraklar

### Çıktı biçimi

| Bayrak | Açıklama |
|---|---|
| `-json` | Okunabilir JSON |
| `-ndjson` | Tek satır JSON |
| `-girintisiz` | JSON girintisiz |
| `-alan <liste>` | JSON alan seçimi (virgülle) |
| `-subs` | Hedefin verisinden çıkan alt alan adları, satır satır |
| `-ips` | IP adresleri, satır satır |
| `-related` | İlişkili alan adları, satır satır |
| `-emails` | E-postalar, satır satır |
| `-orgs` | Kurum adları, satır satır |

### Ağ ve davranış

| Bayrak | Varsayılan | Açıklama |
|---|---|---|
| `-timeout <sn>` | 20 | İstek zaman aşımı |
| `-hiz <ms>` | 1100 | İstekler arası asgari aralık |
| `-sessiz` | kapalı | İlerleme günlüklerini bastır |
| `-renk <kip>` | `auto` | Renk kipi: `auto`, `always`, `never` |
| `-cozumleyici <idler>` | kapalı | Bağımsız DNS filtre çözümleyicileri |

---

## Çıktı biçimleri

### İnsan çıktısı (varsayılan)

Renkli, bölümlere ayrılmış Türkçe rapor. Yalnızca uçbirim (TTY) algılandığında
renk uygulanır; boru hattında renk otomatik kapanır.

### JSON

```bash
domainglass -json example.com
domainglass -json example.com | jq .summary
domainglass -json -alan summary,discovery example.com
```

Bilinmeyen alan adları sessizce yok sayılmaz; stderr üzerinde uyarı ve geçerli
alan listesi yazılır.

### NDJSON

Her satırda tam bir JSON nesnesi. Toplu tarama ve akış işleme için:

```bash
cat hedefler.txt | domainglass toplu -ndjson | jq -c .target
```

### Satır çıktısı

Boru hattı araçlarına beslemek için ham liste. Bu liste pasif keşif sonucudur:
hedefin DNS kayıtları, sertifika SAN'ları, sayfası ve arama sonuçlarından çıkan
adlar. Aktif alt alan adı numaralandırması yapmaz; geniş kapsam için `subfinder`
gibi bir araçla birlikte kullan.

```bash
domainglass -subs example.com | httpx -silent
domainglass -ips example.com | sort -u
domainglass akis top-gainers -subs | head -100
```

---

## Hedef türleri

| Girdi | Sezilen tür | Örnek |
|---|---|---|
| Alan adı | domain | `example.com` |
| IPv4 | ip | `1.1.1.1` |
| IPv6 | ip | `2606:4700::1111` |
| ASN (komutlu) | asn | `domainglass asn 13335` |
| TLD (komutlu) | tld | `domainglass tld com` |
| Akış (komutlu) | feed | `domainglass akis top-gainers` |

Ne komut ne de geçerli hedef olan girdiler kullanım hatası (kod 4) verir.

---

## Bağımsız çözümleyiciler

`-cozumleyici` bayrağı, hedefi seçilen DNS filtrelerine doğrudan sorar. Bu, yerel
DNS yapılandırmasından bağımsız gerçek engelleme durumunu gösterir.

```bash
# Tüm filtreleri dene
domainglass -cozumleyici quad9,adguard,cleanbrowsing,controld,cloudflare-guvenlik example.com

# Yalnızca engel durumunu al
domainglass -cozumleyici quad9 -json supheli.com | jq '.resolvers[0].blocked'
```

Filtre sağlayıcıları engellenen adları `0.0.0.0` veya `::` adresine yönlendirir;
araç bunu `blocked: true` olarak işaretler.

---

## Toplu tarama

### Temel kullanım

```bash
cat hedefler.txt | domainglass toplu -ndjson > sonuclar.ndjson
```

### Girdi biçimi

Her satırda bir hedef. Boş satırlar ve `#` ile başlayanlar atlanır:

```text
# bu bir yorum satırı
example.com
1.1.1.1

tesla.com
```

### Hız kontrolü

Toplu taramada hız sınırı kritiktir. Varsayılan 1100 ms aralıkla 100 hedef
yaklaşık 6-8 dakika sürer. Hızı artırmak 429/403 kapılarını tetikler:

```bash
# Güvenli (varsayılan)
cat hedefler.txt | domainglass toplu -ndjson

# Daha hızlı ama riskli
cat hedefler.txt | domainglass toplu -ndjson -hiz 600
```

---

## Boru hattı tarifleri

```bash
# Alt alan adı keşfi -> canlılık ve durum kodu
domainglass -subs example.com | httpx -silent -status-code -title

# Kayıtlı alanları süz
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null \
  | jq -c 'select(.summary.registered == true) | .target'

# Güvenlik bayraklı alanları bul
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null \
  | jq -c 'select(.safety.flagged == true) | .target'

# 30 günden az kalan sertifikalar
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null \
  | jq -c 'select(.summary.tlsDaysRemaining < 30) | {target, days: .summary.tlsDaysRemaining}'

# IP -> ASN -> komşuluk zinciri
domainglass ip 1.1.1.1 -json | jq -r '.ip.asns[0].number' \
  | xargs -I{} domainglass asn {} -json

# Tüm keşfedilen varlıkları tek listede topla
domainglass -json example.com | jq -r '.discovery | (.subdomains + .related + .ips)[] | .value' | sort -u

# En çok yükselen alanlara hızlı bakış
domainglass akis top-gainers -subs | head -50
```

---

## Hız sınırı ve dayanıklılık

domain.glass uç noktaları hızlı ardışık isteklerde 429 veya 403 döndürür. Araç:

1. İstekler arasında `-hiz` ile belirlenen asgari aralığı uygular (varsayılan 1100 ms).
2. Geçici hatalarda (429, 403 kapı kodları, 5xx) üstel geri çekilme + rastgele
   gecikme ile 3 kez yeniden dener.
3. Kalıcı hatalarda (400 doğrulama, 404) yeniden denemez.
4. Başarısız kaynakları raporun `sources` bölümünde işaretler; kalan veriyi
   yine üretir.

---

## Sorun giderme

### `hiz_siniri` (kod 3) alıyorum

`-hiz` değerini artırın (ör. `-hiz 2500`) ve biraz bekleyip tekrar deneyin.

### Rapor bazı kaynaklarda `error` gösteriyor

`sources` dizisindeki `error` alanı nedeni taşır. TLS, web profili, WHOIS ve
sınıflandırma uç noktaları site tarafında "eşleşen sayfadan gel" kapısıyla
korunur; hız sınırına takıldıklarında düşebilirler.

### JSON çıktısında alan eksik

İlgili kaynak `sources` içinde `error` veya `throttled` olabilir. `summary`
bölümü yalnızca mevcut veriyi yansıtır.

### Renkler bozuk görünüyor

`-renk never` ile renkleri kapatın veya `-renk always` ile zorlayın.

### Boş sonuç alıyorum ama hata yok

Kod 2 (`bulunamadi`) döndüyse hedef kayıtlı değildir. Bu geçerli bir sonuçtur,
hata değildir.

