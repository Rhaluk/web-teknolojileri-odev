package depo

import (
	"context"
	"log"
	"strings"
	"time"

	"yemekhane/modeller"
	"yemekhane/onbellek"
)

// fiyatSecimi kullanıcının rolüne göre fiyat sütununu seçer.
// Yönetici de personel fiyatını öder (stored procedure ile aynı kural).
const fiyatSecimi = "IF(? = 'ogrenci', m.ogrenci_fiyati, m.personel_fiyati)"

// Yemekhaneler bütün yemekhaneleri döner (replika).
func (d *Depo) Yemekhaneler() ([]modeller.Yemekhane, error) {
	var liste []modeller.Yemekhane
	err := d.okuma.Order("id").Find(&liste).Error
	return liste, cevir(err)
}

// GununMenusu bir yemekhanenin verilen gün ve öğündeki menüsünü getirir.
// Menü içeriği önce Redis'e sorulur, yoksa replikadan okunup Redis'e yazılır.
// Fiyat kullanıcının rolüne göre seçilir; boş yer her istekte replikadan okunur.
func (d *Depo) GununMenusu(yemekhaneID, ogunID uint, tarih, rol string) (*modeller.GununMenusu, error) {
	ctx := context.Background()
	anahtar := onbellek.MenuAnahtari(yemekhaneID, tarih, ogunID)

	var ic modeller.MenuIcerigi
	if d.onbellek.Getir(ctx, anahtar, &ic) {
		log.Printf("menü %s → Redis'ten okundu (önbellekte var)", anahtar)
	} else {
		bulunan, err := d.menuIcerigi(yemekhaneID, ogunID, tarih)
		if err != nil {
			return nil, err
		}
		ic = *bulunan
		sure := d.onbellek.Yaz(ctx, anahtar, ic)
		log.Printf("menü %s → veritabanından okundu, Redis'e yazıldı (gün sonuna %s)", anahtar, sure.Round(time.Second))
	}

	menu := &modeller.GununMenusu{
		MenuID:         ic.MenuID,
		Tarih:          ic.Tarih,
		Yemekhane:      ic.Yemekhane,
		Ogun:           ic.Ogun,
		RezervasyonSon: ic.RezervasyonSon,
		Fiyat:          ic.PersonelFiyati,
		ToplamKalori:   ic.ToplamKalori,
		Yemekler:       ic.Yemekler,
	}
	if rol == "ogrenci" {
		menu.Fiyat = ic.OgrenciFiyati
	}

	// Boş yer v_menu_doluluk'tan; her rezervasyonda değiştiği için önbelleğe alınmaz
	var doluluk struct{ Kapasite, BosYer int }
	err := d.okuma.Table("v_menu_doluluk").Select("kapasite, bos_yer").
		Where("menu_id = ?", menu.MenuID).Scan(&doluluk).Error
	menu.Kapasite, menu.BosYer = doluluk.Kapasite, doluluk.BosYer
	return menu, cevir(err)
}

// menuIcerigi menünün yemeklerini v_menu_detay'dan kategori sırasıyla okur
// ve toplam kaloriyi hesaplar (replika).
func (d *Depo) menuIcerigi(yemekhaneID, ogunID uint, tarih string) (*modeller.MenuIcerigi, error) {
	var satirlar []modeller.MenuDetaySatiri
	err := d.okuma.Table("v_menu_detay").
		Where("yemekhane_id = ? AND tarih = ? AND ogun_id = ?", yemekhaneID, tarih, ogunID).
		Order("kategori_sira, yemek").
		Find(&satirlar).Error
	if err != nil {
		return nil, cevir(err)
	}
	if len(satirlar) == 0 {
		return nil, ErrBulunamadi
	}

	ilk := satirlar[0]
	ic := &modeller.MenuIcerigi{
		MenuID:         ilk.MenuID,
		Tarih:          ilk.Tarih.Format(time.DateOnly),
		Yemekhane:      ilk.Yemekhane,
		Ogun:           ilk.Ogun,
		RezervasyonSon: ilk.RezervasyonSon,
		OgrenciFiyati:  ilk.OgrenciFiyati,
		PersonelFiyati: ilk.PersonelFiyati,
	}
	for _, s := range satirlar {
		ic.ToplamKalori += s.Kalori
		y := modeller.MenuYemegi{ID: s.YemekID, Ad: s.Yemek, Kategori: s.Kategori, Kalori: s.Kalori, Vejetaryen: s.Vejetaryen, Alerjenler: []string{}}
		if s.Alerjenler != nil {
			y.Alerjenler = strings.Split(*s.Alerjenler, ", ")
		}
		ic.Yemekler = append(ic.Yemekler, y)
	}
	return ic, nil
}

// HaftalikMenu bir yemekhanenin baslangic gününden itibaren 7 günlük menülerini
// getirir (replika). Her satırda kullanıcının rolüne göre fiyat, boş yer,
// rezervasyon yapılıp yapılamayacağı ve kullanıcının o menüdeki rezervasyonu var.
// Son saat karşılaştırması veritabanının saatiyle (NOW()) yapılır.
func (d *Depo) HaftalikMenu(yemekhaneID uint, baslangic string, kullaniciID uint, rol string) ([]modeller.HaftalikMenu, error) {
	var liste []modeller.HaftalikMenu
	err := d.okuma.Raw(`
		SELECT d.menu_id, d.tarih, d.ogun_id, d.ogun, d.bos_yer, d.rezervasyon_son,
		       `+fiyatSecimi+` AS fiyat,
		       NOW() < d.rezervasyon_son AND d.bos_yer > 0 AS rezerve_edilebilir,
		       (SELECT SUM(v.kalori) FROM v_menu_detay v WHERE v.menu_id = d.menu_id) AS toplam_kalori,
		       (SELECT GROUP_CONCAT(v.yemek ORDER BY v.kategori_sira, v.yemek SEPARATOR '|')
		          FROM v_menu_detay v WHERE v.menu_id = d.menu_id) AS yemek_listesi,
		       r.durum AS rezervasyon_durumu
		  FROM v_menu_doluluk d
		  JOIN menuler m ON m.id = d.menu_id
		  LEFT JOIN rezervasyonlar r ON r.menu_id = d.menu_id AND r.kullanici_id = ?
		 WHERE d.yemekhane_id = ?
		   AND d.tarih BETWEEN ? AND DATE_ADD(?, INTERVAL 6 DAY)
		 ORDER BY d.tarih, d.ogun_id`,
		rol, kullaniciID, yemekhaneID, baslangic, baslangic).Scan(&liste).Error
	if err != nil {
		return nil, cevir(err)
	}
	for i := range liste {
		liste[i].TarihMetin = liste[i].Tarih.Format(time.DateOnly)
		liste[i].Yemekler = strings.Split(liste[i].YemekListesi, "|")
	}
	return liste, nil
}
