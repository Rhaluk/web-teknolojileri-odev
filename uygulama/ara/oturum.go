package ara

import (
	"context"
	"encoding/json"
	"net/http"

	"yemekhane/oturum"
)

// CerezAdi, oturum token'ının tarayıcıda tutulduğu çerezdir.
const CerezAdi = "oturum"

type baglamAnahtari struct{}

// Kullanici, OturumGerekli'nin isteğe eklediği oturum bilgisini döner.
func Kullanici(r *http.Request) *oturum.Bilgi {
	b, _ := r.Context().Value(baglamAnahtari{}).(*oturum.Bilgi)
	return b
}

func reddet(w http.ResponseWriter, durum int, mesaj string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(durum)
	_ = json.NewEncoder(w).Encode(map[string]string{"hata": mesaj})
}

// OturumGerekli, geçerli bir oturum çerezi olmayan isteği 401 ile reddeder.
// Geçerliyse oturumun süresi yenilenir ve kullanıcı bilgisi isteğe eklenir.
func OturumGerekli(od *oturum.Depo) func(http.Handler) http.Handler {
	return func(sonraki http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(CerezAdi)
			if err != nil {
				reddet(w, http.StatusUnauthorized, "Lütfen giriş yapın.")
				return
			}
			b, err := od.Getir(r.Context(), c.Value)
			if err != nil {
				reddet(w, http.StatusUnauthorized, "Oturumunuzun süresi doldu, lütfen yeniden giriş yapın.")
				return
			}
			sonraki.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), baglamAnahtari{}, b)))
		})
	}
}

// YoneticiGerekli, rolü yönetici olmayan kullanıcıyı 403 ile reddeder.
// OturumGerekli'den sonra kullanılmalıdır.
func YoneticiGerekli(sonraki http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if b := Kullanici(r); b == nil || b.Rol != "yonetici" {
			reddet(w, http.StatusForbidden, "Bu işlem için yönetici yetkisi gerekir.")
			return
		}
		sonraki.ServeHTTP(w, r)
	})
}
