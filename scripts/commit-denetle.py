#!/usr/bin/env python3
"""Commit mesajı denetleyicisi.

Neden gerekli: bu depo sürüm numarası taşımaz ve tag yayınlamaz. Bir commit
mesajı, o değişikliğin tek "sürüm notu"dur. Hangi commit'in ne içerdiği, neden
atıldığı ve nasıl doğrulandığı yalnızca buradan öğrenilebilir. Bu yüzden mesaj
biçimi serbest değil, denetlenen bir sözleşmedir.

Kullanım:
  python3 scripts/commit-denetle.py <mesaj-dosyasi>   # commit-msg kancası
  python3 scripts/commit-denetle.py --stdin            # boru hattından
  python3 scripts/commit-denetle.py --commit <ref>     # mevcut bir commit
  python3 scripts/commit-denetle.py --gecmis <n>       # son n commit
  python3 scripts/commit-denetle.py --liste            # kural listesi

Çıkış kodları: 0 geçerli, 1 kural ihlali, 2 kullanım hatası.
"""

import re
import subprocess
import sys

# İzin verilen değişiklik tipleri. Yeni tip eklemek kırıcı bir sözleşme
# değişikliğidir; bu listeyi güncellemeden kullanılamaz.
TIPLER = {
    "feat": "yeni yetenek veya davranış",
    "fix": "hatalı davranışı düzeltme",
    "refactor": "davranışı değiştirmeyen kod düzenlemesi",
    "docs": "yalnızca belge değişikliği",
    "test": "yalnızca test ekleme veya düzeltme",
    "chore": "derleme, araç, bağımlılık gibi bakım işleri",
    "perf": "ölçülebilir performans iyileştirmesi",
    "revert": "önceki bir commit'i geri alma",
}

# Zorunlu bölümler ve en az uzunlukları (karakter).
ZORUNLU_BOLUMLER = [
    ("Neden", 20, "Değişiklik neden gerekliydi; hangi sorun veya eksik tetikledi."),
    ("Ne", 10, "Ne değişti; madde madde somut dosya/davranış listesi."),
    ("Kanıt", 10, "Nasıl doğrulandı; çalıştırılan komut ve görülen sonuç."),
]

# Zorunlu, makine tarafından ayrıştırılabilen etiketler.
ETIKETLER = [
    ("Çekilebilir", r"^(evet|hayır)$",
     "evet: bu commit tek başına çekilip kullanılabilir. "
     "hayır: başka bir commit'e bağımlı, gerekçesi Ne bölümünde olmalı."),
    ("Kırıcı", r".+",
     "yok: geriye dönük uyumlu. Aksi hâlde neyin bozulduğunu yaz "
     "(bayrak, JSON alanı, çıkış kodu, dosya yolu)."),
]

BASLIK_KALIBI = re.compile(r"^([a-z]+)(?:\(([^)]+)\))?!?: (.+)$")


def baslik_denetle(satir, hatalar):
    """İlk satırı denetler: tip(kapsam): özet"""
    if len(satir) > 72:
        hatalar.append(f"başlık 72 karakteri aşıyor ({len(satir)}): {satir[:60]}...")
    eslesme = BASLIK_KALIBI.match(satir)
    if not eslesme:
        hatalar.append(
            "başlık biçimi hatalı. Beklenen: \"tip(kapsam): özet\" veya "
            "\"tip: özet\". Örnek: \"fix(dg): hız sınırında yanlış sınıflandırma\""
        )
        return
    tip, kapsam, ozet = eslesme.group(1), eslesme.group(2), eslesme.group(3).strip()
    if tip not in TIPLER:
        hatalar.append(
            f"bilinmeyen tip: {tip!r}. Geçerli tipler: {", ".join(sorted(TIPLER))}"
        )
    if not ozet:
        hatalar.append("özet boş olamaz")
    if ozet.endswith("."):
        hatalar.append("özet nokta ile bitmemeli")
    if ozet[0:1].isupper() and tip != "revert":
        hatalar.append("özet küçük harfle başlamalı (başlık stili)")


def bolumleri_ayikla(govde):
    """Gövdeyi \"Ad: içerik\" bölümlerine ayırır."""
    bolumler = {}
    aktif = None
    for satir in govde.splitlines():
        eslesme = re.match(r"^([A-ZÇĞİÖŞÜ][A-Za-zÇĞİÖŞÜçğıöşü ]{1,15}):\s*(.*)$", satir)
        if eslesme and not satir.startswith("-"):
            aktif = eslesme.group(1).strip()
            bolumler[aktif] = eslesme.group(2)
        elif aktif is not None:
            bolumler[aktif] += "\n" + satir
    return bolumler


def mesaj_denetle(mesaj):
    """Bir commit mesajını denetler ve bulunan hataları döner."""
    hatalar = []
    # Yorum satırlarını ve git'in eklediği yardım metnini at.
    satirlar = [s for s in mesaj.splitlines() if not s.startswith("#")]
    # Sondaki boş satırları kırp.
    while satirlar and not satirlar[-1].strip():
        satirlar.pop()
    if not satirlar:
        return ["mesaj boş"]

    baslik = satirlar[0].rstrip()
    baslik_denetle(baslik, hatalar)

    if len(satirlar) > 1 and satirlar[1].strip():
        hatalar.append("başlıktan sonra boş satır olmalı")

    govde = "\n".join(satirlar[2:]) if len(satirlar) > 2 else ""
    bolumler = bolumleri_ayikla(govde)

    for ad, asgari, aciklama in ZORUNLU_BOLUMLER:
        if ad not in bolumler:
            hatalar.append(f"{ad}: bölümü eksik. {aciklama}")
            continue
        icerik = bolumler[ad].strip()
        if len(icerik) < asgari:
            hatalar.append(f"{ad}: bölümü çok kısa (en az {asgari} karakter). {aciklama}")

    if "Ne" in bolumler and "-" not in bolumler["Ne"]:
        hatalar.append("Ne: bölümü madde listesi olmalı (her satır \"- \" ile başlamalı)")

    for ad, kalip, aciklama in ETIKETLER:
        if ad not in bolumler:
            hatalar.append(f"{ad}: etiketi eksik. {aciklama}")
            continue
        deger = bolumler[ad].strip().splitlines()[0].strip() if bolumler[ad].strip() else ""
        if not re.match(kalip, deger):
            hatalar.append(f"{ad}: değeri {deger!r} geçersiz. {aciklama}")
    return hatalar


def kural_listesi():
    satirlar = []
    satirlar.append("Commit mesajı biçimi:")
    satirlar.append("")
    satirlar.append("  <tip>(<kapsam>): <özet>")
    satirlar.append("")
    satirlar.append("  Neden:")
    satirlar.append("  <bu değişiklik neden gerekliydi>")
    satirlar.append("")
    satirlar.append("  Ne:")
    satirlar.append("  - <somut değişiklik>")
    satirlar.append("  - <somut değişiklik>")
    satirlar.append("")
    satirlar.append("  Kanıt:")
    satirlar.append("  <çalıştırılan komut ve görülen sonuç>")
    satirlar.append("")
    satirlar.append("  Çekilebilir: evet|hayır")
    satirlar.append("  Kırıcı: yok|<ne bozuldu>")
    satirlar.append("")
    satirlar.append("Geçerli tipler:")
    for tip, aciklama in TIPLER.items():
        satirlar.append(f"  {tip:10s} {aciklama}")
    satirlar.append("")
    satirlar.append("Zorunlu bölümler:")
    for ad, asgari, aciklama in ZORUNLU_BOLUMLER:
        satirlar.append(f"  {ad:10s} {aciklama}")
    satirlar.append("")
    satirlar.append("Zorunlu etiketler:")
    for ad, _, aciklama in ETIKETLER:
        satirlar.append(f"  {ad}: {aciklama}")
    satirlar.append("")
    satirlar.append("Tam kural: docs/COMMIT-STANDART.md")
    return "\n".join(satirlar)


def git(*args):
    return subprocess.run(["git", *args], capture_output=True, text=True).stdout


def main(argv):
    if not argv:
        print(__doc__.strip(), file=sys.stderr)
        return 2

    if argv[0] == "--liste":
        print(kural_listesi())
        return 0

    if argv[0] == "--stdin":
        mesaj = sys.stdin.read()
        hatalar = mesaj_denetle(mesaj)
        return raporla(hatalar, "stdin")

    if argv[0] == "--commit":
        if len(argv) < 2:
            print("--commit için bir ref gerekli", file=sys.stderr)
            return 2
        mesaj = git("log", "-1", "--format=%B", argv[1])
        return raporla(mesaj_denetle(mesaj), argv[1])

    if argv[0] == "--aralik":
        # Belirli bir git aralığındaki commit'leri denetler (ör. origin/main..HEAD).
        # Bu, paylaşılmış eski geçmişi denetlemeden yalnızca yeni commit'lere
        # bakmayı sağlar; kural yürürlüğe girmeden önceki mesajlar kapsam dışıdır.
        if len(argv) < 2:
            print("--aralik için bir git aralığı gerekli (ör. origin/main..HEAD)", file=sys.stderr)
            return 2
        refs = git("rev-list", argv[1]).split()
        if not refs:
            print(f"aralıkta commit yok: {argv[1]}")
            return 0
        return gecmisi_denetle(refs)

    if argv[0] == "--gecmis":
        n = argv[1] if len(argv) > 1 else "10"
        refs = git("log", f"-{n}", "--format=%H").split()
        return gecmisi_denetle(refs)

    # Konumsal argüman: commit-msg kancasından gelen dosya yolu.
    yol = argv[0]
    try:
        with open(yol, encoding="utf-8") as f:
            mesaj = f.read()
    except OSError as e:
        print(f"mesaj dosyası okunamadı: {e}", file=sys.stderr)
        return 2
    return raporla(mesaj_denetle(mesaj), yol)


def gecmisi_denetle(refs):
    """Verilen commit listesini denetler ve özet döner."""
    toplam_hata = 0
    if True:
        for ref in refs:
            mesaj = git("log", "-1", "--format=%B", ref)
            hatalar = mesaj_denetle(mesaj)
            kisa = ref[:7]
            baslik = mesaj.splitlines()[0] if mesaj.strip() else "(boş)"
            if hatalar:
                toplam_hata += len(hatalar)
                print(f"✗ {kisa}  {baslik}")
                for h in hatalar:
                    print(f"      - {h}")
            else:
                print(f"✓ {kisa}  {baslik}")
        print()
        if toplam_hata:
            print(f"{toplam_hata} kural ihlali bulundu.")
            return 1
        print("Tüm commit mesajları kurala uygun.")
        return 0


def raporla(hatalar, kaynak):
    if not hatalar:
        return 0
    print("Commit mesajı kurala uymuyor:", file=sys.stderr)
    for h in hatalar:
        print(f"  - {h}", file=sys.stderr)
    print("", file=sys.stderr)
    print("Beklenen biçim:", file=sys.stderr)
    print("", file=sys.stderr)
    print("  <tip>(<kapsam>): <özet>", file=sys.stderr)
    print("", file=sys.stderr)
    print("  Neden:", file=sys.stderr)
    print("  <bu değişiklik neden gerekliydi>", file=sys.stderr)
    print("", file=sys.stderr)
    print("  Ne:", file=sys.stderr)
    print("  - <somut değişiklik>", file=sys.stderr)
    print("", file=sys.stderr)
    print("  Kanıt:", file=sys.stderr)
    print("  <çalıştırılan komut ve görülen sonuç>", file=sys.stderr)
    print("", file=sys.stderr)
    print("  Çekilebilir: evet|hayır", file=sys.stderr)
    print("  Kırıcı: yok|<ne bozuldu>", file=sys.stderr)
    print("", file=sys.stderr)
    print("Tam kural: docs/COMMIT-STANDART.md", file=sys.stderr)
    print("Kural listesi: python3 scripts/commit-denetle.py --liste", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

