package usecase

import (
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"
)

type PowerCalculator struct {
	DeviceRepo *repo.DeviceRepository
	OrderRepo  *repo.OrderRepository
}

func NewPowerCalculator(deviceRepo *repo.DeviceRepository, orderRepo *repo.OrderRepository) *PowerCalculator {
	return &PowerCalculator{
		DeviceRepo: deviceRepo,
		OrderRepo:  orderRepo,
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

func (c *PowerCalculator) GetDevicesInOrder(order model.Order) []model.DeviceWithQuantityAndIPW {
	var result []model.DeviceWithQuantityAndIPW

	var orderDevices []model.OrderDevice
	c.DeviceRepo.DB.Where("order_id = ?", order.ID).Find(&orderDevices)

	quantityMap := make(map[uint]int)
	for _, od := range orderDevices {
		quantityMap[od.DeviceID] = od.Quantity
	}

	for _, dev := range order.Devices {
		result = append(result, model.DeviceWithQuantityAndIPW{
			Device:         dev,
			Count:          quantityMap[dev.ID],
			InterPowerWatt: quantityMap[dev.ID] * dev.PowerWatt,
		})
	}

	return result
}
