// Package oturum, kullanıcı oturumlarını Redis'te tutar.
//
//	Anahtar: oturum:{token}   Değer: JSON (kullanıcı kimliği, rol, ad)
//	Süre: 30 dakika; her istekte yeniden 30 dakikaya uzatılır.
//
// Oturum uygulamanın belleğinde değil Redis'te durduğu için uygulama
// container'ı yeniden başlasa da kullanıcının oturumu düşmez.
package oturum

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const Sure = 30 * time.Minute

var ErrYok = errors.New("oturum yok ya da süresi dolmuş")

// Bilgi, oturumda saklanan kullanıcı bilgisidir.
type Bilgi struct {
	KullaniciID uint   `json:"kullanici_id"`
	Rol         string `json:"rol"`
	AdSoyad     string `json:"ad_soyad"`
}

type Depo struct{ rdb *redis.Client }

func Yeni(rdb *redis.Client) *Depo { return &Depo{rdb: rdb} }

func anahtar(token string) string { return "oturum:" + token }

// Olustur yeni bir rastgele token üretir ve oturumu Redis'e yazar.
func (d *Depo) Olustur(ctx context.Context, b Bilgi) (string, error) {
	ham := make([]byte, 32)
	if _, err := rand.Read(ham); err != nil {
		return "", err
	}
	token := hex.EncodeToString(ham)
	veri, _ := json.Marshal(b)
	return token, d.rdb.Set(ctx, anahtar(token), veri, Sure).Err()
}

// Getir oturumu okur ve süresini yeniden 30 dakikaya uzatır (GETEX).
func (d *Depo) Getir(ctx context.Context, token string) (*Bilgi, error) {
	veri, err := d.rdb.GetEx(ctx, anahtar(token), Sure).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrYok
	}
	if err != nil {
		return nil, err
	}
	var b Bilgi
	if err := json.Unmarshal(veri, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// Sil çıkış yapıldığında oturumu hemen kaldırır.
func (d *Depo) Sil(ctx context.Context, token string) error {
	return d.rdb.Del(ctx, anahtar(token)).Err()
}
