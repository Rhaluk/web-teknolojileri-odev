package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"yemekhane/ara"
	"yemekhane/kontrolcu"
	"yemekhane/oturum"
)

// rotalar bütün /api/ yollarını tanımlar. Apache /api/ ile başlayan
// istekleri bu uygulamaya iletir (WEBAPI modu).
//
// Bir isteğin yolu:  rota (burası) → kontrolcu → depo (veri katmanı) → JSON yanıt
func rotalar(k *kontrolcu.Kontrolcu, od *oturum.Depo) http.Handler {
	r := mux.NewRouter()
	r.Use(ara.Gunluk)

	api := r.PathPrefix("/api").Subrouter()

	// --- Herkese açık
	api.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		kontrolcu.JSON(w, http.StatusOK, map[string]string{"durum": "çalışıyor"})
	}).Methods(http.MethodGet)
	api.HandleFunc("/giris", k.Giris).Methods(http.MethodPost)
	api.HandleFunc("/cikis", k.Cikis).Methods(http.MethodPost)

	// --- Giriş yapmış kullanıcı
	kul := api.NewRoute().Subrouter()
	kul.Use(ara.OturumGerekli(od))
	kul.HandleFunc("/ben", k.Ben).Methods(http.MethodGet)

	// --- Yalnızca yönetici
	yon := api.PathPrefix("/yonetici").Subrouter()
	yon.Use(ara.OturumGerekli(od), ara.YoneticiGerekli)

	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		kontrolcu.Hata(w, http.StatusNotFound, "Böyle bir adres yok.")
	})
	return r
}
