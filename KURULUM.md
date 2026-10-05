# KURULUM.md — Yapay Zeka Ajanı Kurulum ve Devreye Alma Sözleşmesi

> **Bu dosyayı okuyan ajan:** Aşağıdaki adımları sırayla, kendin uygula. Kullanıcıya
> kontrol listesi dökme. Sadece gerçekten senin yapamadığın bir şey varsa (ör. sudo
> parolası, ağ erişimi yok) tek cümleyle söyle.
>
> Hedef: makinede çalışan bir `domainglass` ikilisi ve doğrulanmış bir rapor.

---

## 0. Bu dosya nedir

`domainglass`, domain.glass istihbaratını terminale taşıyan tek ikili Go aracıdır.
Alan adı, IP, ASN, TLD ve popülerlik akışlarını toplar; JSON/NDJSON üretir.
Harici Go bağımlılığı yoktur. Depo: `https://github.com/void0x14/domainglass`

## 1. Katı kurallar (ihlal etme)

1. **Sırayla ilerle.** Bir adım başarısızsa durma; o adımın başarısızsa bölümünü
   uygula, sonra devam et.
2. **Doğrulamadan kuruldu deme.** Her adımın sonunda verilen doğrulama komutunu
   çalıştır ve çıktısını gör.
3. **stdout ve stderr'i karıştırma.** Araç veriyi stdout'a, günlükleri stderr'e yazar.
4. **Sessiz kur.** Gereksiz paket yöneticisi çağrısı, global dosya değişikliği veya
   `PATH` düzenlemesi yapma. Yalnızca bu dosyada yazan değişiklikleri yap.
5. **Kullanıcı dosyalarını ezme.** Mevcut bir dosyanın üzerine yazman gerekiyorsa
   önce yedekle, sonra yaz.
6. **Sır isteme.** Bu araç hiçbir API anahtarı, token veya hesap gerektirmez.
   Kimlik bilgisi soran bir adım yoktur.
7. **Tahmin etme.** Bayrak ve komutları bu dosyadan veya `domainglass yetenek`
   çıktısından al. README'nin insan bölümlerini okumak zorunda değilsin.
8. **Sürüm arama.** Bu araçta sürüm numarası yoktur. Bir derlemeyi ayırt etmek
   için `domainglass kimlik` çıktısındaki commit hash'ini kullan.

---

## 2. Ön koşul: Go araç zinciri

Araç Go 1.21 veya üzeri ile derlenir. Önce kontrol et:

```bash
go version
```

Beklenen: `go version go1.21` veya üzeri.

### Kurulu değilse veya 1.21'den eskiyse

Linux (paket yöneticisi):

```bash
# Debian/Ubuntu
sudo apt-get update && sudo apt-get install -y golang-go

# Fedora/RHEL
sudo dnf install -y golang

# macOS (Homebrew)
brew install go
```

Paket yöneticisindeki sürüm 1.21'den eskiyse resmî arşivden kur (mimariyi uname -m ile sez):

```bash
mim=$(uname -m)
case "$mim" in x86_64) a=amd64 ;; aarch64|arm64) a=arm64 ;; esac
curl -fsSL -o /tmp/go.tgz "https://go.dev/dl/go1.23.4.linux-$a.tar.gz"
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
export PATH="/usr/local/go/bin:$PATH"
go version
```

**Başarısızsa:** Ağ erişimi yoksa kullanıcıya tek cümleyle bildir: Go kurulamıyor,
önceden derlenmiş ikili gerekiyor. Devam etme.

---

## 3. Kurulum

İki yol var. **Yol A** tercih edilir. Başarısız olursa **Yol B**'ye geç.

### Yol A — go install (tercih edilen)

**DİKKAT:** bu depo tag yayınlamaz, sürüm numarası taşımaz. Bu yüzden `@latest`
güvenilir değildir: modül proxy'si `@latest` isteğini önbellekten eski bir commit
ile yanıtlar. Ölçülen örnek:

```text
proxy @latest        -> 307946d  (eski)
github main          -> 244ff26  (güncel)
```

İki çözüm var. **Birincisi kesin sonuç verir, onu kullan.**

#### Çözüm 1 (önerilen) — commit ile sabitle

```bash
# 1. Güncel commit'i öğren
COMMIT=$(git ls-remote https://github.com/void0x14/domainglass refs/heads/main | cut -f1)
echo "$COMMIT"

# 2. O commit'i kur
go install "github.com/void0x14/domainglass/cmd/domainglass@${COMMIT}"

# 3. Doğrula: çıktıdaki commit, 1. adımda aldığın değerle aynı olmalı
domainglass kimlik
```

#### Çözüm 2 — proxy'yi atla

Proxy önbelleğini devre dışı bırakıp doğrudan git'e git:

```bash
GOPROXY=direct go install github.com/void0x14/domainglass/cmd/domainglass@latest
domainglass kimlik
```

Bu yol git'in kurulu olmasını ve GitHub'a erişimi gerektirir; proxy'ye göre daha
yavaştır ama her zaman güncel kodu getirir.

#### Eski bir commit'i kurmak

Belirli bir geçmiş sürüme dönmek istersen aynı biçimde commit hash'i ver. Bu depo
commit geçmişini korur; eski commit'ler kurulabilir:

```bash
# Commit geçmişini listele
git ls-remote --tags https://github.com/void0x14/domainglass 2>/dev/null
# Tag yok; doğrudan klonlayıp geçmişe bak:
git clone --quiet https://github.com/void0x14/domainglass /tmp/dg-tarih
git -C /tmp/dg-tarih log --oneline --all

# Beğendiğin commit'i kur
go install "github.com/void0x14/domainglass/cmd/domainglass@<commit>"
domainglass kimlik    # kurduğun commit ile eşleşmeli
```

Doğrulanmış örnekler:

| Commit | Ne içerir | Kurulabilir mi |
|---|---|---|
| `c94e95a` | ilk sürüm denemesi | **Hayır** — `cmd/` dizini o commit'te yok |
| `307946d` | çalışan ilk araç | Evet — `domainglass surum` → `1.0.0` |
| `3dbbad9` | KURULUM.md eklendi | Evet — `domainglass surum` → `1.0.0` |
| `8b75be7` | sürüm kaldırıldı, kimlik geldi | Evet — `domainglass kimlik` |
| `244ff26` | go install düzeltmesi | Evet — güncel |

Not: `8b75be7` öncesi commit'lerde `kimlik` komutu yoktur, `surum` vardır.
`8b75be7` sonrasında tersi geçerlidir.

Kurulum dizinini PATH'e ekle (yalnızca eksikse):

```bash
GOBIN_DIR="$(go env GOPATH)/bin"
case ":$PATH:" in *":$GOBIN_DIR:"*) ;; *) export PATH="$GOBIN_DIR:$PATH" ;; esac
command -v domainglass
```

Kalıcı yapmak için kullanıcının kabuk profiline ekle:

```bash
grep -q 'GOPATH/bin' "$HOME/.zshrc" 2>/dev/null || echo 'export PATH="$HOME/go/bin:$PATH"' >> "$HOME/.zshrc"
```

**Başarısızsa** (ağ, proxy, modül erişimi): Yol B'ye geç.

### Yol B — kaynak koddan derleme

```bash
git clone --depth 1 https://github.com/void0x14/domainglass.git /tmp/domainglass-build
cd /tmp/domainglass-build
go build -trimpath -o domainglass ./cmd/domainglass
```

İkiliyi kalıcı bir yere taşı:

```bash
if [ -w /usr/local/bin ]; then
  install -m 0755 domainglass /usr/local/bin/domainglass
else
  mkdir -p "$HOME/.local/bin"
  install -m 0755 domainglass "$HOME/.local/bin/domainglass"
  case ":$PATH:" in *":$HOME/.local/bin:"*) ;; *) export PATH="$HOME/.local/bin:$PATH" ;; esac
fi
command -v domainglass
```

**Başarısızsa:** go build hatasını olduğu gibi kullanıcıya ilet. Devam etme.


---

## 4. Doğrulama (zorunlu)

Her komutu çalıştır ve beklenen sonucu gör.

### 4.1 İkili çalışıyor mu

```bash
domainglass kimlik
```
Beklenen: `<commit> · <kaynak> · <go sürümü> · <platform>` biçiminde tek satır.
Örnek: `3dbbad9 · git · go1.21 · linux/amd64`. Çıkış kodu 0.

### 4.2 Sözleşme okunabilir mi (çevrimdışı)

```bash
domainglass yetenek | python3 -m json.tool > /dev/null && echo KATALOG_OK
domainglass sema    | python3 -m json.tool > /dev/null && echo SEMA_OK
```
Beklenen: `KATALOG_OK` ve `SEMA_OK`. Bu iki komut ağ erişimi gerektirmez.

### 4.3 Canlı servis erişimi

```bash
domainglass saglik
```
Beklenen: `domain.glass servisi erişilebilir.` Çıkış kodu 0.

### 4.4 Uçtan uca rapor

```bash
domainglass -json -sessiz -hiz 1100 example.com > /tmp/domainglass-test.json; echo "exit=$?"
python3 -c "import json;d=json.load(open('/tmp/domainglass-test.json'));print(d['target'], d['kind'], len(d['sources']), 'kaynak')"
```
Beklenen: `exit=0` ve `example.com domain 10 kaynak`.

Çıkış kodu 3 ise hız sınırına takıldın: `-hiz 2500` ile tekrar dene.

---

## 5. Kurulum sonrası: ajan olarak nasıl kullanacaksın

### Altın kural

Bayrak ve komutları **tahmin etme**. Kataloğu oku:

```bash
domainglass yetenek    # komutlar, bayraklar, çıkış kodları, çözümleyiciler
domainglass sema       # rapor JSON şeması
```

### Çağrı kalıbı

```bash
# Tek hedef, tam rapor
domainglass -json -sessiz -hiz 1100 <hedef>

# Çoklu hedef, akış
printf '%s\n' hedef1 hedef2 | domainglass toplu -ndjson -sessiz

# Yalnızca belirli alanlar
domainglass -json -sessiz -alan summary,discovery <hedef>

# Ham satır listesi (boru hattı)
domainglass -sessiz -subs <hedef>
```

### Çıkış kodu sözleşmesi

| Kod | Ad | Senin davranışın |
|---|---|---|
| 0 | basarili | Sonucu kullan |
| 1 | hata | Hatayı raporla |
| 2 | bulunamadi | Kayıtlı değil — geçerli bulgu, hata değil |
| 3 | hiz_siniri | -hiz değerini artır, tekrar dene |
| 4 | kullanim | Komutu düzelt (domainglass yardim) |

### Kanal ayrımı

- Veri: **stdout**
- Günlük ve uyarı: **stderr**

Boru hattında günlükleri bastırmak için `-sessiz` kullan veya `2>/dev/null` ekle.

### Kısıtlar (bunları yapma)

- -json ile -subs/-ips/-related/-emails/-orgs birlikte kullanılamaz, kod 4 döner.
- -json ile -ndjson birlikte kullanılamaz, kod 4 döner.
- -alan tek başına kullanılamaz; -json veya -ndjson gerekir.
- -hiz 0 kullanma; art arda istekler 429/403 kapısını tetikler.

### Ajan konteyneri için önerilen ortam değişkenleri

```bash
export DOMAINGLASS_HIZ=1100        # istekler arası asgari aralık (ms)
export DOMAINGLASS_TIMEOUT=20      # istek zaman aşımı (sn)
export DOMAINGLASS_RENK=never      # günlük çıktısında renk yok
```

---

## 6. Doğrulama kontrol listesi

Bitirmeden önce hepsini işaretle. İşaretleyemediğin madde varsa kurulum tamam değildir.

- [ ] go version ≥ 1.21
- [ ] command -v domainglass bir yol döndürüyor
- [ ] domainglass kimlik çıkış kodu 0
- [ ] domainglass yetenek geçerli JSON
- [ ] domainglass sema geçerli JSON
- [ ] domainglass saglik çıkış kodu 0
- [ ] domainglass -json -sessiz example.com çıkış kodu 0 ve sources dizisi dolu
- [ ] Çıkış kodlarının 0/1/2/3/4 anlamını biliyorsun
- [ ] Veri/günlük kanal ayrımını biliyorsun (stdout / stderr)

---

## 7. Sorun giderme

### domainglass: command not found

Kurulum dizini PATH içinde değil. Ekle ve doğrula:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
command -v domainglass || echo 'bulunamadi'
```

### Çıkış kodu 3 (hız sınırı)

domain.glass hızlı ardışık istekleri sınırlar. Aralığı artır:

```bash
domainglass -json -sessiz -hiz 2500 <hedef>
```

### Rapor bazı kaynaklarda error / throttled gösteriyor

Bu beklenen bir durumdur. Rapor sources dizisinde her kaynağın durumunu bildirir ve
kalan veriyi yine üretir. Eksik alanı temiz sayma.

### Kurulan ikili güncel değil

Modül proxy `@latest` isteğini önbellekten eski bir commit ile yanıtlar. Bu depo
tag yayınlamadığı için proxy'nin güncellenmesini beklemek yerine commit ile
sabitle:

```bash
COMMIT=$(git ls-remote https://github.com/void0x14/domainglass refs/heads/main | cut -f1)
go install "github.com/void0x14/domainglass/cmd/domainglass@${COMMIT}"
domainglass kimlik    # commit eşleşmeli
```

### go install modül indiremiyor

Proxy arkasındaysan:

```bash
go env -w GOPROXY=https://proxy.golang.org,direct
```
Yine olmazsa Yol B'ye (git clone + go build) geç.

### İkili çalışıyor ama boş sonuç dönüyor

Çıkış kodu 2 ise hedef kayıtlı değildir; bu geçerli bir sonuçtur. Kod 0 ama veri
yoksa sources dizisini kontrol et.

---

## 8. Kaldırma

```bash
# go install ile kurulduysa
rm -f "$(go env GOPATH)/bin/domainglass"

# kaynak koddan kurulduysa
rm -f /usr/local/bin/domainglass "$HOME/.local/bin/domainglass"

# geçici derleme dizini
rm -rf /tmp/domainglass-build
```

---

## 9. Sonraki okuma (isteğe bağlı)

Kurulum bittikten sonra ihtiyacına göre:

| Belge | İçerik |
|---|---|
| [AGENTS.md](AGENTS.md) | Kısa ajan sözleşmesi |
| [docs/AI-AJANLARI.md](docs/AI-AJANLARI.md) | Ayrıntılı ajan kılavuzu ve tarifler |
| [docs/KULLANIM.md](docs/KULLANIM.md) | Tam kullanım başvurusu |
| [docs/MIMARI.md](docs/MIMARI.md) | Mimari ve tasarım kararları |
| [docs/API-KAYNAK.md](docs/API-KAYNAK.md) | Uç noktalar ve başlık sözleşmesi |

Bu belgeler kurulum için gerekli değildir; kurulum tamamen bu dosyadadır.

