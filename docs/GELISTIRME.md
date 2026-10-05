# Geliştirme

## Gereksinimler

- Go 1.21 veya üzeri
- Harici bağımlılık yok

## Derleme ve test

```bash
go build -o domainglass ./cmd/domainglass
go vet ./...
go test ./...
```

Canlı ağ testi (gerçek domain.glass uç noktalarına bağlanır):

```bash
DOMAINGLASS_CANLI_TEST=1 go test ./internal/cli/ -run Canli -v
```

## Proje yapısı

```text
cmd/domainglass/     Giriş noktası (ince kabuk)
internal/cli/        Bayrak çözümleme, komutlar, çıkış kodları, şema, katalog
internal/dg/         HTTP istemcisi (başlıklar, hız sınırı, geri çekilme)
internal/dns/        DNS tel formatı + DoH çözümleyici
internal/feeds/      RSS çözümleyici
internal/model/      Ham API tipleri + rapor tipleri
internal/recon/      Paralel toplama + keşif korelasyonu
internal/out/        İnsan/JSON/NDJSON çıktı
```

## Kod kuralları

- **Türkçe kimlikler:** Değişken, fonksiyon, tip ve yorumlar Türkçedir.
  Yalnızca JSON anahtarları, CLI bayrakları ve üçüncü parti adlar İngilizcedir.
- **Yorum nedeni açıklar:** Ne yaptığını değil, neden gerekli olduğunu yazar.
- **Standart kütüphane önceliği:** Yeni bağımlılık eklemeden önce standart
  kütüphaneyle çözüm aranır.
- **Determinizm:** Aynı girdi aynı çıktıyı üretmeli; harita sıraları ve liste
  sıraları sabitlenir.

## Test yaklaşımı

| Katman | Yaklaşım |
|---|---|
| DNS tel formatı | Sabit hex iletimlerle kodlama/çözme doğrulaması |
| HTTP istemcisi | httptest sunucusu ile başlık, yeniden deneme, hata sınıfı |
| RSS | Sabit XML ile çözümleme |
| Keşif | Sabit API yanıtlarıyla korelasyon ve kaynak etiketi |
| Çıktı | JSON geçerliliği, alan seçimi, insan çıktısı içerik denetimi |
| CLI | Bayrak çözümleme, çakışma denetimleri, katalog/şema tutarlılığı |

Ağ erişimi gerektiren testler ortam değişkeniyle açılır; varsayılan olarak
atlanır.

## Katkı akışı

1. Değişikliği küçük ve tek amaçlı tutun.
2. Testi önce yazın; kırmızı olduğunu görün.
3. En küçük değişiklikle yeşile getirin.
4. gofmt, go vet ve go test temiz olmalı.
5. Dokümantasyonu etkilenen bölümle birlikte güncelleyin.

## Derleme kimliği

Bu araç sürüm numarası taşımaz. Bir ikilinin hangi koddan üretildiğini ayırt etmek
için kaynak kimliği kullanılır: git commit hash'i ve çalışma ağacının kirli olup
olmadığı.

```bash
domainglass kimlik
# örnek: 3dbbad9 · git · go1.21 · linux/amd64
# kirli ağaç: 3dbbad9+ · git · go1.21 · linux/amd64
```

Kimlik iki kaynaktan çıkarılır (internal/build paketi):

| Kaynak | Nasıl |
|---|---|
| Yerel git derlemesi | Go'nun gömdüğü `vcs.revision` / `vcs.modified` ayarları |
| Modül proxy (`go install`) | Go'nun atadığı sözde sürümdeki 12 haneli hex önek |

Hiçbiri yoksa kimlik `bilinmeyen` olur; uydurma değer üretilmez.

Kimlik hem insan başlığında, hem JSON raporundaki `tool` bölümünde, hem de
`yetenek` kataloğundaki `build` bölümünde görünür.

```json
"tool": {
  "name": "domainglass",
  "commit": "3dbbad9",
  "dirty": false,
  "go": "go1.21",
  "platform": "linux/amd64"
}
```

## Yeni uç nokta ekleme

1. internal/model/ham.go: yanıt tipini tanımla.
2. internal/dg/istemci.go: metodu ekle (doğru action ve Referer ile).
3. internal/recon/topla.go: paralel çağrıyı ekle.
4. internal/recon/isle.go: rapor dönüşümünü ekle.
5. internal/model/rapor.go: rapor tipini ekle (gerekirse).
6. internal/cli/sema.go: şemayı güncelle.
7. Testleri ekle.

