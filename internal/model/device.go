package model

type Device struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PowerWatt   int    `json:"power_watt"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"` // сервер, СХД, ...
}
