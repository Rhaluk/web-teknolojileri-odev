// Yemekhane Otomasyonu · Web API
// Yalnızca JSON döner. Arayüz frontend container'ındadır; Apache /api/
// isteklerini buraya (uygulama:8000) iletir.
package main

import (
	"log"
	"net/http"
	"time"
	_ "time/tzdata" // container'da saat dilimi verisi olmasa da Europe/Istanbul çalışsın

	"yemekhane/veritabani"
)

func main() {
	b, err := veritabani.Ac()
	if err != nil {
		log.Fatalf("bağlantılar kurulamadı: %v", err)
	}
	log.Println("kaynak, replika ve redis bağlantıları hazır")
	_ = b // sonraki adımlarda kontrolcülere verilecek

	sunucu := &http.Server{
		Addr:              ":8000",
		Handler:           rotalar(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("uygulama :8000 portunu dinliyor")
	log.Fatal(sunucu.ListenAndServe())
}
