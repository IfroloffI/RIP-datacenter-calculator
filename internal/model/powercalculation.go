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

type PowerCalculation struct {
	ID          uint              `gorm:"primaryKey"`
	Status      CalculationStatus `gorm:"not null;type:varchar(20)"`
	CreatedAt   time.Time         `gorm:"not null"`
	CreatedBy   uint              `gorm:"not null"`
	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint
	TotalPower  *int // вычисляется при завершении

	DCName        string  `gorm:"type:varchar(255);default:null"`
	DCDescription string  `gorm:"type:text;default:null"`
	PUE           float64 `gorm:"default:1.5"`

	User             User         `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Moderator        *User        `gorm:"foreignKey:ModeratorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Devices          []Device     `gorm:"many2many:power_calculation_devices;joinForeignKey:calculation_id;joinReferences:device_id"`
	DeviceQuantities map[uint]int `gorm:"-" json:"-"`
}

type PowerCalculationResponse struct {
	ID          uint              `json:"id"`
	Status      CalculationStatus `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	FormedAt    *time.Time        `json:"formed_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	TotalPower  *int              `json:"total_power,omitempty"`
	Creator     string            `json:"creator"`
	Moderator   *string           `json:"moderator,omitempty"`
	Devices     []DeviceResponse  `json:"devices"`
}

func (p *PowerCalculation) ToResponse(minioURL string) PowerCalculationResponse {
	devices := make([]DeviceResponse, len(p.Devices))
	for i, d := range p.Devices {
		devices[i] = d.ToResponse(minioURL)
	}

	creator := p.User.Username
	var moderator *string
	if p.Moderator != nil {
		m := p.Moderator.Username
		moderator = &m
	}

	return PowerCalculationResponse{
		ID:          p.ID,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		FormedAt:    p.FormedAt,
		CompletedAt: p.CompletedAt,
		TotalPower:  p.TotalPower,
		Creator:     creator,
		Moderator:   moderator,
		Devices:     devices,
	}
}

type DeviceWithQuantity struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	PowerWatt   int    `json:"power_watt"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
	Category    string `json:"category"`
	Quantity    int    `json:"quantity"`
}
