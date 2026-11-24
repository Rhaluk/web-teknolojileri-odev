package modeller

import "time"

// Degerlendirme, degerlendirmeler tablosu.
type Degerlendirme struct {
	ID          uint    `gorm:"column:id;primaryKey" json:"id"`
	KullaniciID uint    `gorm:"column:kullanici_id" json:"-"`
	MenuID      uint    `gorm:"column:menu_id" json:"menu_id"`
	Puan        uint8   `gorm:"column:puan" json:"puan"`
	Yorum       *string `gorm:"column:yorum" json:"yorum"`
}

func (Degerlendirme) TableName() string { return "degerlendirmeler" }

// DegerlendirmeSatiri, değerlendirme ekranındaki bir menü: kullanılmış ama
// henüz puanlanmamış (Puan nil) ya da daha önce puanlanmış.
type DegerlendirmeSatiri struct {
	MenuID     uint      `gorm:"column:menu_id" json:"menu_id"`
	Tarih      time.Time `gorm:"column:tarih" json:"-"`
	TarihMetin string    `gorm:"-" json:"tarih"`
	Ogun       string    `gorm:"column:ogun" json:"ogun"`
	Yemekhane  string    `gorm:"column:yemekhane" json:"yemekhane"`
	Puan       *uint8    `gorm:"column:puan" json:"puan"`
	Yorum      *string   `gorm:"column:yorum" json:"yorum"`
}
