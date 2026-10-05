# AGENTS.md — AI Ajanları İçin Çalıştırma Sözleşmesi

Bu doküman, bu depoyu bir CLI aracı veya kütüphane olarak kullanacak yapay zeka ajanları (LLM, Autonomous Agent, Code Assistant) için kesin ve sürtünmesiz çalışma kurallarını tanımlar.

## Temel Kurallar

1. **Tahmin Etme, Şemayı Oku:**
   - Aracın ürettiği JSON çıktısının yapısını öğrenmek için doğrudan çalıştır:
     ```bash
     ./domainglass -schema
     ```
2. **Deterministik Çıkış:**
   - Yapılandırılmış veri okumak istiyorsan her zaman `-json` bayrağını kullan:
     ```bash
     ./domainglass -json <hedef>
     ```
   - Çıktı geçerli bir JSON nesnesidir (`stdout`), hata logları ise `stderr`'e yazılır.
3. **Alt Alan Adı ve IP Ayıklama:**
   - Eğer görevin hedef için alt alan adı veya IP keşfi ise, JSON'ı parse etmekle uğraşma; doğrudan optimize bayrakları kullan:
     ```bash
     ./domainglass -subs <domain>   # Sadece subdomain satırları
     ./domainglass -ips <domain>    # Sadece IP satırları
     ./domainglass -related <domain> # İlişkili kök domainler
     ```
4. **Hata Kodları (Exit Codes):**
   - `0`: Sorgu başarılı tamamlandı.
   - `1`: Geçersiz parametre, bağlantı hatası veya çözümlenemeyen hedef.

## Örnek Ajan Komutları

```bash
# JSON raporu alma
./domainglass -json example.com

# IP yönlendirme ve ASN analizi
./domainglass -json 1.1.1.1

# Toplu liste akışı
cat targets.txt | ./domainglass -subs
```
