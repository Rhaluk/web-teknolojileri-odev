// Package modeller, veritabanı tablolarının ve view'larının Go karşılıklarıdır.
// json etiketleri API yanıtındaki alan adlarını belirler.
package modeller

// Kullanici, kullanicilar tablosu.
type Kullanici struct {
	ID        uint    `gorm:"column:id" json:"id"`
	Ad        string  `gorm:"column:ad" json:"ad"`
	Soyad     string  `gorm:"column:soyad" json:"soyad"`
	Eposta    string  `gorm:"column:eposta" json:"eposta"`
	SifreHash string  `gorm:"column:sifre_hash" json:"-"` // yanıta asla girmez
	Rol       string  `gorm:"column:rol" json:"rol"`      // ogrenci | personel | yonetici
	OgrenciNo *string `gorm:"column:ogrenci_no" json:"ogrenci_no"`
	Bakiye    float64 `gorm:"column:bakiye" json:"bakiye"`
	Aktif     bool    `gorm:"column:aktif" json:"-"`
}

func (Kullanici) TableName() string { return "kullanicilar" }
