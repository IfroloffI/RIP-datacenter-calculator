package dto

// swagger:model
type AsyncResultRequest struct {
	CalculationID uint `json:"calculation_id" example:"42"`
	TotalPower    int  `json:"total_power" example:"12500"`
	Success       bool `json:"success" example:"true"`
}

/*
Example (success):
{
  "calculation_id": 42,
  "total_power": 12500,
  "success": true
}

Example (failure):
{
  "calculation_id": 42,
  "total_power": 0,
  "success": false
}
*/
