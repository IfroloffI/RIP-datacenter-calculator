package dto

type AsyncDevicePayload struct {
	ID        uint `json:"id"`
	Quantity  int  `json:"quantity"`
	PowerWatt int  `json:"power_watt"`
}

type AsyncCalculationRequest struct {
	CalculationID uint                 `json:"calculation_id"`
	Devices       []AsyncDevicePayload `json:"devices"`
}

/*
Example of Request 1 (many devices):
{
  "calculation_id": 42,
  "devices": [
    {
      "id": 1,
      "quantity": 10,
      "power_watt": 500
    },
    {
      "id": 2,
      "quantity": 3,
      "power_watt": 750
    }
  ]
}

Example of Request 2 (single device):
{
  "calculation_id": 7,
  "devices": [
    {
      "id": 5,
      "quantity": 2,
      "power_watt": 1200
    }
  ]
}
*/
