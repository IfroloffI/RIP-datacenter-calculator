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
	userID := uint(1)
	order := c.OrderRepo.GetDraftByUser(userID)
	if order == nil {
		order = c.OrderRepo.CreateDraft(userID)
	}
	return *order
}

func (c *PowerCalculator) CalculateTotalPower(order model.Order) (base int, calculated int) {
	var total int
	for _, dev := range order.Devices {
		total += dev.PowerWatt
	}
	delta := int(float64(total) * 0.5)
	return total, total + delta
}

func (c *PowerCalculator) GetDevicesInOrder(order model.Order) []model.Device {
	return order.Devices
}
