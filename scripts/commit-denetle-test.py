#!/usr/bin/env python3
"""commit-denetle.py için testler.

Çalıştırma: python3 scripts/commit-denetle-test.py
Çıkış kodu: 0 tüm testler geçti, 1 en az bir test başarısız.
"""

import importlib.util
import os
import sys

BURADA = os.path.dirname(os.path.abspath(__file__))


def yukle():
    yol = os.path.join(BURADA, "commit-denetle.py")
    spec = importlib.util.spec_from_file_location("commit_denetle", yol)
    modul = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(modul)
    return modul


cd = yukle()

GECERLI = """fix(dg): hız sınırında yanlış hata sınıflandırması

Neden:
403 yanıtındaki browser_request_required kodu kalıcı erişim engeli olarak
sınıflandırılıyordu; bu kod aslında yeniden denenebilir.

Ne:
- internal/dg/istemci.go: 403 kapı kodlarını HataGecici sınıfına taşı
- internal/dg/istemci_test.go: yeniden deneme testi ekle

Kanıt:
go test ./internal/dg/ => PASS

Çekilebilir: evet
Kırıcı: yok
"""

# (ad, mesaj, gecerli_mi)
DURUMLAR = [
    ("geçerli mesaj", GECERLI, True),
    ("boş mesaj", "", False),
    ("yalnızca başlık", "fix(dg): bir şey", False),
    ("başlık biçimi bozuk", GECERLI.replace("fix(dg): hız", "duzeltme yaptim hız", 1), False),
    ("bilinmeyen tip", GECERLI.replace("fix(dg):", "yenilik(dg):", 1), False),
    ("başlık 72 karakteri aşıyor", "fix(dg): " + "a" * 80 + "\n\nNeden:\n" + "x" * 30 + "\n\nNe:\n- a\n\nKanıt:\ngo test\n\nÇekilebilir: evet\nKırıcı: yok", False),
    ("özet nokta ile bitiyor", GECERLI.replace("sınıflandırması\n", "sınıflandırması.\n", 1), False),
    ("Neden bölümü eksik", GECERLI.replace("Neden:\n403", "Gerekce:\n403", 1), False),
    ("Ne bölümü eksik", GECERLI.replace("Ne:\n-", "Degisiklik:\n-", 1), False),
    ("Kanıt bölümü eksik", GECERLI.replace("Kanıt:\ngo", "Dogrulama:\ngo", 1), False),
    ("Çekilebilir etiketi eksik", GECERLI.replace("Çekilebilir: evet\n", "", 1), False),
    ("Kırıcı etiketi eksik", GECERLI.replace("Kırıcı: yok", "", 1), False),
    ("Çekilebilir değeri geçersiz", GECERLI.replace("Çekilebilir: evet", "Çekilebilir: belki", 1), False),
    ("Neden çok kısa", "fix(dg): bir sey\n\nNeden:\nkisa\n\nNe:\n- a\n\nKanıt:\ngo test\n\nÇekilebilir: evet\nKırıcı: yok", False),
    ("Ne madde listesi değil", GECERLI.replace("- internal/dg/istemci.go: 403 kapı kodlarını HataGecici sınıfına taşı\n- internal/dg/istemci_test.go: yeniden deneme testi ekle", "duz metin yazdim", 1), False),
    ("başlıktan sonra boş satır yok", GECERLI.replace("sınıflandırması\n\nNeden:", "sınıflandırması\nNeden:", 1), False),
    ("kırıcı değişiklik bildirimi", GECERLI.replace("Kırıcı: yok", "Kırıcı: tool.version kaldırıldı", 1), True),
    ("çekilebilir hayır", GECERLI.replace("Çekilebilir: evet", "Çekilebilir: hayır", 1), True),
    ("kapsamsız başlık", GECERLI.replace("fix(dg):", "fix:", 1), True),
    ("yorum satırları yok sayılır", GECERLI + "\n# bu bir yorum\n", True),
]


def main():
    gecen = 0
    kalan = 0
    for ad, mesaj, beklenen in DURUMLAR:
        hatalar = cd.mesaj_denetle(mesaj)
        gecerli = not hatalar
        if gecerli == beklenen:
            gecen += 1
            print(f"  ✓ {ad}")
        else:
            kalan += 1
            print(f"  ✗ {ad}")
            print(f"      beklenen geçerli={beklenen}, alınan geçerli={gecerli}")
            for h in hatalar:
                print(f"      - {h}")

    print()
    print(f"{gecen} geçti, {kalan} başarısız (toplam {len(DURUMLAR)})")
    return 1 if kalan else 0


if __name__ == "__main__":
    sys.exit(main())

