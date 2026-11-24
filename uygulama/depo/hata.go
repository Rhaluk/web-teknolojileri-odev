// Package depo, uygulamanın veri katmanıdır. Bütün SQL ve GORM çağrıları
// burada durur. Okuma sorguları replikaya (Okuma), yazma işlemleri ve
// stored procedure çağrıları kaynağa (Yazma) gider.
package depo

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ErrBulunamadi aranan kayıt yoksa döner.
var ErrBulunamadi = errors.New("kayıt bulunamadı")

// IsKuraliHatasi, kullanıcıya olduğu gibi gösterilecek bir iş kuralı ihlalidir.
// Stored procedure'lerin SIGNAL SQLSTATE '45000' ile döndürdüğü Türkçe
// mesajlar bu tipe çevrilir.
type IsKuraliHatasi struct{ Mesaj string }

func (h *IsKuraliHatasi) Error() string { return h.Mesaj }

// MySQL hata kodları
const (
	kodSignal     = 1644 // SIGNAL SQLSTATE '45000' (procedure iş kuralı hatası)
	kodTekrarKayit = 1062 // UNIQUE ihlali
)

// cevir, sürücü hatasını depo hatasına çevirir. CHECK ihlali (3819) gibi
// programlama hataları çevrilmez; kontrolcü onları genel mesajla gizler.
func cevir(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrBulunamadi
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == kodSignal {
		return &IsKuraliHatasi{Mesaj: me.Message}
	}
	return err
}
