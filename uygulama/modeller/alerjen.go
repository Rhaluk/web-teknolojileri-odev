package modeller

// Alerjen, alerjenler tablosu.
type Alerjen struct {
	ID uint   `gorm:"column:id" json:"id"`
	Ad string `gorm:"column:ad" json:"ad"`
}

func (Alerjen) TableName() string { return "alerjenler" }

// KullaniciAlerjeni, kullanici_alerjenleri tablosu (Aşama 4).
type KullaniciAlerjeni struct {
	KullaniciID uint `gorm:"column:kullanici_id;primaryKey"`
	AlerjenID   uint `gorm:"column:alerjen_id;primaryKey"`
}

func (KullaniciAlerjeni) TableName() string { return "kullanici_alerjenleri" }
