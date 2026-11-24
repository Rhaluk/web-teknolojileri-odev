package depo

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"

	"yemekhane/modeller"
)

// KullanilmisMenuler kullanıcının yediği (durumu 'kullanildi') menüleri ve
// varsa verdiği puanı getirir (replika). Puanı boş olanlar değerlendirme bekler.
func (d *Depo) KullanilmisMenuler(kullaniciID uint) ([]modeller.DegerlendirmeSatiri, error) {
	var liste []modeller.DegerlendirmeSatiri
	err := d.okuma.Raw(`
		SELECT r.menu_id, m.tarih, o.ad AS ogun, yh.ad AS yemekhane, dg.puan, dg.yorum
		  FROM rezervasyonlar r
		  JOIN menuler m       ON m.id  = r.menu_id
		  JOIN ogunler o       ON o.id  = m.ogun_id
		  JOIN yemekhaneler yh ON yh.id = m.yemekhane_id
		  LEFT JOIN degerlendirmeler dg ON dg.menu_id = r.menu_id AND dg.kullanici_id = r.kullanici_id
		 WHERE r.kullanici_id = ? AND r.durum = 'kullanildi'
		 ORDER BY dg.puan IS NOT NULL, m.tarih DESC, o.sira DESC`, kullaniciID).Scan(&liste).Error
	for i := range liste {
		liste[i].TarihMetin = liste[i].Tarih.Format(time.DateOnly)
	}
	return liste, cevir(err)
}

// MenuyuKullandiMi kullanıcının bu menüdeki rezervasyonu kullanılmış mı?
// Kaynaktan okunur: kullanıcı turnikeden yeni geçtiyse replika henüz
// bilmiyor olabilir.
func (d *Depo) MenuyuKullandiMi(kullaniciID, menuID uint) (bool, error) {
	var sayi int64
	err := d.yazma.Table("rezervasyonlar").
		Where("kullanici_id = ? AND menu_id = ? AND durum = 'kullanildi'", kullaniciID, menuID).
		Count(&sayi).Error
	return sayi > 0, cevir(err)
}

// DegerlendirmeEkle puanı kaynağa yazar. Aynı menüye ikinci oy UNIQUE
// kısıtına takılır ve iş kuralı hatası olarak döner.
func (d *Depo) DegerlendirmeEkle(dg *modeller.Degerlendirme) error {
	err := d.yazma.Create(dg).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == kodTekrarKayit {
		return &IsKuraliHatasi{Mesaj: "Bu menüyü zaten değerlendirdiniz."}
	}
	return cevir(err)
}
