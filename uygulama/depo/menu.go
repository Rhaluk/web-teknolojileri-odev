package depo

import (
	"context"
	"log"
	"math"
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
// Kullanıcının profilindeki alerjenleri içeren yemeklere uyarı eklenir.
func (d *Depo) GununMenusu(yemekhaneID, ogunID uint, tarih string, kullaniciID uint, rol string) (*modeller.GununMenusu, error) {
	ctx := context.Background()
	anahtar := onbellek.MenuAnahtari(yemekhaneID, tarih, ogunID)

	var ic modeller.MenuIcerigi
	// Sürümü eski olan kayıt (Aşama 4 öncesi yapı) yokmuş gibi sayılır ve üzerine yazılır
	if d.onbellek.Getir(ctx, anahtar, &ic) && ic.Surum == onbellek.MenuSurumu {
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
	}
	if rol == "ogrenci" {
		menu.Fiyat = ic.OgrenciFiyati
	}
	if ic.YarimOrani != nil {
		// sp_rezervasyon_yap ile aynı hesap: ROUND(fiyat × oran, 2)
		y := math.Round(menu.Fiyat**ic.YarimOrani*100) / 100
		menu.YarimFiyat = &y
	}

	// Alerjen uyarısı: kullanıcıya özel olduğu için önbellekten sonra eklenir
	kulAlerjen, err := d.KullaniciAlerjenIDleri(kullaniciID)
	if err != nil {
		return nil, err
	}
	secili := make(map[uint]bool, len(kulAlerjen))
	for _, id := range kulAlerjen {
		secili[id] = true
	}
	menu.Yemekler = make([]modeller.MenuYemegi, len(ic.Yemekler))
	for i, y := range ic.Yemekler {
		y.UyariAlerjenler = []string{}
		for j, aid := range y.AlerjenIDler {
			if secili[aid] {
				y.UyariAlerjenler = append(y.UyariAlerjenler, y.Alerjenler[j])
			}
		}
		if len(y.UyariAlerjenler) > 0 {
			menu.AlerjenUyarisi = true
		}
		menu.Yemekler[i] = y
	}

	// Boş yer v_menu_doluluk'tan; her rezervasyonda değiştiği için önbelleğe alınmaz
	var doluluk struct{ Kapasite, BosYer int }
	err = d.okuma.Table("v_menu_doluluk").Select("kapasite, bos_yer").
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

	// Öğünün yarım porsiyon oranı (Aşama 4); NULL ise yarım porsiyon yok
	var ogun struct{ YarimPorsiyonOrani *float64 }
	err = d.okuma.Table("ogunler").Select("yarim_porsiyon_orani").Where("id = ?", ogunID).Scan(&ogun).Error
	if err != nil {
		return nil, cevir(err)
	}

	// Alerjen adı → id (view adları ada göre sıralı verir)
	tum, err := d.Alerjenler()
	if err != nil {
		return nil, err
	}
	alerjenID := make(map[string]uint, len(tum))
	for _, a := range tum {
		alerjenID[a.Ad] = a.ID
	}

	ilk := satirlar[0]
	ic := &modeller.MenuIcerigi{
		Surum:          onbellek.MenuSurumu,
		YarimOrani:     ogun.YarimPorsiyonOrani,
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
		y := modeller.MenuYemegi{ID: s.YemekID, Ad: s.Yemek, Kategori: s.Kategori, Kalori: s.Kalori, Vejetaryen: s.Vejetaryen,
			Alerjenler: []string{}, AlerjenIDler: []uint{}}
		if s.Alerjenler != nil {
			y.Alerjenler = strings.Split(*s.Alerjenler, ", ")
		}
		// id'ler adlarla aynı sırada: Alerjenler[i] ↔ AlerjenIDler[i]
		for _, ad := range y.Alerjenler {
			y.AlerjenIDler = append(y.AlerjenIDler, alerjenID[ad])
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
		       ROUND(`+fiyatSecimi+` * o.yarim_porsiyon_orani, 2) AS yarim_fiyat,
		       EXISTS (SELECT 1 FROM menu_yemekleri my
		                 JOIN yemek_alerjenleri ya    ON ya.yemek_id = my.yemek_id
		                 JOIN kullanici_alerjenleri ka ON ka.alerjen_id = ya.alerjen_id AND ka.kullanici_id = ?
		                WHERE my.menu_id = d.menu_id) AS alerjen_uyarisi,
		       NOW() < d.rezervasyon_son AND d.bos_yer > 0 AS rezerve_edilebilir,
		       (SELECT SUM(v.kalori) FROM v_menu_detay v WHERE v.menu_id = d.menu_id) AS toplam_kalori,
		       (SELECT GROUP_CONCAT(v.yemek ORDER BY v.kategori_sira, v.yemek SEPARATOR '|')
		          FROM v_menu_detay v WHERE v.menu_id = d.menu_id) AS yemek_listesi,
		       r.durum AS rezervasyon_durumu, r.porsiyon
		  FROM v_menu_doluluk d
		  JOIN menuler m ON m.id = d.menu_id
		  JOIN ogunler o ON o.id = m.ogun_id
		  LEFT JOIN rezervasyonlar r ON r.menu_id = d.menu_id AND r.kullanici_id = ?
		 WHERE d.yemekhane_id = ?
		   AND d.tarih BETWEEN ? AND DATE_ADD(?, INTERVAL 6 DAY)
		 ORDER BY d.tarih, d.ogun_id`,
		rol, rol, kullaniciID, kullaniciID, yemekhaneID, baslangic, baslangic).Scan(&liste).Error
	if err != nil {
		return nil, cevir(err)
	}
	for i := range liste {
		liste[i].TarihMetin = liste[i].Tarih.Format(time.DateOnly)
		liste[i].Yemekler = strings.Split(liste[i].YemekListesi, "|")
	}
	return liste, nil
}
