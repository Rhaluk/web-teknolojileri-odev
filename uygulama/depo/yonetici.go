package depo

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"yemekhane/modeller"
)

const kodYabanciAnahtar = 1452 // FOREIGN KEY ihlali (olmayan yemek/yemekhane/öğün)

// Yemekler bütün yemekleri kategori sırasıyla getirir (replika).
func (d *Depo) Yemekler() ([]modeller.Yemek, error) {
	var liste []modeller.Yemek
	err := d.okuma.Table("yemekler y").
		Select("y.id, y.ad, k.ad AS kategori, y.kalori, y.vejetaryen").
		Joins("JOIN yemek_kategorileri k ON k.id = y.kategori_id").
		Order("k.sira, y.ad").Scan(&liste).Error
	return liste, cevir(err)
}

// MenuRaporu iki tarih arasındaki menülerin doluluk ve ortalama puanını
// getirir (replika). v_menu_doluluk ve v_menu_puan view'ları birleştirilir.
func (d *Depo) MenuRaporu(baslangic, bitis string) ([]modeller.MenuRaporu, error) {
	var liste []modeller.MenuRaporu
	err := d.okuma.Table("v_menu_doluluk d").
		Select("d.menu_id, d.tarih, d.yemekhane, d.ogun, d.kapasite, d.rezerve, d.doluluk_yuzde, p.oy_sayisi, p.ortalama_puan").
		Joins("JOIN v_menu_puan p ON p.menu_id = d.menu_id").
		Where("d.tarih BETWEEN ? AND ?", baslangic, bitis).
		Order("d.tarih DESC, d.yemekhane_id, d.ogun_id").Scan(&liste).Error
	for i := range liste {
		liste[i].TarihMetin = liste[i].Tarih.Format(time.DateOnly)
	}
	return liste, cevir(err)
}

// MenuGetir düzenleme formu için menüyü ve yemeklerinin kimliklerini getirir (replika).
func (d *Depo) MenuGetir(id uint) (*modeller.Menu, error) {
	var m modeller.Menu
	err := d.okuma.Select("id, yemekhane_id, DATE_FORMAT(tarih, '%Y-%m-%d') AS tarih, ogun_id, ogrenci_fiyati, personel_fiyati").
		First(&m, id).Error
	if err != nil {
		return nil, cevir(err)
	}
	err = d.okuma.Model(&modeller.MenuYemek{}).Where("menu_id = ?", id).Pluck("yemek_id", &m.YemekIDler).Error
	return &m, cevir(err)
}

// MenuKaydet menüyü oluşturur (m.ID == 0) ya da günceller. Menü satırı ve
// yemek listesi aynı transaction içinde kaynağa yazılır; biri başarısız
// olursa hiçbiri kaydedilmez.
func (d *Depo) MenuKaydet(m *modeller.Menu) error {
	err := d.yazma.Transaction(func(tx *gorm.DB) error {
		if m.ID == 0 {
			if err := tx.Create(m).Error; err != nil {
				return err
			}
		} else {
			var sayi int64
			if err := tx.Model(&modeller.Menu{}).Where("id = ?", m.ID).Count(&sayi).Error; err != nil {
				return err
			}
			if sayi == 0 {
				return ErrBulunamadi
			}
			err := tx.Model(&modeller.Menu{ID: m.ID}).Updates(map[string]any{
				"yemekhane_id": m.YemekhaneID, "tarih": m.Tarih, "ogun_id": m.OgunID,
				"ogrenci_fiyati": m.OgrenciFiyati, "personel_fiyati": m.PersonelFiyati,
			}).Error
			if err != nil {
				return err
			}
			// Yemek listesi baştan yazılır: eskiler silinir, yeniler eklenir
			if err := tx.Where("menu_id = ?", m.ID).Delete(&modeller.MenuYemek{}).Error; err != nil {
				return err
			}
		}

		satirlar := make([]modeller.MenuYemek, len(m.YemekIDler))
		for i, yid := range m.YemekIDler {
			satirlar[i] = modeller.MenuYemek{MenuID: m.ID, YemekID: yid}
		}
		return tx.Create(&satirlar).Error
	})

	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case kodTekrarKayit:
			return &IsKuraliHatasi{Mesaj: "Bu yemekhane, tarih ve öğün için zaten bir menü var."}
		case kodYabanciAnahtar:
			return &IsKuraliHatasi{Mesaj: "Seçilen yemekhane, öğün ya da yemeklerden biri bulunamadı."}
		}
	}
	return cevir(err)
}
