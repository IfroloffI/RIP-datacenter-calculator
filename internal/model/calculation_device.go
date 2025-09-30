package model

type CalculationDevice struct {
	CalculationID uint `gorm:"primaryKey;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	DeviceID      uint `gorm:"primaryKey;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Quantity      int  `gorm:"not null;default:1"`
}
