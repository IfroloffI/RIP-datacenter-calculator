package usecase

import (
	"datacenter-calc/internal/minio"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"
	"errors"
	"time"
)

type PowerCalculator struct {
	DeviceRepo      *repo.DeviceRepository
	CalculationRepo *repo.CalculationRepository
	MinIOClient     *minio.MinIOClient
}

func NewPowerCalculator(deviceRepo *repo.DeviceRepository, calcRepo *repo.CalculationRepository, minioClient *minio.MinIOClient) *PowerCalculator {
	return &PowerCalculator{
		DeviceRepo:      deviceRepo,
		CalculationRepo: calcRepo,
		MinIOClient:     minioClient,
	}
}

func (c *PowerCalculator) CalculateTotalPower(devices []model.Device, quantities map[uint]int) int {
	total := 0
	for _, d := range devices {
		q := quantities[d.ID]
		total += d.PowerWatt * q
	}
	return total
}

func (c *PowerCalculator) CompleteCalculation(calcID, moderatorID uint) error {
	calc, err := c.CalculationRepo.GetCalculationWithDevices(calcID)
	if err != nil {
		return err
	}
	if calc.Status != model.StatusFormed {
		return errors.New("можно завершать только сформированные заявки")
	}

	var calcDevices []model.PowerCalculationDevice
	err = c.DeviceRepo.DB.Where("calculation_id = ?", calcID).Find(&calcDevices).Error
	if err != nil {
		return err
	}

	quantities := calc.DeviceQuantities
	total := c.CalculateTotalPower(calc.Devices, quantities)

	now := time.Now()
	calc.ModeratorID = &moderatorID
	calc.CompletedAt = &now
	calc.TotalPower = &total
	calc.Status = model.StatusCompleted

	return c.CalculationRepo.UpdateCalculation(calc)
}

func (c *PowerCalculator) RejectCalculation(calcID, moderatorID uint) error {
	calc, err := c.CalculationRepo.GetCalculationWithDevices(calcID)
	if err != nil {
		return err
	}
	if calc.Status != model.StatusFormed {
		return errors.New("можно отклонять только сформированные заявки")
	}

	now := time.Now()
	calc.ModeratorID = &moderatorID
	calc.CompletedAt = &now
	calc.Status = model.StatusRejected

	return c.CalculationRepo.UpdateCalculation(calc)
}

func (c *PowerCalculator) FormCalculation(calcID uint, requiredFieldsValid bool) error {
	if !requiredFieldsValid {
		return errors.New("обязательные поля не заполнены")
	}
	calc, err := c.CalculationRepo.GetCalculationWithDevices(calcID)
	if err != nil {
		return err
	}
	if calc.Status != model.StatusDraft {
		return errors.New("можно формировать только черновики")
	}
	now := time.Now()
	calc.FormedAt = &now
	calc.Status = model.StatusFormed
	return c.CalculationRepo.UpdateCalculation(calc)
}
