package depo

import (
	"time"

	"yemekhane/modeller"
)

// Para hareketi yapan işlemler stored procedure ile ve KAYNAKTA (d.yazma)
// çalışır. Yordamlar kendi transaction'larını açtığı için burada ayrıca
// Begin/Commit yapılmaz. Dönen bakiye kaynaktan geldiği için günceldir.

// RezervasyonYap  sp_rezervasyon_yap(kullanici_id, menu_id)
func (d *Depo) RezervasyonYap(kullaniciID, menuID uint) (*modeller.RezervasyonSonucu, error) {
	var s modeller.RezervasyonSonucu
	err := d.yazma.Raw("CALL sp_rezervasyon_yap(?, ?)", kullaniciID, menuID).Scan(&s).Error
	return &s, cevir(err)
}

// RezervasyonIptal  sp_rezervasyon_iptal(kullanici_id, rezervasyon_id)
func (d *Depo) RezervasyonIptal(kullaniciID, rezervasyonID uint) (*modeller.IptalSonucu, error) {
	var s modeller.IptalSonucu
	err := d.yazma.Raw("CALL sp_rezervasyon_iptal(?, ?)", kullaniciID, rezervasyonID).Scan(&s).Error
	return &s, cevir(err)
}

// BakiyeYukle  sp_bakiye_yukle(kullanici_id, tutar)
func (d *Depo) BakiyeYukle(kullaniciID uint, tutar float64) (*modeller.YuklemeSonucu, error) {
	var s modeller.YuklemeSonucu
	err := d.yazma.Raw("CALL sp_bakiye_yukle(?, ?)", kullaniciID, tutar).Scan(&s).Error
	return &s, cevir(err)
}

// Rezervasyonlarim kullanıcının bütün rezervasyonlarını yeniden eskiye getirir (replika).
// iptal_edilebilir: aktif ve son saat geçmemiş; değerlendirildi: menüye oy vermiş.
func (d *Depo) Rezervasyonlarim(kullaniciID uint) ([]modeller.Rezervasyon, error) {
	var liste []modeller.Rezervasyon
	err := d.okuma.Raw(`
		SELECT r.id, r.menu_id, m.tarih, o.ad AS ogun, yh.ad AS yemekhane, r.ucret, r.durum,
		       r.durum = 'aktif' AND NOW() < TIMESTAMP(m.tarih, o.son_saat) AS iptal_edilebilir,
		       EXISTS (SELECT 1 FROM degerlendirmeler dg
		                WHERE dg.menu_id = r.menu_id AND dg.kullanici_id = r.kullanici_id) AS degerlendirildi
		  FROM rezervasyonlar r
		  JOIN menuler m       ON m.id  = r.menu_id
		  JOIN ogunler o       ON o.id  = m.ogun_id
		  JOIN yemekhaneler yh ON yh.id = m.yemekhane_id
		 WHERE r.kullanici_id = ?
		 ORDER BY m.tarih DESC, o.sira DESC`, kullaniciID).Scan(&liste).Error
	for i := range liste {
		liste[i].TarihMetin = liste[i].Tarih.Format(time.DateOnly)
	}
	return liste, cevir(err)
}

// BakiyeHareketleri kullanıcının son 50 bakiye hareketini getirir (replika).
func (d *Depo) BakiyeHareketleri(kullaniciID uint) ([]modeller.BakiyeHareketi, error) {
	var liste []modeller.BakiyeHareketi
	err := d.okuma.Where("kullanici_id = ?", kullaniciID).
		Order("olusturma DESC, id DESC").Limit(50).Find(&liste).Error
	return liste, cevir(err)
}
