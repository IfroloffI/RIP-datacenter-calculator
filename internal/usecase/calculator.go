package usecase

import (
	"bytes"
	"datacenter-calc/internal/dto"
	"datacenter-calc/internal/minio"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"
	"errors"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

var zero = 0

const maxRetries = 3
const retryDelay = 500 * time.Millisecond

type PowerCalculator struct {
	DeviceRepo          *repo.DeviceRepository
	CalculationRepo     *repo.CalculationRepository
	MinIOClient         *minio.MinIOClient
	CalcPowerServiceURL string
}

func NewPowerCalculator(deviceRepo *repo.DeviceRepository, calcRepo *repo.CalculationRepository, minioClient *minio.MinIOClient, CPSURL string) *PowerCalculator {
	return &PowerCalculator{
		DeviceRepo:          deviceRepo,
		CalculationRepo:     calcRepo,
		MinIOClient:         minioClient,
		CalcPowerServiceURL: CPSURL,
	}
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

func (c *PowerCalculator) CompleteCalculation(calcID, moderatorID uint) error {
	calc, err := c.CalculationRepo.GetCalculationWithDevices(calcID)
	if err != nil {
		return err
	}

	if calc.Status != model.StatusFormed {
		return errors.New("можно завершать только сформированные заявки")
	}

	if len(calc.Devices) == 0 {
		calc.TotalPower = &zero
		err = c.CalculationRepo.UpdateCalculation(calc)
		if err != nil {
			return errors.New("нельзя завершать некорректно заполненные заявки")
		}
		return errors.New("нельзя запускать расчёт для заявки без устройств")
	}

	payload := c.buildAsyncPayload(calc)

	if err := c.sendAsyncCalculationWithRetry(payload); err != nil {
		return err
	}

	now := time.Now()
	calc.ModeratorID = &moderatorID
	calc.CompletedAt = &now
	calc.Status = model.StatusProcessing

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
	calc.TotalPower = nil

	return c.CalculationRepo.UpdateCalculation(calc)
}

func (c *PowerCalculator) ProcessAsyncResult(req dto.AsyncResultRequest) error {
	calc, err := c.CalculationRepo.GetCalculationWithDevices(req.CalculationID)
	if err != nil {
		return err
	}
	if calc.Status != model.StatusProcessing {
		return errors.New("ожидается статус processing")
	}

	if req.Success {
		calc.TotalPower = &req.TotalPower
		calc.Status = model.StatusCompleted
	} else {
		calc.TotalPower = nil
		calc.Status = model.StatusFailed
	}

	return c.CalculationRepo.UpdateCalculation(calc)
}

func (c *PowerCalculator) buildAsyncPayload(calc *model.PowerCalculation) dto.AsyncCalculationRequest {
	payload := dto.AsyncCalculationRequest{
		CalculationID: calc.ID,
		PUE:           calc.PUE,
		Devices:       make([]dto.AsyncDevicePayload, 0, len(calc.Devices)),
	}

	for _, d := range calc.Devices {
		qty := calc.DeviceQuantities[d.ID]
		payload.Devices = append(payload.Devices, dto.AsyncDevicePayload{
			ID:        d.ID,
			Quantity:  qty,
			PowerWatt: d.PowerWatt,
		})
	}

	return payload
}

func (c *PowerCalculator) sendAsyncCalculationWithRetry(payload dto.AsyncCalculationRequest) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(
			http.MethodPost,
			c.CalcPowerServiceURL,
			bytes.NewReader(body),
		)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return errors.New("async service unreachable after retries: " + err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return errors.New("async service returned client error: " + resp.Status)
		}

		lastErr = errors.New("async service returned server error: " + resp.Status)
		if attempt < maxRetries {
			time.Sleep(retryDelay)
			continue
		}
	}

	return errors.New("async service failed after " + string(rune(maxRetries)) + " retries: " + lastErr.Error())
}
