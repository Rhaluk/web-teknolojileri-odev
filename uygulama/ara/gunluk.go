// Package ara, rotalardan önce çalışan ara katmanları (middleware) içerir.
package ara

import (
	"log"
	"net/http"
	"time"
)

// durumYakalayici, yazılan HTTP durum kodunu günlüğe yazabilmek için saklar.
type durumYakalayici struct {
	http.ResponseWriter
	durum int
}

func (d *durumYakalayici) WriteHeader(kod int) {
	d.durum = kod
	d.ResponseWriter.WriteHeader(kod)
}

// Gunluk her isteği yöntem, yol, durum kodu ve süreyle günlüğe yazar:
//
//	GET /api/menu 200 3ms
func Gunluk(sonraki http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baslangic := time.Now()
		dy := &durumYakalayici{ResponseWriter: w, durum: http.StatusOK}
		sonraki.ServeHTTP(dy, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.RequestURI(), dy.durum, time.Since(baslangic).Round(time.Millisecond))
	})
}
