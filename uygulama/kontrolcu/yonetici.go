package kontrolcu

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"yemekhane/depo"
	"yemekhane/modeller"
)

// Yemekler  GET /api/yonetici/yemekler  menü formundaki seçim listesi.
func (k *Kontrolcu) Yemekler(w http.ResponseWriter, _ *http.Request) {
	liste, err := k.depo.Yemekler()
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}

// MenuRaporu  GET /api/yonetici/menuler?baslangic=2026-11-12&bitis=2026-11-18
// Verilmezse bugünden 7 gün öncesi ile 7 gün sonrası arası.
func (k *Kontrolcu) MenuRaporu(w http.ResponseWriter, r *http.Request) {
	bugun := time.Now()
	baslangic, bitis := r.URL.Query().Get("baslangic"), r.URL.Query().Get("bitis")
	if baslangic == "" {
		baslangic = bugun.AddDate(0, 0, -7).Format(time.DateOnly)
	}
	if bitis == "" {
		bitis = bugun.AddDate(0, 0, 7).Format(time.DateOnly)
	}
	_, e1 := time.Parse(time.DateOnly, baslangic)
	_, e2 := time.Parse(time.DateOnly, bitis)
	if e1 != nil || e2 != nil {
		Hata(w, http.StatusBadRequest, "Tarihler YYYY-AA-GG biçiminde olmalıdır.")
		return
	}
	liste, err := k.depo.MenuRaporu(baslangic, bitis)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}

// MenuGetir  GET /api/yonetici/menuler/{id}
func (k *Kontrolcu) MenuGetir(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	m, err := k.depo.MenuGetir(uint(id))
	if errors.Is(err, depo.ErrBulunamadi) {
		Hata(w, http.StatusNotFound, "Menü bulunamadı.")
		return
	}
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, m)
}

// menuDogrula formdan gelen menüyü kaydetmeden önce denetler.
func menuDogrula(m *modeller.Menu) string {
	if m.YemekhaneID == 0 || m.OgunID == 0 {
		return "Yemekhane ve öğün seçilmelidir."
	}
	if _, err := time.Parse(time.DateOnly, m.Tarih); err != nil {
		return "Tarih YYYY-AA-GG biçiminde olmalıdır."
	}
	if m.OgrenciFiyati < 0 || m.PersonelFiyati < 0 {
		return "Fiyat negatif olamaz."
	}
	if len(m.YemekIDler) == 0 {
		return "Menüye en az bir yemek eklenmelidir."
	}
	return ""
}

// MenuOlustur  POST /api/yonetici/menuler
func (k *Kontrolcu) MenuOlustur(w http.ResponseWriter, r *http.Request) {
	var m modeller.Menu
	if !govdeOku(w, r, &m) {
		return
	}
	m.ID = 0
	if mesaj := menuDogrula(&m); mesaj != "" {
		Hata(w, http.StatusUnprocessableEntity, mesaj)
		return
	}
	if err := k.depo.MenuKaydet(&m); err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusCreated, m)
}

// MenuGuncelle  PUT /api/yonetici/menuler/{id}
func (k *Kontrolcu) MenuGuncelle(w http.ResponseWriter, r *http.Request) {
	var m modeller.Menu
	if !govdeOku(w, r, &m) {
		return
	}
	id, _ := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	m.ID = uint(id)
	if mesaj := menuDogrula(&m); mesaj != "" {
		Hata(w, http.StatusUnprocessableEntity, mesaj)
		return
	}
	err := k.depo.MenuKaydet(&m)
	if errors.Is(err, depo.ErrBulunamadi) {
		Hata(w, http.StatusNotFound, "Menü bulunamadı.")
		return
	}
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, m)
}
