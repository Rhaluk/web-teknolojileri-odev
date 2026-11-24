package depo

import "gorm.io/gorm"

// Depo veri katmanının giriş noktasıdır. Her metot hangi sunucuya gittiğini
// açıkça seçer: d.okuma (replika) ya da d.yazma (kaynak).
type Depo struct {
	yazma *gorm.DB
	okuma *gorm.DB
}

func Yeni(yazma, okuma *gorm.DB) *Depo {
	return &Depo{yazma: yazma, okuma: okuma}
}
