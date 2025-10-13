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
type DeviceResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	PowerWatt   int    `json:"power_watt"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
	Category    string `json:"category"`
}

func (d *Device) ToResponse(minioURL string) DeviceResponse {
	img := ""
	if d.ImageURL != "" {
		img = minioURL + "/" + d.ImageURL
	}
	return DeviceResponse{
		ID:          d.ID,
		Name:        d.Name,
		PowerWatt:   d.PowerWatt,
		Description: d.Description,
		ImageURL:    img,
		Category:    d.Category,
	}
}
