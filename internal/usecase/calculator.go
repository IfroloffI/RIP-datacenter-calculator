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

func (c *PowerCalculator) GetCurrentOrder() model.Order {
	order := c.OrderRepo.GetCurrentOrder()
	if order == nil {
		return model.Order{}
	}
	return *order
}

func (c *PowerCalculator) CalculateTotalPower(order model.Order) (base int, calculated int) {
	var total int
	for _, id := range order.DeviceIDs {
		if dev := c.DeviceRepo.GetByID(id); dev != nil {
			total += dev.PowerWatt
		}
	}
	delta := int(float64(total) * 0.5)
	return total, total + delta
}

func (c *PowerCalculator) GetDevicesInOrder(order model.Order) []model.Device {
	var devices []model.Device
	for _, id := range order.DeviceIDs {
		if dev := c.DeviceRepo.GetByID(id); dev != nil {
			devices = append(devices, *dev)
		}
	}
	return devices
}
