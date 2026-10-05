package dns

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestEncodeQuery(t *testing.T) {
	q, err := EncodeQuery(0x1234, "example.com", TypeA)
	if err != nil {
		t.Fatalf("EncodeQuery hata: %v", err)
	}
	want := "123401000001000000000000076578616d706c6503636f6d0000010001"
	got := hex.EncodeToString(q)
	if got != want {
		t.Fatalf("sorgu iletimi\n istenen: %s\n alınan : %s", want, got)
	}
}

func TestEncodeQueryGecersizAd(t *testing.T) {
	if _, err := EncodeQuery(1, "", TypeA); err == nil {
		t.Fatal("boş ad için hata bekleniyordu")
	}
	uzun := strings.Repeat("a", 64) + ".com"
	if _, err := EncodeQuery(1, uzun, TypeA); err == nil {
		t.Fatal("63 baytı aşan etiket için hata bekleniyordu")
	}
}

// TestDecodeResponse gerçek bir DNS yanıtını ayrıştırır. İletim, A kaydı
// içeren sıkıştırmalı bir yanıttır (example.com -> 93.184.216.34).
func TestDecodeResponseA(t *testing.T) {
	hexMsg := "123481800001000100000000076578616d706c6503636f6d0000010001c00c000100010000003c00045db8d822"
	msg, err := hex.DecodeString(hexMsg)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := DecodeResponse(msg)
	if err != nil {
		t.Fatalf("DecodeResponse hata: %v", err)
	}
	if resp.Rcode != "NOERROR" {
		t.Fatalf("rcode istenen NOERROR, alınan %s", resp.Rcode)
	}
	if len(resp.Records) != 1 {
		t.Fatalf("1 kayıt bekleniyordu, %d alındı", len(resp.Records))
	}
	r := resp.Records[0]
	if r.Name != "example.com" || r.Type != "A" || r.Value != "93.184.216.34" || r.TTL != 60 {
		t.Fatalf("kayıt hatalı: %+v", r)
	}
}

func TestDecodeResponseNXDomain(t *testing.T) {
	hexMsg := "abcd818300010000000000000f62756c756e616d61796f6b7475726b03636f6d0000010001"
	msg, err := hex.DecodeString(hexMsg)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := DecodeResponse(msg)
	if err != nil {
		t.Fatalf("DecodeResponse hata: %v", err)
	}
	if resp.Rcode != "NXDOMAIN" {
		t.Fatalf("rcode istenen NXDOMAIN, alınan %s", resp.Rcode)
	}
	if len(resp.Records) != 0 {
		t.Fatalf("kayıt beklenmiyordu: %+v", resp.Records)
	}
}

func TestDecodeResponseBozuk(t *testing.T) {
	if _, err := DecodeResponse([]byte{0x00}); err == nil {
		t.Fatal("kısa iletim için hata bekleniyordu")
	}
}

func TestEngelli(t *testing.T) {
	engelli := &Response{Records: []Record{{Type: "A", Value: "0.0.0.0"}}}
	if !Engelli(engelli) {
		t.Fatal("0.0.0.0 engelli sayılmalıydı")
	}
	serbest := &Response{Records: []Record{{Type: "A", Value: "93.184.216.34"}}}
	if Engelli(serbest) {
		t.Fatal("genel IP engelli sayılmamalıydı")
	}
	v6 := &Response{Records: []Record{{Type: "AAAA", Value: "::"}}}
	if !Engelli(v6) {
		t.Fatal(":: engelli sayılmalıydı")
	}
}

func TestSaglayiciGetir(t *testing.T) {
	if _, ok := SaglayiciGetir("quad9"); !ok {
		t.Fatal("quad9 sağlayıcısı bulunamadı")
	}
	if _, ok := SaglayiciGetir("yok-boyle-bir-sey"); ok {
		t.Fatal("bilinmeyen sağlayıcı bulunmamalıydı")
	}
}

func TestParseType(t *testing.T) {
	if typ, ok := ParseType("aaaa"); !ok || typ != TypeAAAA {
		t.Fatalf("aaaa ayrıştırılamadı: %v %v", typ, ok)
	}
	if _, ok := ParseType("BILINMEYEN"); ok {
		t.Fatal("bilinmeyen tip kabul edilmemeliydi")
	}
}
