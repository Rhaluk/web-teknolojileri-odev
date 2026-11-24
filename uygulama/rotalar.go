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
	kul.HandleFunc("/yemekhaneler", k.Yemekhaneler).Methods(http.MethodGet)
	kul.HandleFunc("/menu", k.GununMenusu).Methods(http.MethodGet)
	kul.HandleFunc("/menuler/haftalik", k.HaftalikMenu).Methods(http.MethodGet)
	kul.HandleFunc("/rezervasyonlar", k.Rezervasyonlarim).Methods(http.MethodGet)
	kul.HandleFunc("/rezervasyonlar", k.RezervasyonYap).Methods(http.MethodPost)
	kul.HandleFunc("/rezervasyonlar/{id:[0-9]+}", k.RezervasyonIptal).Methods(http.MethodDelete)
	kul.HandleFunc("/bakiye", k.Bakiye).Methods(http.MethodGet)
	kul.HandleFunc("/bakiye/yukle", k.BakiyeYukle).Methods(http.MethodPost)
	kul.HandleFunc("/degerlendirmeler", k.Degerlendirmeler).Methods(http.MethodGet)
	kul.HandleFunc("/degerlendirmeler", k.DegerlendirmeYap).Methods(http.MethodPost)

	// --- Yalnızca yönetici
	yon := api.PathPrefix("/yonetici").Subrouter()
	yon.Use(ara.OturumGerekli(od), ara.YoneticiGerekli)
	yon.HandleFunc("/yemekler", k.Yemekler).Methods(http.MethodGet)
	yon.HandleFunc("/menuler", k.MenuRaporu).Methods(http.MethodGet)
	yon.HandleFunc("/menuler", k.MenuOlustur).Methods(http.MethodPost)
	yon.HandleFunc("/menuler/{id:[0-9]+}", k.MenuGetir).Methods(http.MethodGet)
	yon.HandleFunc("/menuler/{id:[0-9]+}", k.MenuGuncelle).Methods(http.MethodPut)

	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		kontrolcu.Hata(w, http.StatusNotFound, "Böyle bir adres yok.")
	})
	return r
}
