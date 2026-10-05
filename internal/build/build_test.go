package build

import (
	"testing"
)

func TestKisalt(t *testing.T) {
	durumlar := []struct{ girdi, beklenen string }{
		{"3dbbad957c5b4072bd5bf2ba7fb2ae3f45b8969c", "3dbbad9"},
		{"307946d11c23", "307946d"},
		{"abc", "abc"},
		{"", ""},
	}
	for _, d := range durumlar {
		if got := kisalt(d.girdi); got != d.beklenen {
			t.Errorf("kisalt(%q) = %q; beklenen %q", d.girdi, got, d.beklenen)
		}
	}
}

func TestSozdeSurumAyiklama(t *testing.T) {
	durumlar := []struct{ girdi, beklenen string }{
		{"v0.0.0-20261005060518-307946d11c23", "307946d"},
		{"v0.0.0-20261005062709-3dbbad957c5b+dirty", "3dbbad9"},
		{"v1.2.3", ""},
		{"", ""},
		{"devel", ""},
	}
	for _, d := range durumlar {
		var got string
		if m := sozdeSurum.FindStringSubmatch(d.girdi); len(m) == 2 {
			got = kisalt(m[1])
		}
		if got != d.beklenen {
			t.Errorf("sözde sürüm %q -> %q; beklenen %q", d.girdi, got, d.beklenen)
		}
	}
}

func TestEtiket(t *testing.T) {
	durumlar := []struct {
		k        Kimlik
		beklenen string
	}{
		{Kimlik{Commit: "3dbbad9"}, "3dbbad9"},
		{Kimlik{Commit: "3dbbad9", Degisti: true}, "3dbbad9+"},
		{Kimlik{}, "bilinmiyor"},
	}
	for _, d := range durumlar {
		if got := d.k.Etiket(); got != d.beklenen {
			t.Errorf("Etiket() = %q; beklenen %q", got, d.beklenen)
		}
	}
}

func TestOkuTestOrtami(t *testing.T) {
	k := Oku()
	// go test ile çalışırken vcs bilgisi bulunmayabilir; ama fonksiyon
	// panik yapmamalı ve Go/platform her zaman dolu olmalı.
	if k.Go == "" {
		t.Error("Go sürümü boş olmamalı")
	}
	if k.Platform == "" {
		t.Error("platform boş olmamalı")
	}
	if k.Etiket() == "" {
		t.Error("Etiket() boş olmamalı")
	}
}

func TestTamBicim(t *testing.T) {
	k := Kimlik{Commit: "3dbbad9", Kaynak: "git", Go: "go1.21", Platform: "linux/amd64"}
	tam := k.Tam()
	for _, parca := range []string{"3dbbad9", "git", "go1.21", "linux/amd64"} {
		if !contains(tam, parca) {
			t.Errorf("Tam() içinde %q eksik: %s", parca, tam)
		}
	}
}

func contains(s, alt string) bool {
	for i := 0; i+len(alt) <= len(s); i++ {
		if s[i:i+len(alt)] == alt {
			return true
		}
	}
	return false
}
