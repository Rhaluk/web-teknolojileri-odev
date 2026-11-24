package kontrolcu

import (
	"net/http"

	"yemekhane/ara"
)

// Bakiye  GET /api/bakiye  güncel bakiye ve hareket geçmişi (replika).
func (k *Kontrolcu) Bakiye(w http.ResponseWriter, r *http.Request) {
	id := ara.Kullanici(r).KullaniciID
	kul, err := k.depo.KullaniciGetir(id, false)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	hareketler, err := k.depo.BakiyeHareketleri(id)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"bakiye": kul.Bakiye, "hareketler": hareketler})
}

// BakiyeYukle  POST /api/bakiye/yukle  {"tutar": 100}
// 10–1000 TL sınırını stored procedure denetler; mesajı olduğu gibi döner.
func (k *Kontrolcu) BakiyeYukle(w http.ResponseWriter, r *http.Request) {
	var istek struct {
		Tutar float64 `json:"tutar"`
	}
	if !govdeOku(w, r, &istek) {
		return
	}
	sonuc, err := k.depo.BakiyeYukle(ara.Kullanici(r).KullaniciID, istek.Tutar)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusCreated, sonuc)
}
