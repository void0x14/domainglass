package dg

import "github.com/void0x14/domainglass/internal/model"

// Bu dosya, istemcinin genel API yüzeyini tanımlar. İstemci, ham API tiplerini
// (model paketi) doğrudan döner; böylece çağıran taraf tek bir tip evreni kullanır.

// Ham API yanıt tipleri.
type (
	AlanAdiYaniti       = model.AlanAdiBilgisi
	GuvenlikYaniti      = model.GuvenlikBilgisi
	SiniflandirmaYaniti = model.Siniflandirma
	SiralamaYaniti      = model.Siralama
	TLSYaniti           = model.TLSBilgisi
	WebProfiliYaniti    = model.WebProfili
	IPYaniti            = model.IPBilgisi
	ASNYaniti           = model.ASNBilgisi
	TLDianaYaniti       = model.TLDGosterge
	AramaYaniti         = model.AramaBilgisi
	WHOISYaniti         = model.WHOISBilgisi
)
