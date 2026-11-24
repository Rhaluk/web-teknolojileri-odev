package depo

import (
	"gorm.io/gorm"

	"yemekhane/onbellek"
)

// Depo veri katmanının giriş noktasıdır. Her metot hangi sunucuya gittiğini
// açıkça seçer: d.okuma (replika) ya da d.yazma (kaynak). Günün menüsü
// önce d.onbellek'e (Redis) sorulur.
type Depo struct {
	yazma    *gorm.DB
	okuma    *gorm.DB
	onbellek *onbellek.Onbellek
}

func Yeni(yazma, okuma *gorm.DB, ob *onbellek.Onbellek) *Depo {
	return &Depo{yazma: yazma, okuma: okuma, onbellek: ob}
}
