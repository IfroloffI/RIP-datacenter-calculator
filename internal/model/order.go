package model

import "time"

type OrderStatus string

const (
	StatusDraft     OrderStatus = "draft"
	StatusDeleted   OrderStatus = "deleted"
	StatusFormed    OrderStatus = "formed"
	StatusCompleted OrderStatus = "completed"
	StatusRejected  OrderStatus = "rejected"
)

type Order struct {
	ID          uint        `gorm:"primaryKey"`
	Status      OrderStatus `gorm:"not null;type:varchar(20)"`
	CreatedAt   time.Time   `gorm:"not null"`
	CreatedBy   uint        `gorm:"not null"`
	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint
	TotalPower  *int

	User         User          `gorm:"foreignKey:CreatedBy"`
	Moderator    *User         `gorm:"foreignKey:ModeratorID"`
	Devices      []Device      `gorm:"many2many:order_devices;"`
	OrderDevices []OrderDevice `gorm:"foreignKey:OrderID"`
}
