package kontrolcu

import (
	"net/http"
	"strings"

	"yemekhane/ara"
	"yemekhane/modeller"
)

// Degerlendirmeler  GET /api/degerlendirmeler
// Kullanıcının yediği menüler; puan alanı boş olanlar değerlendirme bekler.
func (k *Kontrolcu) Degerlendirmeler(w http.ResponseWriter, r *http.Request) {
	liste, err := k.depo.KullanilmisMenuler(ara.Kullanici(r).KullaniciID)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}

// DegerlendirmeYap  POST /api/degerlendirmeler  {"menu_id": 32, "puan": 4, "yorum": "..."}
// İş kuralları: puan 1–5 arası, yorum en fazla 500 karakter ve yalnızca
// kullanılmış (yenmiş) bir menü değerlendirilebilir.
func (k *Kontrolcu) DegerlendirmeYap(w http.ResponseWriter, r *http.Request) {
	var istek struct {
		MenuID uint   `json:"menu_id"`
		Puan   uint8  `json:"puan"`
		Yorum  string `json:"yorum"`
	}
	if !govdeOku(w, r, &istek) {
		return
	}
	if istek.Puan < 1 || istek.Puan > 5 {
		Hata(w, http.StatusUnprocessableEntity, "Puan 1 ile 5 arasında olmalıdır.")
		return
	}
	istek.Yorum = strings.TrimSpace(istek.Yorum)
	if len([]rune(istek.Yorum)) > 500 {
		Hata(w, http.StatusUnprocessableEntity, "Yorum en fazla 500 karakter olabilir.")
		return
	}

	kulID := ara.Kullanici(r).KullaniciID
	kullandi, err := k.depo.MenuyuKullandiMi(kulID, istek.MenuID)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	if !kullandi {
		Hata(w, http.StatusUnprocessableEntity, "Yalnızca yemek yediğiniz menüleri değerlendirebilirsiniz.")
		return
	}

	dg := &modeller.Degerlendirme{KullaniciID: kulID, MenuID: istek.MenuID, Puan: istek.Puan}
	if istek.Yorum != "" {
		dg.Yorum = &istek.Yorum
	}
	if err := k.depo.DegerlendirmeEkle(dg); err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusCreated, dg)
}
