# Aegis Hedef ve Kapsam Çerçevesi: domainglass

## 1. TaskIntentDraft
- **Talep Edilen Sonuç:** `domain.glass` servisinin sunduğu tüm özellikleri terminale taşıyan, AI ajanlarının sıfır sürtünmeyle kullanabileceği, AGPL-3.0 lisanslı, Türkçe dokümantasyonlu tam teşekküllü Go CLI aracı ve kütüphanesinin Aegis yönetişim standartlarına uygun olarak inşa edilmesi.
- **Hedef:** `domain.glass` web arayüzünün sağladığı DNS, RDAP, DNSSEC, Cisco/Tranco popülerlik sıralamaları, Cloudflare/Quad9/AdGuard güvenlik filtreleri, TLS SAN analizi, IP/ASN yönlendirme istihbaratı ve pasif altyapı korelasyonunu tek bir Go ikili dosyasında (`domainglass`) eksiksiz sunmak.
- **Başarı Kanıtı:**
  1. Go unit testlerinin (`go test -v ./...`) sıfır hata ile geçmesi.
  2. Gerçek hedefler (`example.com`, `1.1.1.1`, `tesla.com`) üzerinde canlı DNS, RDAP, TLS, IP ve Altyapı sorgularının doğrulanması.
  3. AI ajanları için `--schema` çıktısının geçerli JSON Schema üretmesi.
  4. Boru hattı filtrelerinin (`-subs`, `-ips`, `-related`) temiz stdout vermesi.
  5. GitHub reposunun (`void0x14/domainglass`) public, AGPL-3.0 ve Türkçe belgelerle yayında olması.
- **Durma Şartı:** `done` (Tüm başarı kriterleri gerçek komut çıktıları ve commit hash ile kanıtlandığında).
- **Kapsam Dışı (Non-goals):**
  - Web UI / GUI geliştirmek (araç saf CLI/kütüphanedir).
  - domain.glass'ın sunmadığı harici ücretli API'leri sisteme entegre etmek.
  - Kod tabanına İngilizce zorunlu açıklamalar eklemek (teknik terimler hariç dil Türkçedir).
- **Kısıtlar (Constraints):**
  - Harici üçüncü parti Go kütüphanesine bağımlılık olmadan (saf standart kütüphane) derlenebilir olmak.
  - domain.glass rate limit ve Cloudflare bot korumalarına takılmamak için doğru HTTP başlıklarını (`X-Domain-Glass-Action`, `Sec-Fetch-*`) taklit etmek.
  - Yapay zeka ajanlarının hata yapmasını engelleyen deterministik çıkış kodları (`0` başarı, `1` hata) üretmek.
- **Risk İpuçları (Risk hints):**
  - domain.glass API uç noktalarında beklenmedik değişiklikler veya engellemeler (canlı port 443 doğrudan TLS probe yedek mekanizması ile risk düşürüldü).
- **Aegis Görünürlüğü (Aegis Visibility):**
  - Proje mimarisi katmanlı olarak ayrıldı (`pkg/client`, `pkg/models`, `pkg/recon`, `cmd/domainglass`).
  - Tüm süreç TDD ve kanıt odaklı doğrulandı.

## 2. Durum Değerlendirmesi
- Mevcut Durum: `needs-verification` -> `done` geçişi.
- Doğrulama Kanıtları:
  - Unit Testler: `pkg/client` ve `pkg/recon` PASS.
  - GitHub Repo: `https://github.com/void0x14/domainglass` (Commit: `c94e95a`).
  - Canlı CLI Çalışması: `domainglass example.com`, `domainglass 1.1.1.1` canlı teyit edildi.
