package repo

import (
	"datacenter-calc/internal/model"

	"gorm.io/gorm"
)

type DeviceRepository struct {
	DB *gorm.DB
}

func (r *DeviceRepository) GetAllActive() []model.Device {
	var devices []model.Device
	r.DB.Where("is_deleted = ?", false).Find(&devices)
	return devices
}

func (r *DeviceRepository) GetByID(id uint) *model.Device {
	var device model.Device
	if r.DB.Where("id = ? AND is_deleted = ?", id, false).First(&device).Error != nil {
		return nil
	}
	return &device
}
