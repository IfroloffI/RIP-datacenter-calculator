package model

import "time"

type CalculationStatus string

const (
	StatusDraft     CalculationStatus = "draft"
	StatusDeleted   CalculationStatus = "deleted"
	StatusFormed    CalculationStatus = "formed"
	StatusCompleted CalculationStatus = "completed"
	StatusRejected  CalculationStatus = "rejected"
)

type Calculation struct {
	ID          uint              `gorm:"primaryKey"`
	Status      CalculationStatus `gorm:"not null;type:varchar(20)"`
	CreatedAt   time.Time         `gorm:"not null"`
	CreatedBy   uint              `gorm:"not null"`
	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint
	TotalPower  *int

	User               User                `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Moderator          *User               `gorm:"foreignKey:ModeratorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Devices            []Device            `gorm:"many2many:calculation_devices;"`
	CalculationDevices []CalculationDevice `gorm:"foreignKey:CalculationID"`
}
