// Package veritabani, MySQL kaynak (yazma) ve replika (okuma) bağlantılarını
// ve Redis istemcisini kurar. Bağlantı bilgileri docker-compose.yml'deki
// ortam değişkenlerinden okunur.
package veritabani

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Baglantilar uygulamanın kullandığı üç bağlantıyı bir arada tutar.
//   Yazma: mysql-primary  → INSERT/UPDATE ve stored procedure çağrıları
//   Okuma: mysql-replica  → listeleme ve raporlar (salt okunur)
type Baglantilar struct {
	Yazma *gorm.DB
	Okuma *gorm.DB
	Redis *redis.Client
}

func ortam(ad, varsayilan string) string {
	if d := os.Getenv(ad); d != "" {
		return d
	}
	return varsayilan
}

// mysqlAc verilen sunucuya bağlanır. Container'lar aynı anda kalktığı için
// veritabanı henüz hazır değilse birkaç kez yeniden dener.
func mysqlAc(host string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		ortam("DB_KULLANICI", "yemekhane_app"), ortam("DB_SIFRE", "Uyg_2026"),
		host, ortam("DB_PORT", "3306"), ortam("DB_ADI", "yemekhane"))

	var db *gorm.DB
	var err error
	for deneme := 1; deneme <= 10; deneme++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
		if err == nil {
			sqlDB, _ := db.DB()
			if err = sqlDB.Ping(); err == nil {
				sqlDB.SetMaxOpenConns(10)
				sqlDB.SetConnMaxLifetime(5 * time.Minute)
				return db, nil
			}
		}
		log.Printf("%s bağlantısı kurulamadı (%d/10): %v", host, deneme, err)
		time.Sleep(3 * time.Second)
	}
	return nil, err
}

// Ac üç bağlantıyı da kurar; biri kurulamazsa hata döner.
func Ac() (*Baglantilar, error) {
	yazma, err := mysqlAc(ortam("DB_YAZMA_HOST", "mysql-primary"))
	if err != nil {
		return nil, fmt.Errorf("kaynak veritabanı: %w", err)
	}
	okuma, err := mysqlAc(ortam("DB_OKUMA_HOST", "mysql-replica"))
	if err != nil {
		return nil, fmt.Errorf("replika veritabanı: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: ortam("REDIS_HOST", "redis") + ":" + ortam("REDIS_PORT", "6379"),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}

	return &Baglantilar{Yazma: yazma, Okuma: okuma, Redis: rdb}, nil
}
