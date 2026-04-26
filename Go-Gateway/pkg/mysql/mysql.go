package mysql

import (
	"log"

	"github.com/nym/go-gateway/internal/config"
	"github.com/nym/go-gateway/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(mysql.Open(config.AppConfig.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	
	// AutoMigrate all models
	err = DB.AutoMigrate(
		&models.Admin{},
		&models.RouteRule{},
		&models.BlacklistRule{},
		&models.BlockLog{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	log.Println("MySQL connection established and migration completed.")
}
