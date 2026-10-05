# AGENTS.md — Yapay Zeka Ajanları İçin Çalıştırma Sözleşmesi

Bu belge, `domainglass` aracını çağıracak ajanlar (LLM, otonom ajan, kod
asistanı) için bağlayıcı kuralları tanımlar. Buradaki her komut canlı ortamda
doğrulanmıştır.

## Altın kural

**Tahmin etme, kataloğu oku.** Araç kendi sözleşmesini makine tarafından
okunabilir biçimde yayınlar:

```bash
domainglass yetenek   # komutlar, bayraklar, çıkış kodları, çözümleyici listesi
domainglass sema      # rapor JSON şeması (JSON Schema 2020-12)
```

Bu iki komut her zaman çevrimdışı çalışır (ağ erişimi gerekmez) ve sürümle
birlikte sabittir.

## Çıktı kanalları

- **stdout**: yalnızca veri (JSON, NDJSON veya satır çıktısı).
- **stderr**: ilerleme ve hata günlükleri.

Boru hattında günlükleri bastırmak için `2>/dev/null` veya `-sessiz` kullanın.

## Bayrak seçim kılavuzu

| Görev | Doğru çağrı |
|---|---|
| Yapılandırılmış veri okumak | `domainglass -json <hedef>` |
| Akış hâlinde işlemek | `domainglass -ndjson <hedef>` |
| Yalnızca belirli alanlar | `domainglass -json -alan summary,discovery <hedef>` |
| Alt alan adı listesi | `domainglass -subs <hedef>` |
| IP listesi | `domainglass -ips <hedef>` |
| İlişkili alan adları | `domainglass -related <hedef>` |
| E-posta / kurum | `domainglass -emails` / `domainglass -orgs` |
| Engel durumu | `domainglass -cozumleyici quad9,adguard <hedef>` |
| Çoklu hedef | `cat liste.txt \| domainglass toplu -ndjson` |

## Çıkış kodları (kesin sözleşme)

| Kod | Ad | Ajan davranışı |
|---|---|---|
| 0 | `basarili` | Sonucu kullan |
| 1 | `hata` | Hatayı bildir; parametreleri veya ağı kontrol et |
| 2 | `bulunamadi` | Hedef kayıtlı değil; bu bir hata değil, sonuçtur |
| 3 | `hiz_siniri` | `-hiz` değerini artırıp yeniden dene |
| 4 | `kullanim` | Komutu düzelt; `domainglass yardim` oku |

`2` kodunu hata olarak ele almayın: NXDOMAIN veya kayıtsız alan adı geçerli bir
bulgudur.

## Rapor yapısı

Her JSON raporu şu üst düzey alanları taşır:

```json
{
  "target": "example.com",
  "kind": "domain",
  "tool": { "name": "domainglass", "version": "1.0.0" },
  "generatedAt": "2026-10-05T00:00:00Z",
  "summary": { "registered": true, "ciscoRank": 5098 },
  "discovery": { "subdomains": [], "related": [], "ips": [] },
  "sources": [ { "id": "domain", "name": "DNS, RDAP ve içerik güvenliği", "status": "ok", "latencyMs": 210 } ],
  "warnings": []
}
```

`sources` bölümü **her zaman** vardır ve her veri kaynağının durumunu bildirir.
Bir kaynak `error` veya `throttled` ise ilgili üst düzey alan eksik olabilir; bunu
"veri yok" olarak yorumlayın, "veri temiz" olarak değil.

## Ajan tarifleri

```bash
# Kayıtlı alan adlarını süz
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null \
  | jq -c 'select(.summary.registered == true) | .target'

# Güvenlik bayrağı olanları bul
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null \
  | jq -c 'select(.safety.flagged == true) | {target, flags: .summary.safetyFlags}'

# Süresi yaklaşan sertifikaları bul
cat hedefler.txt | domainglass toplu -ndjson 2>/dev/null \
  | jq -c 'select(.summary.tlsDaysRemaining != null and .summary.tlsDaysRemaining < 30) | .target'

# Alt alan adı topla ve canlılık kontrolüne ver
domainglass -subs example.com | httpx -silent

# Bağımsız engelleme kontrolü
domainglass -cozumleyici quad9,adguard,cleanbrowsing supheli.com -json \
  | jq '.resolvers[] | {name, blocked}'
```

## Kaynak güvenilirliği

- `sources[].status == "ok"` → veri taze ve geçerli.
- `sources[].status == "throttled"` → hız sınırı; `-hiz` artırıp tekrar deneyin.
- `sources[].status == "error"` → uç nokta başarısız; `error` alanı nedeni taşır.
- `sources[].status == "empty"` → uç nokta yanıt verdi ama kayıt yok.

## Yapılmaması gerekenler

- `sema` veya `yetenek` çıktısını ayrıştırmak için elle yazılmış regex kullanmayın;
  ikisi de geçerli JSON üretir.
- `-json` ile `-subs`/`-ips`/`-related`/`-emails`/`-orgs` bayraklarını birlikte
  kullanmayın; araç bunu kullanım hatası (kod 4) olarak reddeder.
- `-json` ile `-ndjson` birlikte kullanılamaz.
- `-alan` bayrağını tek başına kullanmayın; `-json` veya `-ndjson` gerekir.
- Hız sınırını `-hiz 0` ile kapatmayın; art arda istekler 429/403 kapısını
  tetikler ve kaynaklar düşer. Varsayılan 1100 ms güvenlidir.

## Ortam değişkenleri

Ajan konteynerlerinde sabit yapılandırma için:

| Değişken | Değer |
|---|---|
| `DOMAINGLASS_HIZ` | `1100` (ms) |
| `DOMAINGLASS_TIMEOUT` | `20` (sn) |
| `DOMAINGLASS_RENK` | `never` |
| `DOMAINGLASS_COZUMSYICI` | `quad9,adguard` |

## Örnek ajan döngüsü

```python
import json, subprocess

def istihbarat(hedef: str) -> dict:
    p = subprocess.run(
        ["domainglass", "-json", "-sessiz", "-hiz", "1100", hedef],
        capture_output=True, text=True, timeout=180,
    )
    if p.returncode == 2:
        return {"target": hedef, "registered": False}
    if p.returncode == 3:
        raise RuntimeError("hız sınırı; -hiz değerini artır")
    if p.returncode != 0:
        raise RuntimeError(p.stderr.strip())
    return json.loads(p.stdout)
```

## Kurulum

Aracı kurmak veya güncellemek için [KURULUM.md](KURULUM.md) dosyasını oku ve
adımları sırayla uygula. Kurulum tamamen o dosyadadır; README'nin insan
bölümlerini okumana gerek yoktur.

## İlgili belgeler

- [KURULUM.md](KURULUM.md) — ajana kurdurma ve devreye alma sözleşmesi
- [README.md](README.md) — genel bakış
- [docs/AI-AJANLARI.md](docs/AI-AJANLARI.md) — ayrıntılı ajan kılavuzu
- [docs/KULLANIM.md](docs/KULLANIM.md) — tüm bayraklar ve tarifler
- [docs/API-KAYNAK.md](docs/API-KAYNAK.md) — uç nokta sözleşmesi

