package depo

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"yemekhane/modeller"
)

// Alerjenler bütün alerjenleri ada göre getirir (replika).
func (d *Depo) Alerjenler() ([]modeller.Alerjen, error) {
	var liste []modeller.Alerjen
	err := d.okuma.Order("ad").Find(&liste).Error
	return liste, cevir(err)
}

// KullaniciAlerjenIDleri kullanıcının profilinde seçtiği alerjenler (replika).
func (d *Depo) KullaniciAlerjenIDleri(kullaniciID uint) ([]uint, error) {
	idler := []uint{}
	err := d.okuma.Model(&modeller.KullaniciAlerjeni{}).
		Where("kullanici_id = ?", kullaniciID).Order("alerjen_id").
		Pluck("alerjen_id", &idler).Error
	return idler, cevir(err)
}

// KullaniciAlerjenleriKaydet profildeki alerjen listesini baştan yazar
// (kaynak, tek transaction: eskiler silinir, yeniler eklenir).
func (d *Depo) KullaniciAlerjenleriKaydet(kullaniciID uint, alerjenIDler []uint) error {
	err := d.yazma.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("kullanici_id = ?", kullaniciID).Delete(&modeller.KullaniciAlerjeni{}).Error; err != nil {
			return err
		}
		if len(alerjenIDler) == 0 {
			return nil
		}
		satirlar := make([]modeller.KullaniciAlerjeni, len(alerjenIDler))
		for i, id := range alerjenIDler {
			satirlar[i] = modeller.KullaniciAlerjeni{KullaniciID: kullaniciID, AlerjenID: id}
		}
		return tx.Create(&satirlar).Error
	})
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case kodYabanciAnahtar:
			return &IsKuraliHatasi{Mesaj: "Seçilen alerjenlerden biri bulunamadı."}
		case kodTekrarKayit:
			return &IsKuraliHatasi{Mesaj: "Aynı alerjen birden fazla kez seçilmiş."}
		}
	}
	return cevir(err)
}
