package model

type OrderDevice struct {
	OrderID  uint `gorm:"primaryKey"`
	DeviceID uint `gorm:"primaryKey"`
	Quantity int  `gorm:"not null;default:1"`
}
