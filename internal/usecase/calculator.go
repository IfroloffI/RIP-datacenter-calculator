package usecase

import (
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"
)

type PowerCalculator struct {
	DeviceRepo      *repo.DeviceRepository
	CalculationRepo *repo.CalculationRepository
}

func NewPowerCalculator(deviceRepo *repo.DeviceRepository, calcRepo *repo.CalculationRepository) *PowerCalculator {
	return &PowerCalculator{
		DeviceRepo:      deviceRepo,
		CalculationRepo: calcRepo,
	}
}

func (c *PowerCalculator) CalculateTotalPower(devices []model.DeviceWithQuantityAndIPW, pue float64) (base int, calculated int) {
	var total int
	for _, d := range devices {
		total += d.PowerWatt * d.Count
	}
	base = total
	calculated = int(float64(total) * pue)
	return base, calculated
}

func (c *PowerCalculator) GetDevicesInCalculation(calc model.Calculation) []model.DeviceWithQuantityAndIPW {
	var result []model.DeviceWithQuantityAndIPW

	var calcDevices []model.CalculationDevice
	c.DeviceRepo.DB.Where("calculation_id = ?", calc.ID).Find(&calcDevices)

	quantityMap := make(map[uint]int)
	for _, cd := range calcDevices {
		quantityMap[cd.DeviceID] = cd.Quantity
	}

	for _, dev := range calc.Devices {
		result = append(result, model.DeviceWithQuantityAndIPW{
			Device:         dev,
			Count:          quantityMap[dev.ID],
			InterPowerWatt: quantityMap[dev.ID] * dev.PowerWatt,
		})
	}

	return result
}
