# Commit Standardı

## Neden bu belge var

Bu depo sürüm numarası taşımaz ve tag yayınlamaz. Bir ikiliyi ayırt eden şey
commit hash'idir (`domainglass kimlik`).

Bunun doğrudan sonucu şudur: **commit mesajı, o değişikliğin tek sürüm notudur.**

Numaralı bir projede "1.4.2'de ne değişti" sorusunun cevabı ayrı bir CHANGELOG
dosyasındadır. Burada öyle bir dosya yok; cevap commit mesajının kendisidir.
Mesaj eksikse, hangi commit'i neden çekeceğini kimse anlayamaz. Bu yüzden biçim
serbest değil, denetlenen bir sözleşmedir.

## Denetim

Biçim `scripts/commit-denetle.py` ile denetlenir. Kurulum:

```bash
sh scripts/kur-kancalar.sh    # core.hooksPath=.githooks
```

Bu komuttan sonra `commit-msg` kancası her commit'i denetler; kurala uymayan mesaj
commit edilmez. Kancayı bilinçli olarak atlamak gerekirse:

```bash
git commit --no-verify    # gerekçesi mesaj gövdesinde belirtilmeli
```

Denetleyiciyi elle çalıştırmak için:

```bash
python3 scripts/commit-denetle.py --liste           # kural özeti
python3 scripts/commit-denetle.py --gecmis 10       # son 10 commit'i denetle
python3 scripts/commit-denetle.py --commit <ref>    # tek commit'i denetle
```

## Biçim

```text
<tip>(<kapsam>): <özet>

Neden:
<bu değişiklik neden gerekliydi; hangi sorun veya eksik tetikledi>

Ne:
- <somut değişiklik>
- <somut değişiklik>

Kanıt:
<çalıştırılan komut ve görülen sonuç>

Çekilebilir: evet|hayır
Kırıcı: yok|<ne bozuldu>
```

### Başlık

- En fazla 72 karakter.
- Tip küçük harfle yazılır, parantez içinde kapsam verilir.
- Özet küçük harfle başlar, nokta ile bitmez.
- Emir kipi veya durum bildirimi; "düzelttim" gibi birinci tekil şahıs kullanılmaz.

### Tipler

| Tip | Anlam |
|---|---|
| `feat` | Yeni yetenek veya davranış |
| `fix` | Hatalı davranışı düzeltme |
| `refactor` | Davranışı değiştirmeyen kod düzenlemesi |
| `docs` | Yalnızca belge değişikliği |
| `test` | Yalnızca test ekleme veya düzeltme |
| `chore` | Derleme, araç, bağımlılık gibi bakım işleri |
| `perf` | Ölçülebilir performans iyileştirmesi |
| `revert` | Önceki bir commit'i geri alma |

Yeni tip eklemek sözleşme değişikliğidir; denetleyicideki `TIPLER` sözlüğünü ve bu
belgeyi birlikte güncelle.

### Kapsam

Kapsam, değişikliğin hangi pakete veya alana dokunduğunu söyler. Yaygın değerler:

| Kapsam | Alan |
|---|---|
| `dg` | HTTP istemcisi (internal/dg) |
| `dns` | DNS tel formatı ve çözümleyici (internal/dns) |
| `recon` | Toplama ve keşif korelasyonu (internal/recon) |
| `out` | Çıktı üreticileri (internal/out) |
| `cli` | Komut satırı (internal/cli) |
| `model` | Veri tipleri (internal/model) |
| `build` | Derleme kimliği (internal/build) |
| `feeds` | RSS akışları (internal/feeds) |
| `docs` | Belgeler |
| `ci` | CI iş akışları |

Birden fazla alanı kapsayan değişiklikte en belirleyici olanı seç; ayrıntıyı `Ne`
bölümünde aç.

### Neden

En az 20 karakter. Şu soruları yanıtlar:

- Hangi sorun, eksik veya yanlış davranış vardı?
- Neden şimdi değiştirildi?
- Değiştirilmeseydi ne olurdu?

"Düzeltme yapıldı" gibi içi boş ifade kabul edilmez. Ölçüm, hata mesajı veya somut
gözlem beklenir.

### Ne

En az 10 karakter ve **madde listesi** olmalı (her satır `- ` ile başlar). Dosya
yolu veya davranış düzeyinde somut olmalı:

```text
Ne:
- internal/dg/istemci.go: 403 kapı kodlarını HataGecici sınıfına taşı
- internal/dg/istemci_test.go: yeniden deneme testi ekle
```

### Kanıt

En az 10 karakter. Çalıştırılan komut ve görülen sonuç. Test, derleme veya canlı
komut çıktısı olabilir:

```text
Kanıt:
go test ./internal/dg/ -run Gecici403 -v => PASS
Canlı: domainglass -hiz 1100 example.com => exit 0, 10 kaynak
```

"Testler geçti" gibi komutsuz ifade zayıftır; hangi komutun ne döndürdüğünü yaz.

### Çekilebilir

`evet` veya `hayır`. Bu etiket, hangi commit'i çekeceğine karar veren kişi için en
önemli alandır.

| Değer | Anlam |
|---|---|
| `evet` | Commit tek başına çekilip kullanılabilir; derlenir, testleri geçer, çalışır |
| `hayır` | Başka bir commit'e veya dış koşula bağımlı; gerekçesi `Ne` bölümünde olmalı |

`hayır` yazıyorsan `Ne` bölümünde bağımlılığı açıkça belirt. Örnek:

```text
Çekilebilir: hayır

Ne:
- Yalnızca docs/COMMIT-STANDART.md eklendi
- Denetleyici betiği ayrı bir commit'te gelecek; o gelmeden kanca çalışmaz
```

### Kırıcı

`yok` veya neyin bozulduğu. Geriye dönük uyumsuzluk varsa hangi yüzeyin
bozulduğunu yaz: CLI bayrağı, JSON alanı, çıkış kodu, dosya yolu, ortam değişkeni.

```text
Kırıcı: yok
```

```text
Kırıcı: tool.version alanı kaldırıldı; yerine tool.commit geldi. JSON tüketen
betikler güncellenmeli.
```

Başlıkta `!` işareti kırıcı değişikliği vurgular: `feat(cli)!: ...`

## Kapsam ve çekilebilirlik ilişkisi

Bu depoda tek commit'in çekilebilir olması tercih edilir. Bir değişikliği üç commit'e
bölmek yerine tek commit'te topla; ama commit yalnızca tek bir konuyu anlatsın.

| Durum | Doğru davranış |
|---|---|
| Tek konu, tek commit | Tercih edilen |
| İki ilgisiz konu | İki ayrı commit |
| Belge ve kod aynı konuya ait | Tek commit, `Ne` bölümünde ikisi de yazılır |
| Yalnızca belge | `docs:` tipi, `Çekilebilir: evet` |

## Geriye dönük düzeltme

Geçmişte atılmış ve kurala uymayan mesajları değiştirmek `git rebase` veya
`filter-repo` gerektirir ve paylaşılmış geçmişi bozar. Kural yürürlüğe girdikten
sonraki commit'ler için geçerlidir. Eski commit'lerin uygunluğunu görmek için:

```bash
python3 scripts/commit-denetle.py --gecmis 50
```

## Örnek

```text
fix(dg): hız sınırında yanlış hata sınıflandırması

Neden:
403 yanıtındaki browser_request_required kodu kalıcı erişim engeli olarak
sınıflandırılıyordu. Bu kod aslında hız sınırıyla birlikte görülüyor ve yeniden
denenebilir; yanlış sınıf yüzünden kaynaklar gereksizce düşüyordu.

Ne:
- internal/dg/istemci.go: 403 kapı kodlarını HataGecici sınıfına taşı
- internal/dg/istemci_test.go: yeniden deneme davranışı için test ekle

Kanıt:
go test ./internal/dg/ -run Gecici403 -v => PASS (3 deneme sonrası 200)
Canlı: domainglass -hiz 1100 example.com => exit 0, 10 kaynak

Çekilebilir: evet
Kırıcı: yok
```

## Yapay zeka ajanları için not

Commit atan bir ajan bu biçime uymak zorundadır. Kısa sözleşme [AGENTS.md](../AGENTS.md)
dosyasındadır. Kanca etkin değilse bile mesajı bu biçimde yaz; denetim
`--no-verify` ile atlansa dahi biçim korunmalıdır.

## İlgili belgeler

| Belge | İçerik |
|---|---|
| [AGENTS.md](../AGENTS.md) | Ajan çalıştırma sözleşmesi |
| [KURULUM.md](../KURULUM.md) | Kurulum ve devreye alma |
| [GELISTIRME.md](GELISTIRME.md) | Geliştirme akışı |

