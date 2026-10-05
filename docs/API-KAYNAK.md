# domain.glass Uç Noktaları ve Başlık Sözleşmesi

Bu belge, aracın hangi uç noktaları nasıl kullandığını kaydeder. Uç noktalar
canlı olarak doğrulanmıştır (2026-10-05).

## Genel başlık sözleşmesi

Site, API isteklerini tarayıcı bağlamına bağlar. Araç her isteğe şu başlıkları
ekler:

```http
User-Agent: Mozilla/5.0 (X11; Linux x86_64) ... Chrome/128.0.0.0 Safari/537.36
Accept: application/json
Accept-Language: en-US,en;q=0.9
Sec-Fetch-Site: same-origin
Sec-Fetch-Mode: cors
Sec-Fetch-Dest: empty
X-Domain-Glass-Action: <eylem>
Referer: https://domain.glass/<hedef>
Origin: https://domain.glass
```

Her uç nokta kendi X-Domain-Glass-Action değerini ve Referer yolunu bekler.

## Uç noktalar

| Uç nokta | Yöntem | Action | Hız sınırı |
|---|---|---|---|
| /api/v1/domain/{hedef} | GET | domain-page-load | Hayır |
| /api/v1/domain/{hedef}/safety | GET | domain-safety-load | Evet |
| /api/v1/classifications/{hedef} | GET | classification-load | Evet |
| /api/v1/tls/{tur}/{hedef} | GET | tls-page-load | Evet |
| /api/v1/web-profile/{hedef} | GET | domain-web-profile-load | Evet |
| /api/v1/whois/{hedef} | POST | whois-click | Evet |
| /api/v1/ip/{ip} | GET | ip-page-load | Evet |
| /api/v1/asn/{numara} | GET | asn-page-load | Hayır |
| /api/v1/rankings/domain/{hedef}?source={kaynak} | GET | — | Hayır |
| /api/v1/search?q={sorgu} | GET | search-click | Hayır |
| /api/v1/tld-iana/{tld} | GET | — | Hayır |
| /feeds/{ad}.xml | GET | — | Hayır |

## Kapı davranışı

Korunan uç noktalar (safety, classifications, tls, web-profile, whois, ip) hız
sınırına takıldığında 403 döner ve gövdede bir hata kodu taşır:

```json
{ "error": { "code": "browser_request_required", "message": "..." } }
```

Bu kodlar geçici sayılır ve araç üstel geri çekilmeyle yeniden dener. Kalıcı
doğrulama hataları (400) yeniden denenmez.

## Hız sınırı ölçümleri

Canlı test sonuçları (2026-10-05):

| Senaryo | Sonuç |
|---|---|
| 20 istek, beklemesiz | 3 başarılı, 17 adet 429 |
| 25 istek, 700 ms aralık | Tamamı 200 |
| 15 istek, 500 ms aralık | Tamamı 200 |
| 30 istek, 50 ms aralık | Tamamı 200 |

Not: 50 ms aralıkta bile başarı görülmüştür; ancak bu ölçüm tek bir zaman
penceresine aittir. Araç, güvenli pay bırakmak için varsayılan 1100 ms aralık
kullanır ve geçici hatalarda geri çekilir.

## Yanıt biçimleri

Tüm JSON uç noktaları ya doğrudan nesne ya da error zarfı döner:

```json
{ "error": { "source": "...", "code": "...", "message": "..." } }
```

domain.glass yanıtlarındaki generatedAt, maxAgeSeconds ve cache alanları
kaynak tazeliğini bildirir; araç bunları rapora taşımaz ama kaynak durumunu
sources bölümünde işaretler.

## RSS akışları

| Akış | Kaynak |
|---|---|
| trending.xml | Cisco Umbrella trend |
| top-gainers.xml | Cisco Umbrella en çok yükselenler |
| newly-ranked.xml | Cisco Umbrella yeni girenler |
| trending-tranco.xml | Tranco trend |
| top-gainers-tranco.xml | Tranco en çok yükselenler |
| newly-ranked-tranco.xml | Tranco yeni girenler |

Akış öğelerindeki bağlantılar https://domain.glass/<host> biçimindedir; araç
buradan host adını çıkarır.

## Bakım notu

domain.glass resmî bir API sözleşmesi yayınlamaz. Uç nokta yolları veya başlık
beklentileri değişirse ilgili kaynak düşer; rapor bunu sources bölümünde
error/throttled olarak bildirir. Yeni bir uç nokta eklemek için:

1. internal/dg/istemci.go içine metot ekleyin.
2. internal/model/ham.go içine yanıt tipini ekleyin.
3. internal/recon/topla.go içine paralel çağrıyı ekleyin.
4. internal/recon/isle.go içinde rapor dönüşümünü yapın.
5. internal/dg/istemci_test.go içine başlık sözleşmesi testini ekleyin.

