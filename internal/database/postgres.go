package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"laci-v3/be/internal/config"
)

var DB *gorm.DB

func ConnectPostgres() *gorm.DB {
	cfg := config.Get()

	var dsn string
	if cfg.DBPassword != "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=Asia/Jakarta",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	} else {
		dsn = fmt.Sprintf(
			"postgres://%s@%s:%s/%s?sslmode=disable&TimeZone=Asia/Jakarta",
			cfg.DBUser,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	}

	log.Printf("[DATABASE] Connecting with DSN: host=%s user=%s dbname=%s port=%s", cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}

	log.Printf("[DATABASE] Connected to PostgreSQL: %s (host: %s, port: %s)", cfg.DBName, cfg.DBHost, cfg.DBPort)

	DB = db
	return db
}
