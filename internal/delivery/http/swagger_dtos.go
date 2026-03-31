package http

import (
	"datacenter-calc/internal/model"
	"time"
)

// RegisterRequest represents user registration input
type RegisterRequest struct {
	Username string `json:"username" example:"newuser"`
	Password string `json:"password" example:"secure123"`
}

// LoginRequest represents login input
type LoginRequest struct {
	Username string `json:"username" example:"user"`
	Password string `json:"password" example:"user"`
}

// UpdateUserRequest represents user update input
type UpdateUserRequest struct {
	Username *string `json:"username" example:"newUserName"`
}

// CreateDeviceRequest represents device creation input
type CreateDeviceRequest struct {
	Name        string `json:"name" example:"Новый сервер"`
	PowerWatt   int    `json:"power_watt" example:"1000"`
	Description string `json:"description" example:"Описание"`
	Category    string `json:"category" example:"Сервер"`
}

// UpdateDeviceRequest represents device update input
type UpdateDeviceRequest struct {
	Name        *string `json:"name,omitempty" example:"Обновлённый сервер"`
	PowerWatt   *int    `json:"power_watt,omitempty" example:"1100"`
	Description *string `json:"description,omitempty" example:"Новое описание"`
	Category    *string `json:"category,omitempty" example:"СХД"`
}

// UpdateCalculationRequest represents calculation update input
type UpdateCalculationRequest struct {
	DCName        *string  `json:"dc_name,omitempty" example:"Мой ЦОД"`
	DCDescription *string  `json:"dc_description,omitempty" example:"Описание ЦОД"`
	PUE           *float64 `json:"pue,omitempty" example:"1.55"`
}

// AddDeviceRequest represents adding device to calculation
type AddDeviceRequest struct {
	DeviceID uint `json:"device_id" example:"1"`
	Quantity int  `json:"quantity" example:"2"`
}

// UpdateDeviceQuantityRequest represents quantity update
type UpdateDeviceQuantityRequest struct {
	Quantity int `json:"quantity" example:"3"`
}

// SuccessMessage represents generic success
type SuccessMessage struct {
	Message string `json:"message" example:"updated"`
}

// ErrorResponse represents error
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}

// UserResponse represents user profile
type UserResponse struct {
	UserID   uint   `json:"user_id" example:"1"`
	Username string `json:"username" example:"user"`
	Role     string `json:"role" example:"user"`
}

// CartResponse represents cart info
type CartResponse struct {
	CalculationID uint `json:"calculation_id" example:"5"`
	TotalItems    int  `json:"total_items" example:"3"`
}

// ImageUploadResponse represents image upload result
type ImageUploadResponse struct {
	ImageURL string `json:"image_url" example:"http://localhost:8050/devices/..."`
}

// LoginResponse represents login output
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJ..."`
	TokenType   string `json:"token_type" example:"Bearer"`
}

// models
type PowerCalculationResponse struct {
	ID          uint                       `json:"id"`
	Status      model.CalculationStatus    `json:"status"`
	CreatedAt   time.Time                  `json:"created_at"`
	FormedAt    *time.Time                 `json:"formed_at,omitempty"`
	CompletedAt *time.Time                 `json:"completed_at,omitempty"`
	TotalPower  *int                       `json:"total_power,omitempty"`
	Creator     string                     `json:"creator"`
	Moderator   *string                    `json:"moderator,omitempty"`
	Devices     []model.DeviceWithQuantity `json:"devices"`
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

type DeviceResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	PowerWatt   int    `json:"power_watt"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
	Category    string `json:"category"`
}
