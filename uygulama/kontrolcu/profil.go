package kontrolcu

import (
	"net/http"

	"yemekhane/ara"
)

// Alerjenler  GET /api/alerjenler  profil sayfasındaki seçim listesi.
func (k *Kontrolcu) Alerjenler(w http.ResponseWriter, _ *http.Request) {
	liste, err := k.depo.Alerjenler()
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, liste)
}

// ProfilAlerjenleri  GET /api/profil/alerjenler  → {"alerjen_idler": [1, 4]}
func (k *Kontrolcu) ProfilAlerjenleri(w http.ResponseWriter, r *http.Request) {
	idler, err := k.depo.KullaniciAlerjenIDleri(ara.Kullanici(r).KullaniciID)
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"alerjen_idler": idler})
}

// ProfilAlerjenleriKaydet  PUT /api/profil/alerjenler  {"alerjen_idler": [1, 4]}
// Boş liste "alerjenim yok" demektir.
func (k *Kontrolcu) ProfilAlerjenleriKaydet(w http.ResponseWriter, r *http.Request) {
	var istek struct {
		AlerjenIDler []uint `json:"alerjen_idler"`
	}
	if !govdeOku(w, r, &istek) {
		return
	}
	if err := k.depo.KullaniciAlerjenleriKaydet(ara.Kullanici(r).KullaniciID, istek.AlerjenIDler); err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"alerjen_idler": istek.AlerjenIDler})
}
