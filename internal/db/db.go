package db

import (
	"datacenter-calc/config"
	"datacenter-calc/internal/model"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.Port,
		cfg.DB.SSLMode,
	)
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

	// Фикс фичи с FK:

	db.Exec(`
    ALTER TABLE orders 
    ADD CONSTRAINT fk_orders_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT;
`)

	db.Exec(`
    ALTER TABLE orders 
    ADD CONSTRAINT fk_orders_moderator 
    FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE SET NULL;
`)

	db.Exec(`
    ALTER TABLE order_devices 
    ADD CONSTRAINT fk_order_devices_order 
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT;
`)

	db.Exec(`
    ALTER TABLE order_devices 
    ADD CONSTRAINT fk_order_devices_device 
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE RESTRICT;
`)

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
