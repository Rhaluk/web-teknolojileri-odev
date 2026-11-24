package kontrolcu

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"yemekhane/ara"
	"yemekhane/depo"
	"yemekhane/oturum"
)

// sahteHash, olmayan bir e-postada da bcrypt karşılaştırması yapılsın diye
// kullanılır; böylece yanıt süresinden e-postanın kayıtlı olup olmadığı
// anlaşılamaz.
var sahteHash, _ = bcrypt.GenerateFromPassword([]byte("kayitli-olmayan-kullanici"), bcrypt.DefaultCost)

// Giris  POST /api/giris  {"eposta": "...", "sifre": "..."}
func (k *Kontrolcu) Giris(w http.ResponseWriter, r *http.Request) {
	var istek struct {
		Eposta string `json:"eposta"`
		Sifre  string `json:"sifre"`
	}
	if !govdeOku(w, r, &istek) {
		return
	}
	if istek.Eposta == "" || istek.Sifre == "" {
		Hata(w, http.StatusBadRequest, "E-posta ve şifre gerekli.")
		return
	}

	kul, err := k.depo.KullaniciEpostaIle(istek.Eposta)
	if err != nil && !errors.Is(err, depo.ErrBulunamadi) {
		DepoHatasi(w, err)
		return
	}
	hash := sahteHash
	if err == nil {
		hash = []byte(kul.SifreHash)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(istek.Sifre)) != nil || err != nil {
		Hata(w, http.StatusUnauthorized, "E-posta ya da şifre hatalı.")
		return
	}
	if !kul.Aktif {
		Hata(w, http.StatusForbidden, "Hesabınız pasif durumda.")
		return
	}

	token, err := k.oturum.Olustur(r.Context(), oturum.Bilgi{
		KullaniciID: kul.ID, Rol: kul.Rol, AdSoyad: kul.Ad + " " + kul.Soyad,
	})
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     ara.CerezAdi,
		Value:    token,
		Path:     "/",
		HttpOnly: true, // JavaScript okuyamaz
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(oturum.Sure.Seconds()),
	})
	JSON(w, http.StatusOK, kul)
}

// Cikis  POST /api/cikis  oturumu Redis'ten siler, çerezi temizler.
func (k *Kontrolcu) Cikis(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(ara.CerezAdi); err == nil {
		_ = k.oturum.Sil(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: ara.CerezAdi, Value: "", Path: "/", MaxAge: -1})
	JSON(w, http.StatusOK, map[string]string{"mesaj": "Çıkış yapıldı."})
}

// Ben  GET /api/ben  giriş yapmış kullanıcının bilgisi ve bakiyesi.
// ?taze=1 verilirse bakiye kaynaktan okunur (işlemden hemen sonra).
func (k *Kontrolcu) Ben(w http.ResponseWriter, r *http.Request) {
	kul, err := k.depo.KullaniciGetir(ara.Kullanici(r).KullaniciID, r.URL.Query().Get("taze") == "1")
	if err != nil {
		DepoHatasi(w, err)
		return
	}
	JSON(w, http.StatusOK, kul)
}
