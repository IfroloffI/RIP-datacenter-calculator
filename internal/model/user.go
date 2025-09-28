package model

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Username    string `gorm:"uniqueIndex;not null"`
	Password    string `gorm:"not null"`
	IsModerator bool   `gorm:"not null;default:false"`
}
