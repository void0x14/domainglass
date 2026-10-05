# Yapay Zeka Ajanları İçin domainglass

Bu belge, ajanların `domainglass` aracını sıfır sürtünmeyle kullanması için
gereken her şeyi içerir. Kısa sözleşme için [AGENTS.md](../AGENTS.md) dosyasına
bakın; bu belge ayrıntı ve tarifleri taşır.

## 1. Keşif: aracın kendini tanıtması

Ajanların tahmine ihtiyacı yok. Araç iki çevrimdışı komutla kendini tanımlar:

```bash
domainglass yetenek
domainglass sema
```

### yetenek çıktısı

```json
{
  "tool": "domainglass",
  "version": "1.0.0",
  "commands": [ { "name": "domain", "summary": "...", "example": "..." } ],
  "flags":    [ { "names": ["json"], "takesValue": false, "description": "..." } ],
  "exitCodes":[ { "code": 0, "name": "basarili", "description": "..." } ],
  "resolvers":[ { "id": "quad9", "name": "Quad9", "role": "malware" } ],
  "feedNames":[ "newly-ranked", "top-gainers", "trending" ]
}
```

### sema çıktısı

JSON Schema 2020-12 belgesi. Zorunlu üst düzey alanlar:

`target`, `kind`, `tool`, `generatedAt`, `summary`, `discovery`, `sources`, `warnings`

## 2. Çağrı sözleşmesi

### Temel çağrı

```bash
domainglass -json -sessiz -hiz 1100 <hedef>
```

| Öğe | Değer |
|---|---|
| Veri kanalı | stdout |
| Günlük kanalı | stderr |
| Çıktı biçimi | JSON (tek nesne) |
| Önerilen aralık | >=1100 ms |

### Çoklu hedef

```bash
printf '%s\n' hedef1 hedef2 | domainglass toplu -ndjson -sessiz
```

Her satırda tam bir rapor nesnesi. NDJSON akışı sayesinde tüm listeyi belleğe
almadan işleyebilirsiniz.

## 3. Çıkış kodu mantığı

```python
import json, subprocess

def sorgula(hedef):
    p = subprocess.run(
        ['domainglass', '-json', '-sessiz', hedef],
        capture_output=True, text=True, timeout=300,
    )
    if p.returncode == 0:
        return json.loads(p.stdout)
    if p.returncode == 2:   # kayıtlı değil: geçerli bulgu, hata değil
        return {'target': hedef, 'registered': False}
    if p.returncode == 3:   # hız sınırı: geri çekil ve tekrar dene
        raise TransientError(p.stderr)
    if p.returncode == 4:   # kullanım hatası: komutu düzelt
        raise ValueError(p.stderr)
    raise RuntimeError(p.stderr)
```

## 4. Rapor alanları

### summary — hızlı karar

| Alan | Tip | Anlam |
|---|---|---|
| registered | bool | Alan adı kayıtlı mı |
| ciscoRank | int/null | Cisco Umbrella sırası |
| trancoRank | int/null | Tranco sırası |
| dnssecSigned | bool | DNSSEC imzalı mı |
| tlsStatus | string | available / unavailable |
| tlsDaysRemaining | int/null | Sertifika kalan gün |
| safetyFlags | string[] | Bayraklı kategoriler |
| subdomainCount | int | Keşfedilen alt alan adı sayısı |
| relatedCount | int | İlişkili alan adı sayısı |
| ipCount | int | Bağlantılı IP sayısı |

### sources — güvenilirlik

Her kaynağın durumu. Bir kaynak error/throttled ise ilgili üst düzey alan eksik
olabilir. Eksik alanı "temiz" olarak yorumlamayın.

```python
def guvenilir(rapor, kaynak):
    for s in rapor.get('sources', []):
        if s['id'] == kaynak:
            return s['status'] in ('ok', 'empty')
    return False
```

### discovery — pasif keşif

Her öğe {value, sources} biçimindedir. sources alanı varlığın hangi kaynaktan
çıktığını bildirir (ör. DNS NS, TLS sertifikası, Ana sayfa URL). Bu, ortak
sahiplik kanıtı değildir; ipucu niteliğindedir.

## 5. Ajan tarifleri

### Kayıtlı ve kayıtsız alanları ayır

```bash
cat hedefler.txt | domainglass toplu -ndjson -sessiz 2>/dev/null \
  | jq -c 'select(.summary.registered == true) | .target'
```

### Güvenlik bayraklarını tara

```bash
cat hedefler.txt | domainglass toplu -ndjson -sessiz 2>/dev/null \
  | jq -c 'select(.safety.flagged == true) | {target, flags: .summary.safetyFlags}'
```

### Sertifika yenileme gerektirenleri bul

```bash
cat hedefler.txt | domainglass toplu -ndjson -sessiz 2>/dev/null \
  | jq -c 'select(.summary.tlsDaysRemaining != null and .summary.tlsDaysRemaining < 30)'
```

### Alt alan adı envanteri çıkar

```bash
domainglass -json -sessiz example.com | jq -r '.discovery.subdomains[].value' | sort -u
```

### Bağımsız engelleme doğrulaması

```bash
domainglass -json -sessiz -cozumleyici quad9,adguard,cleanbrowsing supheli.com \
  | jq -c '.resolvers[] | select(.blocked == true) | {id, name, role}'
```

### ASN zincirini takip et

```bash
domainglass ip 1.1.1.1 -json -sessiz | jq -r '.ip.asns[0].number' \
  | xargs -I{} domainglass asn {} -json -sessiz | jq '.asn | {label, holder, upstreamCount}'
```

### Akıştan hedef üret

```bash
domainglass akis top-gainers -sessiz -subs | head -50 \
  | domainglass toplu -ndjson -sessiz
```

## 6. Hız sınırı yönetimi

domain.glass hızlı ardışık isteklerde 429 veya 403 döndürür. Ajan tarafında:

1. Varsayılan 1100 ms aralığı koruyun.
2. Kod 3 (hiz_siniri) alırsanız -hiz değerini artırıp tekrar deneyin.
3. Toplu taramada eşzamanlı süreç sayısını sınırlayın (1-2 yeterlidir).

```bash
# Örnek: kademeli geri çekilme
for hiz in 1100 2500 5000; do
  if domainglass -json -sessiz -hiz $hiz example.com > out.json 2>/dev/null; then
    break
  fi
  sleep 15
done
```

## 7. Kaçınılması gerekenler

- -json ile -subs/-ips/-related/-emails/-orgs birlikte kullanılamaz (kod 4).
- -json ile -ndjson birlikte kullanılamaz (kod 4).
- -alan tek başına kullanılamaz; -json veya -ndjson gerekir.
- -hiz 0 ile hız sınırını kapatmayın; kaynaklar düşer ve kod 3 alırsınız.
- sema/yetenek çıktısını regex ile ayrıştırmayın; JSON ayrıştırıcı kullanın.
- stderr veri sanmayın; JSON yalnızca stdout üzerindedir.

## 8. Ortam değişkenleri

Konteyner veya CI ortamında sabit yapılandırma:

```bash
export DOMAINGLASS_HIZ=1100
export DOMAINGLASS_TIMEOUT=20
export DOMAINGLASS_RENK=never
export DOMAINGLASS_COZUMSYICI=quad9,adguard
```

## 9. Doğrulama listesi

Bir ajan entegrasyonunu bitirmeden önce:

- [ ] domainglass yetenek geçerli JSON döndürüyor.
- [ ] domainglass sema geçerli JSON Schema döndürüyor.
- [ ] Kod 0, 2, 3, 4 yolları ayrı ayrı ele alınmış.
- [ ] stdout yalnızca JSON olarak ayrıştırılıyor, stderr yok sayılıyor.
- [ ] sources dizisi okunuyor; eksik alan temiz sayılmıyor.
- [ ] Hız sınırı için geri çekilme var.
- [ ] Toplu tarama NDJSON ile satır satır işleniyor.

