package depo

import "yemekhane/modeller"

// KullaniciEpostaIle giriş için kullanıcıyı e-postasıyla bulur (replika).
func (d *Depo) KullaniciEpostaIle(eposta string) (*modeller.Kullanici, error) {
	var k modeller.Kullanici
	err := d.okuma.Where("eposta = ?", eposta).First(&k).Error
	return &k, cevir(err)
}

// KullaniciGetir kullanıcıyı kimliğiyle getirir.
// taze true ise kaynaktan okunur: rezervasyon ya da yüklemeden hemen sonra
// replika birkaç yüz milisaniye geride kalabilir ve eski bakiyeyi gösterebilir.
func (d *Depo) KullaniciGetir(id uint, taze bool) (*modeller.Kullanici, error) {
	db := d.okuma
	if taze {
		db = d.yazma
	}
	var k modeller.Kullanici
	err := db.First(&k, id).Error
	return &k, cevir(err)
}
