#!/bin/sh
# Git kancalarını bu depo için etkinleştirir.
#
# Neden gerekli: commit mesajı bu depoda tek "sürüm notu"dur (sürüm numarası ve
# tag yok). Mesaj biçiminin denetlenmesi için kancaların kurulu olması gerekir.
# Kancalar .githooks/ altında sürümlenir; bu betik core.hooksPath ayarını yapar.
#
# Kullanım:
#   sh scripts/kur-kancalar.sh        # bu depoda etkinleştir
#
# Çıkış kodları: 0 başarılı, 1 hata.

set -eu

if [ ! -d .git ]; then
  echo "Bu dizin bir git deposu değil: $(pwd)" >&2
  exit 1
fi

git config core.hooksPath .githooks
chmod +x .githooks/* 2>/dev/null || true

echo "Kancalar etkinleştirildi: core.hooksPath=.githooks"
echo "Commit mesajı denetimi: python3 scripts/commit-denetle.py --liste"

