package model

type Order struct {
	ID        int    `json:"id"`
	DeviceIDs []int  `json:"device_ids"`
	CreatedAt string `json:"created_at"`
}
