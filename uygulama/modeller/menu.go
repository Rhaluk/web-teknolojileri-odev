package modeller

import "time"

// Yemekhane, yemekhaneler tablosu.
type Yemekhane struct {
	ID       uint   `gorm:"column:id" json:"id"`
	Ad       string `gorm:"column:ad" json:"ad"`
	Konum    string `gorm:"column:konum" json:"konum"`
	Kapasite int    `gorm:"column:kapasite" json:"kapasite"`
}

func (Yemekhane) TableName() string { return "yemekhaneler" }

// MenuDetaySatiri, v_menu_detay view'ının bir satırı (menüdeki bir yemek).
type MenuDetaySatiri struct {
	MenuID         uint      `gorm:"column:menu_id"`
	Tarih          time.Time `gorm:"column:tarih"`
	Yemekhane      string    `gorm:"column:yemekhane"`
	OgunID         uint      `gorm:"column:ogun_id"`
	Ogun           string    `gorm:"column:ogun"`
	RezervasyonSon time.Time `gorm:"column:rezervasyon_son"`
	OgrenciFiyati  float64   `gorm:"column:ogrenci_fiyati"`
	PersonelFiyati float64   `gorm:"column:personel_fiyati"`
	YemekID        uint      `gorm:"column:yemek_id"`
	Yemek          string    `gorm:"column:yemek"`
	Kategori       string    `gorm:"column:kategori"`
	Kalori         int       `gorm:"column:kalori"`
	Vejetaryen     bool      `gorm:"column:vejetaryen"`
	Alerjenler     *string   `gorm:"column:alerjenler"`
}

// MenuYemegi, API yanıtında menüdeki bir yemek.
type MenuYemegi struct {
	ID         uint     `json:"id"`
	Ad         string   `json:"ad"`
	Kategori   string   `json:"kategori"`
	Kalori     int      `json:"kalori"`
	Vejetaryen bool     `json:"vejetaryen"`
	Alerjenler []string `json:"alerjenler"`
}

// MenuIcerigi, günün menüsünün Redis'te önbelleğe alınan kısmı. Gün içinde
// değişmeyen bilgiler burada; kullanıcıya göre değişen fiyat ve her
// rezervasyonda değişen boş yer önbelleğe girmez.
type MenuIcerigi struct {
	MenuID         uint         `json:"menu_id"`
	Tarih          string       `json:"tarih"`
	Yemekhane      string       `json:"yemekhane"`
	Ogun           string       `json:"ogun"`
	RezervasyonSon time.Time    `json:"rezervasyon_son"`
	OgrenciFiyati  float64      `json:"ogrenci_fiyati"`
	PersonelFiyati float64      `json:"personel_fiyati"`
	ToplamKalori   int          `json:"toplam_kalori"`
	Yemekler       []MenuYemegi `json:"yemekler"`
}

// GununMenusu, günün menüsü ekranının yanıtı.
type GununMenusu struct {
	MenuID         uint         `json:"menu_id"`
	Tarih          string       `json:"tarih"`
	Yemekhane      string       `json:"yemekhane"`
	Ogun           string       `json:"ogun"`
	RezervasyonSon time.Time    `json:"rezervasyon_son"`
	Fiyat          float64      `json:"fiyat"` // kullanıcının rolüne göre
	BosYer         int          `json:"bos_yer"`
	Kapasite       int          `json:"kapasite"`
	ToplamKalori   int          `json:"toplam_kalori"`
	Yemekler       []MenuYemegi `json:"yemekler"`
}

// HaftalikMenu, haftalık menü tablosunun bir satırı.
type HaftalikMenu struct {
	MenuID            uint      `gorm:"column:menu_id" json:"menu_id"`
	Tarih             time.Time `gorm:"column:tarih" json:"-"`
	TarihMetin        string    `gorm:"-" json:"tarih"`
	OgunID            uint      `gorm:"column:ogun_id" json:"ogun_id"`
	Ogun              string    `gorm:"column:ogun" json:"ogun"`
	YemekListesi      string    `gorm:"column:yemek_listesi" json:"-"`
	Yemekler          []string  `gorm:"-" json:"yemekler"`
	ToplamKalori      int       `gorm:"column:toplam_kalori" json:"toplam_kalori"`
	Fiyat             float64   `gorm:"column:fiyat" json:"fiyat"`
	BosYer            int       `gorm:"column:bos_yer" json:"bos_yer"`
	RezervasyonSon    time.Time `gorm:"column:rezervasyon_son" json:"rezervasyon_son"`
	RezerveEdilebilir bool      `gorm:"column:rezerve_edilebilir" json:"rezerve_edilebilir"`
	RezervasyonDurumu *string   `gorm:"column:rezervasyon_durumu" json:"rezervasyon_durumu"` // aktif | kullanildi | iptal | null
}
