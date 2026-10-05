# Uygulama Planı: domainglass (yeniden inşa)

## Hedef

`domain.glass` servisinin sunduğu tüm istihbaratı terminale taşıyan; hem insan
hem de yapay zeka ajanları tarafından sıfır sürtünmeyle kullanılabilen, AGPL-3.0
lisanslı, Türkçe arayüzlü Go CLI aracı.

## Mevcut Durum Kanıtı (bu plan öncesi)

- `origin/main` ağacında `cmd/` dizini **yok**; `.gitignore:1` içindeki `domainglass`
  deseni `cmd/domainglass/main.go` dosyasını da yok sayıyor (`git check-ignore` çıktısı).
- Klonlanan depo derlenmiyor: `stat cmd/domainglass: directory not found`.
- `README.md` var olmayan `--schema` bayrağını belgeliyor; gerçek bayrak `-schema`.
- API uç noktaları canlı olarak doğrulandı; TLS, web-profile, safety, WHOIS ve
  classifications uç noktaları "matching domain page" kapısı ile korunuyor.
- `domain.glass` uç noktaları hızlı ardışık istekte `429` döndürüyor (20 istekte 17 adet).

## Kapsam

**Dahil:** domain/subdomain/TLD, IP, ASN, RSS akışları, TLS, WHOIS, RDAP, DNS,
DNSSEC, sıralamalar, güvenlik filtreleri, HaGeZi sınıflandırması, pasif keşif
korelasyonu, JSON/NDJSON/human çıktı, alan seçimi, hız sınırlama, Türkçe dokümantasyon.

**Hariç:** web arayüzü, ücretli API entegrasyonu, üçüncü parti Go bağımlılığı.

## Mimari

```
cmd/domainglass/main.go     -> ince kabuk, cli.Run çağırır
internal/cli/               -> bayrak çözümleme, boru hattı, çıkış kodları
internal/dg/                -> HTTP taşıma katmanı, uç noktalar, tipli hatalar, throttle
internal/model/             -> ham API tipleri + rapor tipleri
internal/recon/             -> paralel toplama + keşif korelasyonu
internal/feeds/             -> RSS akış çözümleyici
internal/out/               -> human / json / ndjson / alan seçimi yazıcıları
```

## TDD Rotası

- Mod: `auto`
- Karar: `strict`
- Yetki: proje talimatı (`AEGIS-ROUTING`: davranış/uyarı/çekirdek değişiklikleri `auto` altında `strict` ister)
- Gerekçe: yeni genel API yüzeyi (CLI sözleşmesi), çekirdek taşıma katmanı, üretici/tüketici
  ilişkisi (JSON şeması ↔ çıktı), gerçek regresyon riski.
- Doğrulama: `go test ./...`, `go vet ./...`, canlı uçtan uca komut çıktıları.

## Görevler

1. **T1 — Modül ve taşıma katmanı:** `internal/dg` (başlıklar, POST WHOIS, 429/geri çekilme,
   tipli hatalar) + `internal/model` ham tipler. Kırmızı testler: başlık sözleşmesi,
   hata sınıflandırma, geri çekilme.
2. **T2 — Uç nokta metotları:** domain, tld, safety, rankings, classifications, tls,
   web-profile, whois, search, ip, asn, tld-iana, feeds.
3. **T3 — Keşif korelasyonu:** DNS/CNAME/MX/NS/RDAP/SAN/web-profile/search kaynaklarından
   subdomain, ilişkili alan adı, IP ayıklama; kaynak etiketli çıktı.
4. **T4 — CLI sözleşmesi:** komutlar (`domain`, `ip`, `asn`, `tld`, `akış`, `toplu`),
   çıkış kodları (`0/1/2/3/4`), `-json`, `-ndjson`, `-alan`, `-subs/-ips/-related`,
   `-paralel`, `-timeout`, `-sürüm`, `-şema`.
5. **T5 — Çıktı katmanı:** Türkçe insan çıktısı (renkli, TTY algılamalı), deterministik JSON,
   alan seçimi doğrulaması.
6. **T6 — Dokümantasyon:** `README.md`, `AGENTS.md`, `docs/KULLANIM.md`,
   `docs/AI-AJANLARI.md`, `docs/API-KAYNAK.md`, `docs/MIMARI.md`, `.skills/domainglass/SKILL.md`.
7. **T7 — Yayın:** `.gitignore` düzeltmesi, derleme, test, GitHub'a yeni kök commit ile gönderim.

## Doğrulama

- `go build ./...`, `go vet ./...`, `go test ./...`
- Canlı: `domainglass example.com`, `-json 1.1.1.1`, `asn 13335`, `-subs tesla.com`
- Şema geçerliliği: `-şema` çıktısı `python3 -m json.tool` ile ayrıştırılabilir.
- Ajan sözleşmesi: `AGENTS.md` içindeki her komut çalıştırılıp çıktısı doğrulanır.

## Emeklilik

Eski `pkg/` ağacı ve `domainglass` ikili dosyası kaldırılır; yerine `internal/` geçer.
`.gitignore` deseni kök ikiliye sabitlenir (`/domainglass`).

## Durma Şartı

`done`: tüm testler geçer, canlı komutlar doğrulanır, depo public/AGPL-3.0/Türkçe olarak yayınlanır.
