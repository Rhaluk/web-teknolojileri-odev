// Package onbellek, günün menüsünü Redis'te tutar.
//
//	Anahtar: menu:{yemekhane_id}:{tarih}:{ogun_id}   Değer: JSON
//	Süre: o günün sonuna (gece yarısına) kadar
//
// Menü içeriği gün içinde nadiren değişir ve çok okunur; her istekte
// veritabanına gitmek yerine ilk istekten sonra Redis'ten verilir. Yönetici
// menüyü değiştirdiğinde anahtar silinir (depo.MenuKaydet).
//
// Redis'e ulaşılamazsa uygulama durmaz: hata günlüğe yazılır ve veri
// veritabanından okunmaya devam eder.
package onbellek

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Onbellek struct{ rdb *redis.Client }

func Yeni(rdb *redis.Client) *Onbellek { return &Onbellek{rdb: rdb} }

// MenuAnahtari  menu:1:2026-11-25:2
func MenuAnahtari(yemekhaneID uint, tarih string, ogunID uint) string {
	return fmt.Sprintf("menu:%d:%s:%d", yemekhaneID, tarih, ogunID)
}

// gunSonunaKadar şu andan bu gecenin 00:00'ına kalan süre.
func gunSonunaKadar() time.Duration {
	simdi := time.Now()
	yil, ay, gun := simdi.Date()
	return time.Date(yil, ay, gun+1, 0, 0, 0, 0, simdi.Location()).Sub(simdi)
}

// Getir anahtar Redis'te varsa değeri hedef'e açar ve true döner.
func (o *Onbellek) Getir(ctx context.Context, anahtar string, hedef any) bool {
	veri, err := o.rdb.Get(ctx, anahtar).Bytes()
	if errors.Is(err, redis.Nil) {
		return false
	}
	if err != nil {
		log.Printf("önbellek: %s okunamadı: %v", anahtar, err)
		return false
	}
	if err := json.Unmarshal(veri, hedef); err != nil {
		log.Printf("önbellek: %s bozuk, yok sayıldı: %v", anahtar, err)
		return false
	}
	return true
}

// Yaz değeri gün sonuna kadar saklar ve kalan süreyi döner.
func (o *Onbellek) Yaz(ctx context.Context, anahtar string, deger any) time.Duration {
	veri, err := json.Marshal(deger)
	if err != nil {
		log.Printf("önbellek: %s yazılamadı: %v", anahtar, err)
		return 0
	}
	sure := gunSonunaKadar()
	if err := o.rdb.Set(ctx, anahtar, veri, sure).Err(); err != nil {
		log.Printf("önbellek: %s yazılamadı: %v", anahtar, err)
		return 0
	}
	return sure
}

// Sil anahtarları kaldırır (yönetici menüyü değiştirdiğinde).
func (o *Onbellek) Sil(ctx context.Context, anahtarlar ...string) {
	n, err := o.rdb.Del(ctx, anahtarlar...).Result()
	if err != nil {
		log.Printf("önbellek: %v silinemedi: %v", anahtarlar, err)
		return
	}
	log.Printf("önbellek: %v geçersiz kılındı (%d anahtar silindi)", anahtarlar, n)
}
