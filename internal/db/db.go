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

	// Миграции
	db.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.PowerCalculation{},
		&model.PowerCalculationDevice{},
	)

	// Удаляем старые FK (если есть)
	db.Exec("ALTER TABLE power_calculations DROP CONSTRAINT IF EXISTS fk_power_calculations_created_by;")
	db.Exec("ALTER TABLE power_calculations DROP CONSTRAINT IF EXISTS fk_power_calculations_moderator;")
	db.Exec("ALTER TABLE power_calculation_devices DROP CONSTRAINT IF EXISTS fk_power_calculation_devices_power_calculation;")
	db.Exec("ALTER TABLE power_calculation_devices DROP CONSTRAINT IF EXISTS fk_power_calculation_devices_device;")

	// Новые FK без каскадного удаления
	db.Exec(`
		ALTER TABLE power_calculations 
		ADD CONSTRAINT fk_power_calculations_created_by 
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT;
	`)

	db.Exec(`
		ALTER TABLE power_calculations 
		ADD CONSTRAINT fk_power_calculations_moderator 
		FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE SET NULL;
	`)

	db.Exec(`
		ALTER TABLE power_calculation_devices 
		ADD CONSTRAINT fk_power_calculation_devices_power_calculation
		FOREIGN KEY (calculation_id) REFERENCES power_calculations(id) ON DELETE RESTRICT;
	`)

	db.Exec(`
		ALTER TABLE power_calculation_devices 
		ADD CONSTRAINT fk_power_calculation_devices_device 
		FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE RESTRICT;
	`)

	// Инициализация данных
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	if userCount == 0 {
		db.Create(&model.User{Username: "user", Password: "user", IsModerator: false})
		db.Create(&model.User{Username: "admin", Password: "admin", IsModerator: true})
	}

	var deviceCount int64
	db.Model(&model.Device{}).Count(&deviceCount)
	if deviceCount == 0 {
		devices := []model.Device{
			{Name: "Сервер Dell R760", PowerWatt: 850, Description: "Мощный сервер...", ImageURL: "dell-r760.png", Category: "Сервер", IsDeleted: false},
			{Name: "Коммутатор Dell N2024", PowerWatt: 30, Description: "Коммутатор...", ImageURL: "dell-n2024.png", Category: "Коммутатор", IsDeleted: false},
			{Name: "СХД Dell PowerVault ME5024", PowerWatt: 1400, Description: "СХД...", ImageURL: "dell-me5024.png", Category: "СХД", IsDeleted: false},
		}
		db.Create(devices)
	}

	return db, nil
}
