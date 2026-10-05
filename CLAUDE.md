# CLAUDE.md

Kurulum için [KURULUM.md](KURULUM.md) dosyasını oku ve adımları sırayla uygula.

Bu depo için ajan sözleşmesi [AGENTS.md](AGENTS.md) dosyasındadır. Özet:

- Aracın kataloğunu tahmin etme; `domainglass yetenek` ve `domainglass sema` çalıştır.
- Yapılandırılmış veri için `-json`, akış için `-ndjson` kullan.
- Veri stdout'ta, günlükler stderr'dedir.
- Çıkış kodları: 0 başarılı, 1 hata, 2 kayıtlı değil, 3 hız sınırı, 4 kullanım hatası.
- `-json` ile `-subs`/`-ips`/`-related` birlikte kullanılamaz.
- Commit mesajı biçimi zorunludur; kanca denetler. Bkz. docs/COMMIT-STANDART.md.

Ayrıntı: [docs/AI-AJANLARI.md](docs/AI-AJANLARI.md)

