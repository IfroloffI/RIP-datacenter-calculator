package model

type Device struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PowerWatt   int    `json:"power_watt"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"` // сервер, СХД, ...
}

type Order struct {
	ID        int    `json:"id"`
	DeviceIDs []int  `json:"device_ids"`
	CreatedAt string `json:"created_at"`
}
