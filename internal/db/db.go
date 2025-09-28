package db

import (
	"datacenter-calc/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dsn := "host=127.0.0.1 user=postgres password=secret dbname=rip port=5481 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.Order{},
		&model.OrderDevice{},
	)

	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		db.Create(&model.User{
			Username:    "admin",
			Password:    "admin",
			IsModerator: true,
		})
	}

	return db, nil
}
