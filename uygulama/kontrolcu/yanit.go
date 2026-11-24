// Package kontrolcu, HTTP isteklerini karşılayan Controller katmanıdır.
// Burada SQL yazılmaz; veri depo paketinden istenir, sonuç JSON olarak döner.
package kontrolcu

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"yemekhane/depo"
)

// JSON, v değerini verilen durum koduyla JSON olarak yazar.
func JSON(w http.ResponseWriter, durum int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(durum)
	_ = json.NewEncoder(w).Encode(v)
}

// Hata, arayüzün gösterebileceği {"hata": "..."} yanıtını yazar.
func Hata(w http.ResponseWriter, durum int, mesaj string) {
	JSON(w, durum, map[string]string{"hata": mesaj})
}

// DepoHatasi, veri katmanından gelen hatayı uygun HTTP yanıtına çevirir.
// Stored procedure'ün Türkçe mesajı (ör. "Bakiye yetersiz.") olduğu gibi
// kullanıcıya gider; beklenmeyen hatalar ise günlüğe yazılıp genel bir
// mesajla gizlenir.
func DepoHatasi(w http.ResponseWriter, err error) {
	var ih *depo.IsKuraliHatasi
	switch {
	case errors.As(err, &ih):
		Hata(w, http.StatusUnprocessableEntity, ih.Mesaj)
	case errors.Is(err, depo.ErrBulunamadi):
		Hata(w, http.StatusNotFound, "Kayıt bulunamadı.")
	default:
		log.Printf("beklenmeyen hata: %v", err)
		Hata(w, http.StatusInternalServerError, "Beklenmeyen bir hata oluştu, lütfen tekrar deneyin.")
	}
}

// govdeOku istek gövdesindeki JSON'u hedef'e çözer.
func govdeOku(w http.ResponseWriter, r *http.Request, hedef any) bool {
	if err := json.NewDecoder(r.Body).Decode(hedef); err != nil {
		Hata(w, http.StatusBadRequest, "İstek gövdesi okunamadı.")
		return false
	}
	return true
}
