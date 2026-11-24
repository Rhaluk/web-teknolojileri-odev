package kontrolcu

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"yemekhane/ara"
)

// RezervasyonYap  POST /api/rezervasyonlar  {"menu_id": 68, "porsiyon": "yarim"}
// porsiyon verilmezse "tam". Kahvaltıda yarım porsiyon yoktur kuralını ve
// yarım ücreti (%60) stored procedure uygular; mesajı olduğu gibi döner.
// Yanıttaki bakiye stored procedure'den, yani kaynaktan gelir; arayüz bunu
// gösterir. Replikadan okunsaydı replikasyon gecikmesi yüzünden eski bakiye
// görünebilirdi.
func (k *Kontrolcu) RezervasyonYap(w http.ResponseWriter, r *http.Request) {
	var istek struct {
		MenuID   uint   `json:"menu_id"`
		Porsiyon string `json:"porsiyon"`
	}
	if !govdeOku(w, r, &istek) {
		return
	}
	if istek.MenuID == 0 {
		Hata(w, http.StatusBadRequest, "menu_id gerekli.")
		return
	}
	switch istek.Porsiyon {
	case "":
		istek.Porsiyon = "tam"
	case "tam", "yarim":
	default:
		Hata(w, http.StatusUnprocessableEntity, "Geçersiz porsiyon seçimi.")
		return
	}
	sonuc, err := k.depo.RezervasyonYap(ara.Kullanici(r).KullaniciID, istek.MenuID, istek.Porsiyon)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusCreated, sonuc)
}

// RezervasyonIptal  DELETE /api/rezervasyonlar/{id}
func (k *Kontrolcu) RezervasyonIptal(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		Hata(w, http.StatusBadRequest, "Geçersiz rezervasyon numarası.")
		return
	}
	sonuc, err := k.depo.RezervasyonIptal(ara.Kullanici(r).KullaniciID, uint(id))
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, sonuc)
}

// Rezervasyonlarim  GET /api/rezervasyonlar
func (k *Kontrolcu) Rezervasyonlarim(w http.ResponseWriter, r *http.Request) {
	liste, err := k.depo.Rezervasyonlarim(ara.Kullanici(r).KullaniciID)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}
