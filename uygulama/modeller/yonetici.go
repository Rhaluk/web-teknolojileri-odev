package modeller

import "time"

// Menu, menuler tablosu.
type Menu struct {
	ID             uint    `gorm:"column:id;primaryKey" json:"id"`
	YemekhaneID    uint    `gorm:"column:yemekhane_id" json:"yemekhane_id"`
	Tarih          string  `gorm:"column:tarih" json:"tarih"` // YYYY-AA-GG
	OgunID         uint    `gorm:"column:ogun_id" json:"ogun_id"`
	OgrenciFiyati  float64 `gorm:"column:ogrenci_fiyati" json:"ogrenci_fiyati"`
	PersonelFiyati float64 `gorm:"column:personel_fiyati" json:"personel_fiyati"`
	YemekIDler     []uint  `gorm:"-" json:"yemek_idler"`
}

func (Menu) TableName() string { return "menuler" }

// MenuYemek, menu_yemekleri tablosu (menü ile yemek arasındaki çoka çok ilişki).
type MenuYemek struct {
	MenuID  uint `gorm:"column:menu_id;primaryKey"`
	YemekID uint `gorm:"column:yemek_id;primaryKey"`
}

func (MenuYemek) TableName() string { return "menu_yemekleri" }

// Yemek, yönetici formundaki yemek seçimi için.
type Yemek struct {
	ID         uint   `gorm:"column:id" json:"id"`
	Ad         string `gorm:"column:ad" json:"ad"`
	Kategori   string `gorm:"column:kategori" json:"kategori"`
	Kalori     int    `gorm:"column:kalori" json:"kalori"`
	Vejetaryen bool   `gorm:"column:vejetaryen" json:"vejetaryen"`
}

// MenuRaporu, menü bazında doluluk ve ortalama puan.
type MenuRaporu struct {
	MenuID       uint      `gorm:"column:menu_id" json:"menu_id"`
	Tarih        time.Time `gorm:"column:tarih" json:"-"`
	TarihMetin   string    `gorm:"-" json:"tarih"`
	Yemekhane    string    `gorm:"column:yemekhane" json:"yemekhane"`
	Ogun         string    `gorm:"column:ogun" json:"ogun"`
	Kapasite     int       `gorm:"column:kapasite" json:"kapasite"`
	Rezerve      int       `gorm:"column:rezerve" json:"rezerve"`
	DolulukYuzde float64   `gorm:"column:doluluk_yuzde" json:"doluluk_yuzde"`
	OySayisi     int       `gorm:"column:oy_sayisi" json:"oy_sayisi"`
	OrtalamaPuan *float64  `gorm:"column:ortalama_puan" json:"ortalama_puan"` // oy yoksa null
}
