package usecase

import (
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"
)

type PowerCalculator struct {
	Repo *repo.DeviceRepository
}

func NewPowerCalculator(repo *repo.DeviceRepository) *PowerCalculator {
	return &PowerCalculator{Repo: repo}
}

func (c *PowerCalculator) CalculateTotalPower(order model.Order) (base int, calculated int) {
	var total int
	for _, id := range order.DeviceIDs {
		if dev := c.Repo.GetByID(id); dev != nil {
			total += dev.PowerWatt
		}
	}
	delta := int(float64(total) * 0.5)
	return total, total + delta
}

func (c *PowerCalculator) GetDevicesInOrder(order model.Order) []model.Device {
	var devices []model.Device
	for _, id := range order.DeviceIDs {
		if dev := c.Repo.GetByID(id); dev != nil {
			devices = append(devices, *dev)
		}
	}
	return devices
}
