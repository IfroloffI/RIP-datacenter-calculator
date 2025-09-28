package db

import (
	"datacenter-calc/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dsn := "host=127.0.0.1 user=postgres password=secret dbname=rip port=8082 sslmode=disable"
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

	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	if userCount == 0 {
		db.Create(&model.User{
			Username:    "user",
			Password:    "user",
			IsModerator: false,
		})
		db.Create(&model.User{
			Username:    "admin",
			Password:    "admin",
			IsModerator: true,
		})
	}

	var deviceCount int64
	db.Model(&model.Device{}).Count(&deviceCount)
	if deviceCount == 0 {
		db.Create([]model.Device{
			{
				Name:        "Сервер Dell R760",
				PowerWatt:   850,
				Description: "Мощный сервер для критичных workloads. 2x Intel Xeon, до 3TB RAM.",
				ImageURL:    "dell-r760.jpg",
				Category:    "Сервер",
				IsDeleted:   false,
			},
			{
				Name:        "Коммутатор Dell N2024",
				PowerWatt:   30,
				Description: "Коммутатор уровня доступа, 24 порта 1G.",
				ImageURL:    "dell-n2024.jpg",
				Category:    "Коммутатор",
				IsDeleted:   false,
			},
			{
				Name:        "СХД Dell PowerVault ME5024",
				PowerWatt:   1400,
				Description: "Система хранения данных начального уровня, 24 дисковых слота.",
				ImageURL:    "dell-me5024.jpg",
				Category:    "СХД",
				IsDeleted:   false,
			},
		})
	}

	return db, nil
}
