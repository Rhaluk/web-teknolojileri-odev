package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"yemekhane/ara"
	"yemekhane/kontrolcu"
)

// rotalar bütün /api/ yollarını tanımlar. Apache /api/ ile başlayan
// istekleri bu uygulamaya iletir (WEBAPI modu).
func rotalar() http.Handler {
	r := mux.NewRouter()
	r.Use(ara.Gunluk)

	api := r.PathPrefix("/api").Subrouter()

	// Sağlık kontrolü: compose healthcheck ve başlangıç sayfası kullanır
	api.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		kontrolcu.JSON(w, http.StatusOK, map[string]string{"durum": "çalışıyor"})
	}).Methods(http.MethodGet)

	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		kontrolcu.Hata(w, http.StatusNotFound, "Böyle bir adres yok.")
	})
	return r
}
