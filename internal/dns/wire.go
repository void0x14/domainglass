// Package dns, DNS iletim biçimini (RFC 1035) kodlar ve çözer.
//
// Neden gerekli: domain.glass uç noktaları hız sınırına takıldığında (429) veya
// erişilemediğinde aracın DNS verisini bağımsız toplayabilmesi gerekir. Ayrıca
// Quad9, AdGuard, CleanBrowsing ve Control D filtre sağlayıcıları yalnızca ikili
// DNS iletimi (application/dns-message) sunar; JSON tabanlı DoH yalnızca Google
// ve Cloudflare tarafında bulunur.
package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

// Type, DNS kayıt tipini temsil eder.
type Type uint16

// Desteklenen DNS kayıt tipleri.
const (
	TypeA     Type = 1
	TypeNS    Type = 2
	TypeCNAME Type = 5
	TypeSOA   Type = 6
	TypePTR   Type = 12
	TypeMX    Type = 15
	TypeTXT   Type = 16
	TypeAAAA  Type = 28
	TypeCAA   Type = 257
)

// AllTypes, raporlamada kullanılan standart kayıt tipi sırasıdır.
var AllTypes = []Type{TypeA, TypeAAAA, TypeCNAME, TypeMX, TypeNS, TypeTXT, TypeCAA, TypeSOA}

// TypeName, kayıt tipini okunabilir adına çevirir.
func TypeName(t Type) string {
	switch t {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeCNAME:
		return "CNAME"
	case TypeSOA:
		return "SOA"
	case TypePTR:
		return "PTR"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	case TypeCAA:
		return "CAA"
	default:
		return fmt.Sprintf("TYPE%d", uint16(t))
	}
}

// ParseType, kayıt tipi adını tip sabitine çevirir.
func ParseType(s string) (Type, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "A":
		return TypeA, true
	case "NS":
		return TypeNS, true
	case "CNAME":
		return TypeCNAME, true
	case "SOA":
		return TypeSOA, true
	case "PTR":
		return TypePTR, true
	case "MX":
		return TypeMX, true
	case "TXT":
		return TypeTXT, true
	case "AAAA":
		return TypeAAAA, true
	case "CAA":
		return TypeCAA, true
	default:
		return 0, false
	}
}

// Rcode, DNS yanıt kodudur.
type Rcode uint8

// DNS yanıt kodları.
const (
	RcodeSuccess  Rcode = 0
	RcodeFormErr  Rcode = 1
	RcodeServFail Rcode = 2
	RcodeNXDomain Rcode = 3
	RcodeNotImpl  Rcode = 4
	RcodeRefused  Rcode = 5
)

// String, yanıt kodunu okunabilir adına çevirir.
func (r Rcode) String() string {
	switch r {
	case RcodeSuccess:
		return "NOERROR"
	case RcodeFormErr:
		return "FORMERR"
	case RcodeServFail:
		return "SERVFAIL"
	case RcodeNXDomain:
		return "NXDOMAIN"
	case RcodeNotImpl:
		return "NOTIMP"
	case RcodeRefused:
		return "REFUSED"
	default:
		return fmt.Sprintf("RCODE%d", uint8(r))
	}
}

// Record, çözümlenmiş tek bir DNS kaydıdır.
type Record struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   uint32 `json:"ttl"`
	Value string `json:"value"`
}

// Response, bir DNS yanıtının ayrıştırılmış hâlidir.
type Response struct {
	ID        uint16   `json:"id"`
	Rcode     string   `json:"rcode"`
	Truncated bool     `json:"truncated"`
	Records   []Record `json:"records"`
}

// ErrMalformed, bozuk DNS iletimi için döner.
var ErrMalformed = errors.New("bozuk DNS iletimi")

// EncodeQuery, verilen ad ve tip için DNS sorgu iletimi üretir.
func EncodeQuery(id uint16, name string, qtype Type) ([]byte, error) {
	name = strings.TrimSuffix(strings.TrimSpace(name), ".")
	if name == "" {
		return nil, errors.New("boş DNS adı")
	}
	buf := make([]byte, 0, 32+len(name))
	buf = binary.BigEndian.AppendUint16(buf, id)
	buf = binary.BigEndian.AppendUint16(buf, 0x0100)
	buf = binary.BigEndian.AppendUint16(buf, 1)
	buf = binary.BigEndian.AppendUint16(buf, 0)
	buf = binary.BigEndian.AppendUint16(buf, 0)
	buf = binary.BigEndian.AppendUint16(buf, 0)

	for _, label := range strings.Split(name, ".") {
		if label == "" {
			return nil, fmt.Errorf("geçersiz DNS adı: %q", name)
		}
		if len(label) > 63 {
			return nil, fmt.Errorf("DNS etiketi 63 baytı aşıyor: %q", label)
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, label...)
	}
	buf = append(buf, 0)
	buf = binary.BigEndian.AppendUint16(buf, uint16(qtype))
	buf = binary.BigEndian.AppendUint16(buf, 1)
	return buf, nil
}

// DecodeResponse, DNS yanıt iletimini ayrıştırır.
func DecodeResponse(msg []byte) (*Response, error) {
	if len(msg) < 12 {
		return nil, ErrMalformed
	}
	id := binary.BigEndian.Uint16(msg[0:2])
	flags := binary.BigEndian.Uint16(msg[2:4])
	qd := int(binary.BigEndian.Uint16(msg[4:6]))
	an := int(binary.BigEndian.Uint16(msg[6:8]))
	ns := int(binary.BigEndian.Uint16(msg[8:10]))
	ar := int(binary.BigEndian.Uint16(msg[10:12]))

	resp := &Response{
		ID:        id,
		Rcode:     Rcode(flags & 0x000F).String(),
		Truncated: flags&0x0200 != 0,
		Records:   []Record{},
	}

	off := 12
	for i := 0; i < qd; i++ {
		_, next, err := decodeName(msg, off)
		if err != nil {
			return nil, err
		}
		off = next + 4
		if off > len(msg) {
			return nil, ErrMalformed
		}
	}

	total := an + ns + ar
	for i := 0; i < total; i++ {
		name, next, err := decodeName(msg, off)
		if err != nil {
			return nil, err
		}
		off = next
		if off+10 > len(msg) {
			return nil, ErrMalformed
		}
		rtype := Type(binary.BigEndian.Uint16(msg[off : off+2]))
		ttl := binary.BigEndian.Uint32(msg[off+4 : off+8])
		rdlen := int(binary.BigEndian.Uint16(msg[off+8 : off+10]))
		off += 10
		if off+rdlen > len(msg) {
			return nil, ErrMalformed
		}
		rdataStart := off
		off += rdlen

		if i >= an {
			continue
		}
		value, err := decodeRData(msg, rdataStart, rdlen, rtype)
		if err != nil {
			continue
		}
		resp.Records = append(resp.Records, Record{
			Name:  name,
			Type:  TypeName(rtype),
			TTL:   ttl,
			Value: value,
		})
	}
	return resp, nil
}

// decodeName, msg[off] konumundan başlayarak sıkıştırma işaretçilerini izleyip
// bir DNS adını çözer. İkinci dönen değer, işaretçi izlenmeden önceki bir
// sonraki okuma konumudur.
func decodeName(msg []byte, off int) (string, int, error) {
	var labels []string
	pos := off
	next := -1
	guard := 0
	for {
		if pos >= len(msg) {
			return "", 0, ErrMalformed
		}
		guard++
		if guard > 256 {
			return "", 0, ErrMalformed
		}
		length := int(msg[pos])
		if length == 0 {
			pos++
			break
		}
		if length&0xC0 == 0xC0 {
			if pos+1 >= len(msg) {
				return "", 0, ErrMalformed
			}
			ptr := int(binary.BigEndian.Uint16(msg[pos:pos+2]) & 0x3FFF)
			if next == -1 {
				next = pos + 2
			}
			if ptr >= len(msg) {
				return "", 0, ErrMalformed
			}
			pos = ptr
			continue
		}
		if length > 63 || pos+1+length > len(msg) {
			return "", 0, ErrMalformed
		}
		labels = append(labels, string(msg[pos+1:pos+1+length]))
		pos += 1 + length
	}
	if next == -1 {
		next = pos
	}
	return strings.Join(labels, "."), next, nil
}

// decodeRData, kayıt verisini (RDATA) metne çevirir. rdataStart, RDATA konumunun
// iletim içindeki mutlak konumudur; sıkıştırma işaretçileri bu sayede çözülür.
func decodeRData(msg []byte, rdataStart, rdlen int, t Type) (string, error) {
	rdata := msg[rdataStart : rdataStart+rdlen]
	switch t {
	case TypeA:
		if len(rdata) != 4 {
			return "", ErrMalformed
		}
		return net.IP(rdata).String(), nil
	case TypeAAAA:
		if len(rdata) != 16 {
			return "", ErrMalformed
		}
		return net.IP(rdata).String(), nil
	case TypeCNAME, TypeNS, TypePTR:
		name, _, err := decodeName(msg, rdataStart)
		return name, err
	case TypeMX:
		if rdlen < 3 {
			return "", ErrMalformed
		}
		pref := binary.BigEndian.Uint16(rdata[0:2])
		name, _, err := decodeName(msg, rdataStart+2)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d %s", pref, name), nil
	case TypeTXT:
		var parts []string
		for i := 0; i < len(rdata); {
			n := int(rdata[i])
			i++
			if i+n > len(rdata) {
				break
			}
			parts = append(parts, string(rdata[i:i+n]))
			i += n
		}
		return strings.Join(parts, ""), nil
	case TypeSOA:
		mname, next, err := decodeName(msg, rdataStart)
		if err != nil {
			return "", err
		}
		rname, next2, err := decodeName(msg, next)
		if err != nil {
			return "", err
		}
		rel := next2 - rdataStart
		if rel < 0 || rdlen < rel+20 {
			return fmt.Sprintf("%s %s", mname, rname), nil
		}
		serial := binary.BigEndian.Uint32(rdata[rel : rel+4])
		refresh := binary.BigEndian.Uint32(rdata[rel+4 : rel+8])
		retry := binary.BigEndian.Uint32(rdata[rel+8 : rel+12])
		expire := binary.BigEndian.Uint32(rdata[rel+12 : rel+16])
		minimum := binary.BigEndian.Uint32(rdata[rel+16 : rel+20])
		return fmt.Sprintf("%s %s %d %d %d %d %d", mname, rname, serial, refresh, retry, expire, minimum), nil
	case TypeCAA:
		if rdlen < 2 {
			return "", ErrMalformed
		}
		flags := rdata[0]
		tagLen := int(rdata[1])
		if 2+tagLen > len(rdata) {
			return "", ErrMalformed
		}
		tag := string(rdata[2 : 2+tagLen])
		value := string(rdata[2+tagLen:])
		return fmt.Sprintf("%d %s %q", flags, tag, value), nil
	default:
		return fmt.Sprintf("%x", rdata), nil
	}
}
