package model

type Device struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	PowerWatt   int    `gorm:"not null"`
	Description string
	ImageURL    string `gorm:"default:null"`
	Category    string `gorm:"not null"`
	IsDeleted   bool   `gorm:"not null;default:false"`
}
