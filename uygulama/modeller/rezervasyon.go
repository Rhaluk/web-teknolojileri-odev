package modeller

import "time"

// Rezervasyon, "rezervasyonlarım" listesinin bir satırı.
type Rezervasyon struct {
	ID              uint      `gorm:"column:id" json:"id"`
	MenuID          uint      `gorm:"column:menu_id" json:"menu_id"`
	Tarih           time.Time `gorm:"column:tarih" json:"-"`
	TarihMetin      string    `gorm:"-" json:"tarih"`
	Ogun            string    `gorm:"column:ogun" json:"ogun"`
	Yemekhane       string    `gorm:"column:yemekhane" json:"yemekhane"`
	Ucret           float64   `gorm:"column:ucret" json:"ucret"`
	Durum           string    `gorm:"column:durum" json:"durum"` // aktif | kullanildi | iptal
	IptalEdilebilir bool      `gorm:"column:iptal_edilebilir" json:"iptal_edilebilir"`
	Degerlendirildi bool      `gorm:"column:degerlendirildi" json:"degerlendirildi"`
}

// RezervasyonSonucu, sp_rezervasyon_yap'ın döndürdüğü satır.
type RezervasyonSonucu struct {
	RezervasyonID uint    `gorm:"column:rezervasyon_id" json:"rezervasyon_id"`
	Ucret         float64 `gorm:"column:ucret" json:"ucret"`
	Bakiye        float64 `gorm:"column:bakiye" json:"bakiye"`
}

// IptalSonucu, sp_rezervasyon_iptal'in döndürdüğü satır.
type IptalSonucu struct {
	RezervasyonID uint    `gorm:"column:rezervasyon_id" json:"rezervasyon_id"`
	Iade          float64 `gorm:"column:iade" json:"iade"`
	Bakiye        float64 `gorm:"column:bakiye" json:"bakiye"`
}

// BakiyeHareketi, bakiye_hareketleri tablosu.
type BakiyeHareketi struct {
	ID        uint      `gorm:"column:id" json:"id"`
	Tur       string    `gorm:"column:tur" json:"tur"` // yukleme | rezervasyon | iade
	Tutar     float64   `gorm:"column:tutar" json:"tutar"`
	Aciklama  string    `gorm:"column:aciklama" json:"aciklama"`
	Olusturma time.Time `gorm:"column:olusturma" json:"olusturma"`
}

func (BakiyeHareketi) TableName() string { return "bakiye_hareketleri" }

// YuklemeSonucu, sp_bakiye_yukle'nin döndürdüğü satır.
type YuklemeSonucu struct {
	HareketID uint    `gorm:"column:hareket_id" json:"hareket_id"`
	Bakiye    float64 `gorm:"column:bakiye" json:"bakiye"`
}
