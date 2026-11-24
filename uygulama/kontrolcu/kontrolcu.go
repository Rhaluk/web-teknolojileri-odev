package kontrolcu

import (
	"yemekhane/depo"
	"yemekhane/oturum"
)

// Kontrolcu bütün rota işleyicilerinin bağlı olduğu yapıdır.
// Veriye yalnızca depo üzerinden, oturuma yalnızca oturum üzerinden ulaşır.
type Kontrolcu struct {
	depo   *depo.Depo
	oturum *oturum.Depo
}

func Yeni(d *depo.Depo, o *oturum.Depo) *Kontrolcu {
	return &Kontrolcu{depo: d, oturum: o}
}
