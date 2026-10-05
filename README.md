# domainglass

> **domain.glass** servisini en hızlı, tam yetenekli ve efektif şekilde terminale taşıyan, yapay zeka (AI) ajanları ve güvenlik araştırmacıları için optimize edilmiş Go aracı.

`domainglass`, web tarayıcısına ihtiyaç duymadan `domain.glass` platformunun sağladığı tüm DNS, RDAP, DNSSEC, Cisco/Tranco popülerlik sıralamaları, tehdit filtreleri (Cloudflare, Quad9, AdGuard vb.), TLS sertifika SAN analizleri, IP/ASN yönlendirme istihbaratı ve pasif altyapı keşiflerini tek bir ikili dosyada (`binary`) sunar.

---

## Özellikler

- **DNS & DNSSEC İstihbaratı:** A, AAAA, CNAME, MX, NS, TXT, SOA kayıtları, TTL süreleri ve resolver DNSSEC doğrulama durumu.
- **RDAP & Kayıt Bilgisi:** Yetkili yazman (Registrar), kayıt / bitiş / değişiklik tarihleri, authoritative NameServer listesi.
- **Popülerlik Sıralamaları:** Günlük Cisco Umbrella ve Tranco Top 1M derecelendirmesi ve günlük sıra değişimi (`change`).
- **Güvenlik ve Tehdit Filtreleri:** Cloudflare Malware/Family, Quad9, AdGuard DNS, CleanBrowsing ve Control D güvenlik/oltalama/reklam engelleme kontrolleri.
- **TLS Sertifika & SAN Keşfi:** Canlı port 443 SNI incelemesi, yayımcı, kalan gün sayısı ve tüm SAN (Subject Alternative Names) alan adları.
- **Pasif Altyapı Keşfi (`Discovered Hosts`):** DNS yanıtları, TLS SAN kayıtları ve web profili üzerinden hedefe ait alt alan adları, ilişkili alan adları ve IP adreslerinin otomatik korelasyonu.
- **IP & ASN Yönlendirme:** IP sorgularında BGP anons durumu, prefix, Origin ASN ve organizasyon bilgisi, yaklaşık coğrafi konum ve RIR RDAP tahsisi.
- **Agentic & Boru Hattı (Pipeline) Desteği:** 
  - Standart unix filtreleri (`-subs`, `-ips`, `-related`) sayesinde doğrudan `httpx`, `subfinder` veya diğer araçlara beslenebilir.
  - `--schema` bayrağı ile AI ajanları aracı sıfır sürtünmeyle çağırabilir.
  - `--json` bayrağı ile tam yapılandırılmış çıktı üretir.
  - `stdin` üzerinden toplu liste kabul eder.

---

## Kurulum

### Kaynak Koddan Derleme
```bash
git clone https://github.com/void0x14/domainglass.git
cd domainglass
go build -o domainglass cmd/domainglass/main.go
sudo mv domainglass /usr/local/bin/
```

---

## Kullanım

### 1. Temel Domain İstihbaratı
```bash
domainglass example.com
```

### 2. IP Adresi ve ASN Analizi
```bash
domainglass 1.1.1.1
```

### 3. Alt Alan Adlarını Satır Satır Çekme (Pipeline)
```bash
domainglass -subs tesla.com | httpx -silent
```

### 4. İlişkili IP Adreslerini Çekme
```bash
domainglass -ips tesla.com
```

### 5. Yapay Zeka Ajanları ve Otomasyon İçin JSON Çıktısı
```bash
domainglass -json tesla.com > tesla_recon.json
```

### 6. Toplu Alan Adı Tarama (Stdin)
```bash
cat domains.txt | domainglass -subs > all_subdomains.txt
```

### 7. AI Ajanları İçin JSON Şeması
```bash
domainglass -schema
```

---

## Seçenekler (CLI Flags)

| Bayrak | Açıklama |
|---|---|
| `<hedef>` | İncelenecek alan adı veya IP adresi |
| `-subs` | Yalnızca keşfedilen alt alan adlarını (subdomains) yazdırır |
| `-ips` | Yalnızca tespit edilen IP adreslerini yazdırır |
| `-related` | Yalnızca ilişkili kök alan adlarını yazdırır |
| `-json` | Tüm istihbarat raporunu saf JSON formatında döndürür |
| `-schema` | Agentic araç entegrasyonları için JSON şemasını basar |
| `-timeout` | İstek zaman aşımı süresi saniye cinsinden (varsayılan: 15) |
| `-version` | domainglass sürümünü gösterir |

---

## Lisans

Bu proje **GNU Affero General Public License v3.0 (AGPL-3.0)** altında lisanslanmıştır. Detaylar için [LICENSE](LICENSE) dosyasına bakabilirsiniz.
