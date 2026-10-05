// Package build, ikiliyi üreten kaynak ağacının kimliğini çıkarır.
//
// Neden gerekli: bu araç sürüm numarası taşımaz. Bir ikilinin hangi koddan
// üretildiğini ayırt etmek için sürüm yerine doğrudan kaynak kimliği kullanılır:
// git commit hash'i ve çalışma ağacının kirli olup olmadığı.
//
// İki kaynaktan beslenir:
//
//  1. Yerel git derlemesi: Go, ikiliye vcs.revision / vcs.time / vcs.modified
//     ayarlarını gömer. Bunlar en güvenilir kaynaktır.
//  2. Modül proxy üzerinden go install: vcs ayarları gelmez, ancak Go sözde
//     sürüm atar (v0.0.0-20261005060518-307946d11c23). Sondaki 12 haneli hex
//     parçası commit önekidir; buradan çıkarılır.
//
// Hiçbiri yoksa kimlik "bilinmiyor" olur; uydurma değer üretilmez.
package build

import (
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/void0x14/domainglass/internal/model"
)

// Kimlik, ikiliyi üreten kaynağın parmak izidir.
type Kimlik struct {
	// Kaynak, kimliğin nereden çıkarıldığını bildirir: "git", "modul" veya "bilinmiyor".
	Kaynak string
	// Commit, kısa commit hash'idir (7 hane). Yoksa boştur.
	Commit string
	// Zaman, commit zamanıdır (RFC3339). Yoksa boştur.
	Zaman string
	// Degisti, derleme sırasında çalışma ağacının kirli olduğunu bildirir.
	// Kirli bir derleme, commit ile birebir aynı kod değildir.
	Degisti bool
	// Go, derlemede kullanılan Go sürümüdür.
	Go string
	// Platform, GOOS/GOARCH çiftidir.
	Platform string
}

// Etiket, kimliğin tek satırlık okunabilir biçimidir.
// Örnek: "3dbbad9", "3dbbad9+" (kirli), "bilinmiyor".
func (k Kimlik) Etiket() string {
	if k.Commit == "" {
		return "bilinmiyor"
	}
	if k.Degisti {
		return k.Commit + "+"
	}
	return k.Commit
}

// Tam, kimliğin açıklayıcı tek satırlık biçimidir.
func (k Kimlik) Tam() string {
	parcalar := []string{k.Etiket()}
	if k.Kaynak != "" && k.Kaynak != "bilinmiyor" {
		parcalar = append(parcalar, k.Kaynak)
	}
	if k.Go != "" {
		parcalar = append(parcalar, k.Go)
	}
	if k.Platform != "" {
		parcalar = append(parcalar, k.Platform)
	}
	return strings.Join(parcalar, " · ")
}

// sozdeSurum, Go'nun modül proxy üzerinden kurulan ikililere atadığı sözde
// sürümden commit önekini ayıklar. Biçim: v0.0.0-<zaman>-<12 hex>.
var sozdeSurum = regexp.MustCompile(`v[0-9]+\.[0-9]+\.[0-9]+-[0-9]{14}-([0-9a-f]{12})`)

// Oku, çalışan ikilinin kaynak kimliğini döner.
func Oku() Kimlik {
	k := Kimlik{
		Kaynak:   "bilinmiyor",
		Go:       runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return k
	}

	var rev, zaman string
	var kirli bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			zaman = s.Value
		case "vcs.modified":
			kirli = s.Value == "true"
		}
	}
	if rev != "" {
		k.Kaynak = "git"
		k.Commit = kisalt(rev)
		k.Zaman = zaman
		k.Degisti = kirli
		return k
	}

	// Modül proxy yolu: commit bilgisi yalnızca sözde sürümde bulunur.
	if m := sozdeSurum.FindStringSubmatch(info.Main.Version); len(m) == 2 {
		k.Kaynak = "modul"
		k.Commit = kisalt(m[1])
		if !strings.Contains(info.Main.Version, "+dirty") {
			k.Degisti = false
		}
		return k
	}
	if strings.HasSuffix(info.Main.Version, "+dirty") {
		k.Degisti = true
	}
	return k
}

// kisalt, 40 haneli commit hash'ini 7 haneye indirir. Kısa girdilerde
// yalnızca ilk 7 haneyi döner; boşsa boş döner.
func kisalt(h string) string {
	h = strings.TrimSpace(h)
	if len(h) <= 7 {
		return h
	}
	return h[:7]
}

// AracBilgisi, rapora yazılan araç kimliğini üretir.
//
// Araç adı sabittir; ayırt edici bilgi commit hash'idir. Sürüm numarası yoktur.
func AracBilgisi(ad string) model.AracBilgisi {
	k := Oku()
	return model.AracBilgisi{
		Ad:       ad,
		Commit:   k.Commit,
		Degisti:  k.Degisti,
		Go:       k.Go,
		Platform: k.Platform,
	}
}
