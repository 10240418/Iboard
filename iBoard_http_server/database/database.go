package databases

import (
	"fmt"
	"log"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB_CONN *gorm.DB

func InitDB(host, user, password, port, dbname string) *gorm.DB {
	// 先连接MySQL（不指定数据库）
	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	db, err := gorm.Open(mysql.Open(rootDsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to MySQL: %v\n", err)
		return nil
	}

	// 创建数据库
	createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbname)
	if err := db.Exec(createDB).Error; err != nil {
		log.Printf("Failed to create database: %v\n", err)
		return nil
	}

	// 连接指定的数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	DB_CONN, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		return nil
	}

	// 添加自动迁移
	if err := DB_CONN.AutoMigrate(
		&base_models.SuperAdmin{},
		&base_models.BuildingAdmin{},
		&base_models.Building{},
		&base_models.Advertisement{},
		&base_models.Notice{},
		&base_models.File{},
		&base_models.Device{},
	); err != nil {
		log.Printf("Failed to auto migrate: %v\n", err)
		return nil
	}

	log.Println("Database migration completed successfully")
	return DB_CONN
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	// 在这里添加所有需要迁移的模型
	err := db.AutoMigrate(
		&base_models.SuperAdmin{},
		&base_models.BuildingAdmin{},
		&base_models.Building{},
		&base_models.Advertisement{},
		&base_models.Notice{},
		&base_models.File{},
		&base_models.Device{},
	)

	if err != nil {
		return fmt.Errorf("auto migration failed: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}
