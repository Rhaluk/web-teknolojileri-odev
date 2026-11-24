package kontrolcu

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"yemekhane/ara"
	"yemekhane/depo"
)

// sayiParametre sorgu dizgisindeki pozitif tam sayıyı okur.
func sayiParametre(r *http.Request, ad string) (uint, bool) {
	n, err := strconv.ParseUint(r.URL.Query().Get(ad), 10, 32)
	return uint(n), err == nil && n > 0
}

// tarihParametre YYYY-AA-GG biçimindeki tarihi okur; verilmezse bugünü döner.
func tarihParametre(r *http.Request, ad string) (string, bool) {
	t := r.URL.Query().Get(ad)
	if t == "" {
		return time.Now().Format(time.DateOnly), true
	}
	_, err := time.Parse(time.DateOnly, t)
	return t, err == nil
}

// Yemekhaneler  GET /api/yemekhaneler
func (k *Kontrolcu) Yemekhaneler(w http.ResponseWriter, _ *http.Request) {
	liste, err := k.depo.Yemekhaneler()
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}

// GununMenusu  GET /api/menu?yemekhane_id=1&ogun_id=2&tarih=2026-11-24
func (k *Kontrolcu) GununMenusu(w http.ResponseWriter, r *http.Request) {
	yemekhaneID, ok1 := sayiParametre(r, "yemekhane_id")
	ogunID, ok2 := sayiParametre(r, "ogun_id")
	tarih, ok3 := tarihParametre(r, "tarih")
	if !ok1 || !ok2 || !ok3 {
		Hata(w, http.StatusBadRequest, "yemekhane_id, ogun_id ve geçerli bir tarih gerekli.")
		return
	}

	menu, err := k.depo.GununMenusu(yemekhaneID, ogunID, tarih, ara.Kullanici(r).Rol)
	if errors.Is(err, depo.ErrBulunamadi) {
		Hata(w, http.StatusNotFound, "Bu öğün için menü yok.")
		return
	}
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, menu)
}

// HaftalikMenu  GET /api/menuler/haftalik?yemekhane_id=1&baslangic=2026-11-24
func (k *Kontrolcu) HaftalikMenu(w http.ResponseWriter, r *http.Request) {
	yemekhaneID, ok1 := sayiParametre(r, "yemekhane_id")
	baslangic, ok2 := tarihParametre(r, "baslangic")
	if !ok1 || !ok2 {
		Hata(w, http.StatusBadRequest, "yemekhane_id ve geçerli bir başlangıç tarihi gerekli.")
		return
	}

	kul := ara.Kullanici(r)
	liste, err := k.depo.HaftalikMenu(yemekhaneID, baslangic, kul.KullaniciID, kul.Rol)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}
