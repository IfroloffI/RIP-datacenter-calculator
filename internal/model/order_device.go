package model

type OrderDevice struct {
	OrderID  uint `gorm:"primaryKey;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	DeviceID uint `gorm:"primaryKey;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Quantity int  `gorm:"not null;default:1"`
}
