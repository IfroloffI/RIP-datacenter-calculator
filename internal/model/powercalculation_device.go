package model

type PowerCalculationDevice struct {
	CalculationID uint `gorm:"primaryKey;column:calculation_id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	DeviceID      uint `gorm:"primaryKey;column:device_id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Quantity      int  `gorm:"not null;default:1"`
}
