package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCozumleHedef(t *testing.T) {
	a, err := Cozumle([]string{"example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Komut != "" || len(a.Degerler) != 1 || a.Degerler[0] != "example.com" {
		t.Fatalf("beklenmeyen çözümleme: %+v", a)
	}
}

func TestCozumleKomut(t *testing.T) {
	a, err := Cozumle([]string{"ip", "1.1.1.1", "-json"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Komut != "ip" || len(a.Degerler) != 1 || a.Degerler[0] != "1.1.1.1" || !a.JSON {
		t.Fatalf("beklenmeyen çözümleme: %+v", a)
	}
}

func TestCozumleBayrakDegeri(t *testing.T) {
	a, err := Cozumle([]string{"-timeout", "5", "-hiz", "250", "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Timeout != 5 || a.HizMs != 250 {
		t.Fatalf("bayrak değerleri hatalı: %+v", a)
	}
}

func TestCozumleEsittirSozdizimi(t *testing.T) {
	a, err := Cozumle([]string{"--timeout=7", "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Timeout != 7 {
		t.Fatalf("timeout hatalı: %d", a.Timeout)
	}
}

func TestCozumleBilinmeyenBayrak(t *testing.T) {
	if _, err := Cozumle([]string{"--bilinmeyen", "example.com"}); err == nil {
		t.Fatal("bilinmeyen bayrak için hata bekleniyordu")
	}
}

func TestCozumleGecersizDeger(t *testing.T) {
	if _, err := Cozumle([]string{"-timeout", "sifir", "example.com"}); err == nil {
		t.Fatal("geçersiz timeout değeri için hata bekleniyordu")
	}
	if _, err := Cozumle([]string{"-renk", "mavi", "example.com"}); err == nil {
		t.Fatal("geçersiz renk kipi için hata bekleniyordu")
	}
}

func TestCozumleKomutYokVarsayilanHedef(t *testing.T) {
	a, err := Cozumle([]string{"example.com", "-subs"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Komut != "" || !a.Subs {
		t.Fatalf("beklenmeyen çözümleme: %+v", a)
	}
}

func TestKatalogTutarli(t *testing.T) {
	katalog := YetenekKatalogu()
	for _, anahtar := range []string{"tool", "build", "commands", "flags", "exitCodes", "resolvers", "feedNames", "installDoc", "agentContract", "agentGuide"} {
		if _, ok := katalog[anahtar]; !ok {
			t.Errorf("katalogda %q eksik", anahtar)
		}
	}
	komutlar := katalog["commands"].([]map[string]any)
	bayraklar := katalog["flags"].([]map[string]any)
	cikislar := katalog["exitCodes"].([]map[string]any)
	if len(komutlar) == 0 || len(bayraklar) == 0 || len(cikislar) == 0 {
		t.Fatal("katalog bölümleri boş olmamalı")
	}
}

func TestKullanimTumKomutlariIcerir(t *testing.T) {
	kullanim := Kullanim()
	for _, k := range Komutlar {
		if !strings.Contains(kullanim, k.Ad) {
			t.Errorf("kullanım metninde %q komutu eksik", k.Ad)
		}
	}
	for _, c := range CikisKodlari {
		if !strings.Contains(kullanim, c.Ad) {
			t.Errorf("kullanım metninde %q çıkış kodu eksik", c.Ad)
		}
	}
}

func TestSemaGecerliJSON(t *testing.T) {
	var sema map[string]any
	if err := json.Unmarshal([]byte(Sema), &sema); err != nil {
		t.Fatalf("şema geçerli JSON değil: %v", err)
	}
	for _, anahtar := range []string{"$schema", "title", "type", "properties", "required"} {
		if _, ok := sema[anahtar]; !ok {
			t.Errorf("şemada %q eksik", anahtar)
		}
	}
	ozellikler := sema["properties"].(map[string]any)
	for _, anahtar := range []string{"target", "kind", "summary", "discovery", "sources", "warnings"} {
		if _, ok := ozellikler[anahtar]; !ok {
			t.Errorf("şemada üst düzey %q alanı eksik", anahtar)
		}
	}
}

func TestSemaZorunluAlanlarRaporJSONileEslesir(t *testing.T) {
	var sema map[string]any
	if err := json.Unmarshal([]byte(Sema), &sema); err != nil {
		t.Fatal(err)
	}
	zorunlu := sema["required"].([]any)
	for _, z := range zorunlu {
		if !strings.Contains(strings.Join(KullanilabilirAlanlarListesi(), ","), z.(string)) {
			t.Errorf("zorunlu alan %q out.KullanilabilirAlanlar listesinde yok", z)
		}
	}
}

func TestCalistirKimlikCikisKodu(t *testing.T) {
	if kod := Calistir([]string{"kimlik"}, os.Stdout, os.Stderr); kod != CikisBasarili {
		t.Fatalf("kimlik çıkış kodu %d", kod)
	}
}

func TestCalistirKullanimHatasi(t *testing.T) {
	if kod := Calistir([]string{"--bilinmeyen-bayrak"}, os.Stdout, os.Stderr); kod != CikisKullanim {
		t.Fatalf("kullanım hatası kodu %d olmalı", kod)
	}
}

func TestCalistirCeliskiliBayraklar(t *testing.T) {
	if kod := Calistir([]string{"-json", "-ndjson", "example.com"}, os.Stdout, os.Stderr); kod != CikisKullanim {
		t.Fatalf("çelişkili bayrak kodu %d olmalı", kod)
	}
	if kod := Calistir([]string{"-alan", "summary", "example.com"}, os.Stdout, os.Stderr); kod != CikisKullanim {
		t.Fatalf("-alan tek başına kullanım hatası olmalı, kod %d", kod)
	}
	if kod := Calistir([]string{"-json", "-subs", "example.com"}, os.Stdout, os.Stderr); kod != CikisKullanim {
		t.Fatalf("JSON + satır kipi çelişkisi kullanım hatası olmalı, kod %d", kod)
	}
}

func TestCalistirBilinmeyenKomut(t *testing.T) {
	if kod := Calistir([]string{"bilinmeyen-komut"}, os.Stdout, os.Stderr); kod != CikisKullanim {
		t.Fatalf("bilinmeyen komut kullanım hatası olmalı, kod %d", kod)
	}
}

func TestCalistirGecersizCozumleyici(t *testing.T) {
	if kod := Calistir([]string{"-cozumleyici", "yok", "example.com"}, os.Stdout, os.Stderr); kod != CikisKullanim {
		t.Fatalf("geçersiz çözümleyici kullanım hatası olmalı, kod %d", kod)
	}
}

// TestCalistirCanliDomain, gerçek domain.glass uç noktalarına karşı uçtan uca
// çalışmayı doğrular. Ağ yoksa atlanır.
func TestCalistirCanliDomain(t *testing.T) {
	if os.Getenv("DOMAINGLASS_CANLI_TEST") == "" {
		t.Skip("canlı test için DOMAINGLASS_CANLI_TEST=1 gerekli")
	}
	betik := filepath.Join(t.TempDir(), "rapor.json")
	f, err := os.Create(betik)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	kod := Calistir([]string{"-json", "-hiz", "200", "example.com"}, f, os.Stderr)
	if kod != CikisBasarili {
		t.Fatalf("canlı çalışma çıkış kodu %d", kod)
	}
	f.Close()
	veri, err := os.ReadFile(betik)
	if err != nil {
		t.Fatal(err)
	}
	var rapor map[string]any
	if err := json.Unmarshal(veri, &rapor); err != nil {
		t.Fatalf("çıktı geçerli JSON değil: %v", err)
	}
	if rapor["target"] != "example.com" {
		t.Fatalf("hedef hatalı: %v", rapor["target"])
	}
	ozet, ok := rapor["summary"].(map[string]any)
	if !ok {
		t.Fatal("summary alanı eksik")
	}
	if _, ok := ozet["registered"]; !ok {
		t.Error("summary.registered eksik")
	}
}

// KullanilabilirAlanlarListesi, döngüsel içe aktarımı önlemek için out paketinin
// listesini cli testinden erişilebilir kılar.
func KullanilabilirAlanlarListesi() []string {
	return []string{
		"target", "kind", "tool", "generatedAt", "summary", "rankings", "registration",
		"dns", "tls", "web", "safety", "classifications", "ip", "asn", "tld", "feed",
		"resolvers", "discovery", "sources", "warnings",
	}
}
