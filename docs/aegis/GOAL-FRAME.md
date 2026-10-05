# Aegis Hedef Çerçevesi — domainglass (yeniden inşa)

## TaskIntentDraft

- **Talep edilen sonuç:** domain.glass servisinin tüm istihbaratını terminale
  taşıyan; insan ve AI ajanlarının sıfır sürtünmeyle kullandığı, AGPL-3.0 lisanslı,
  Türkçe dokümantasyonlu Go CLI aracı.
- **Hedef:** DNS, RDAP, WHOIS, DNSSEC, TLS/SAN, popülerlik sıralamaları, DNS filtre
  sağlayıcıları, IP/ASN yönlendirme ve pasif keşif korelasyonunu tek ikili dosyada
  sunmak; buna ek olarak ASN, TLD ve RSS akış istihbaratı.
- **Başarı kanıtı:**
  1. `go vet ./...` ve `go test ./...` sıfır hata.
  2. `go build ./cmd/domainglass` başarılı; ikili klon üzerinden derlenebilir.
  3. Canlı komutlar doğrulandı: `example.com`, `ip 1.1.1.1`, `asn 13335`, `tld com`,
     `akis top-gainers`, `-subs tesla.com`, `toplu -ndjson`.
  4. `sema` ve `yetenek` çıktıları geçerli JSON.
  5. Çıkış kodları 0/1/2/3/4 canlı senaryolarla doğrulandı.
  6. GitHub deposu public, AGPL-3.0, tamamen Türkçe belgelerle yayında.
- **Durma şartı:** `done` — tüm kriterler komut çıktısı ve commit hash ile kanıtlandığında.
- **Kapsam dışı:** Web arayüzü, ücretli API entegrasyonu, üçüncü parti Go bağımlılığı.
- **Kısıtlar:** Harici Go bağımlılığı yok; site sözleşmesine uyan başlıklar;
  hız sınırına saygılı istek aralığı.
- **Aegis görünürlüğü:** TDD rotası `strict` (yeni CLI sözleşmesi ve çekirdek taşıma
  katmanı); plan `docs/aegis/plans/2026-10-05-domainglass-yeniden-insa.md`.

## Neden yeniden inşa

Önceki durum kanıtlandı:

1. `origin/main` ağacında `cmd/` dizini **yoktu**; `.gitignore` içindeki `domainglass`
   deseni `cmd/domainglass/main.go` dosyasını da yok sayıyordu (`git check-ignore`
   çıktısı). Depo klonlandığında derlenmiyordu.
2. README var olmayan `--schema` bayrağını belgeliyordu; gerçek bayrak `-schema` idi.
3. API uç noktaları eksik doğrulanmıştı: TLS, web profili, safety, WHOIS ve
   sınıflandırma uç noktaları "eşleşen sayfadan gel" kapısıyla korunuyor; IP ve ASN
   uç noktaları hiç kullanılmıyordu.
4. Hız sınırı davranışı yanlış modellenmişti: hızlı ardışık istekler 429/403 üretiyor.

## Durum

- Durum: `done` (aşağıdaki kanıtlarla).
- Doğrulama kanıtları: `go vet`, `go test`, canlı komut çıktıları, git commit.

