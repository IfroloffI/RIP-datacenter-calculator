package repo

import (
	"datacenter-calc/internal/model"
	"time"

	"gorm.io/gorm"
)

type CalculationRepository struct {
	DB *gorm.DB
}

func (r *CalculationRepository) GetDraftByUser(userID uint) *model.PowerCalculation {
	var calc model.PowerCalculation
	if r.DB.Where("created_by = ? AND status = ?", userID, model.StatusDraft).First(&calc).Error != nil {
		return nil
	}
	return &calc
}

func (r *CalculationRepository) CreateDraft(userID uint) *model.PowerCalculation {
	calc := &model.PowerCalculation{
		Status:    model.StatusDraft,
		CreatedBy: userID,
	}
	r.DB.Create(calc)
	return calc
}

func (r *CalculationRepository) AddDeviceToCalculation(calcID, deviceID uint, quantity int) error {
	var existing model.PowerCalculationDevice
	err := r.DB.Where("calculation_id = ? AND device_id = ?", calcID, deviceID).First(&existing).Error
	if err != nil {
		cd := model.PowerCalculationDevice{
			CalculationID: calcID,
			DeviceID:      deviceID,
			Quantity:      quantity,
		}
		return r.DB.Create(&cd).Error
	}

	return r.DB.Model(&existing).
		Update("quantity", existing.Quantity+quantity).Error
}

func (r *CalculationRepository) GetCalculationWithDevices(calcID uint) (*model.PowerCalculation, error) {
	var calc model.PowerCalculation
	if err := r.DB.First(&calc, calcID).Error; err != nil {
		return nil, err
	}

	if calc.CreatedBy != 0 {
		var user model.User
		r.DB.Select("id, username").Where("id = ?", calc.CreatedBy).First(&user)
		calc.User = user
	}

	type DeviceWithQty struct {
		model.Device
		Quantity int
	}

	var devicesWithQty []DeviceWithQty
	r.DB.Table("power_calculation_devices").
		Select("devices.*, power_calculation_devices.quantity").
		Joins("JOIN devices ON devices.id = power_calculation_devices.device_id").
		Where("power_calculation_devices.calculation_id = ?", calcID).
		Scan(&devicesWithQty)

	devices := make([]model.Device, len(devicesWithQty))
	quantities := make(map[uint]int)
	for i, dq := range devicesWithQty {
		devices[i] = dq.Device
		quantities[dq.ID] = dq.Quantity
	}
	calc.Devices = devices
	calc.DeviceQuantities = quantities

	return &calc, nil
}

func (r *CalculationRepository) SoftDeleteCalculationSQL(calcID uint) error {
	return r.DB.Exec("UPDATE power_calculations SET status = 'deleted' WHERE id = ?", calcID).Error
}

func (r *CalculationRepository) GetTotalItemsInDraft(userID uint) int {
	var total int64
	err := r.DB.Table("power_calculation_devices").
		Select("COALESCE(SUM(quantity), 0)").
		Joins("JOIN power_calculations ON power_calculations.id = power_calculation_devices.calculation_id").
		Where("power_calculations.created_by = ? AND power_calculations.status = ?", userID, model.StatusDraft).
		Scan(&total).Error
	if err != nil {
		return 0
	}
	return int(total)
}

func (r *CalculationRepository) UpdateCalculation(calc *model.PowerCalculation) error {
	return r.DB.Save(calc).Error
}

func (r *CalculationRepository) GetCalculationsFiltered(status []model.CalculationStatus, fromDate, toDate *time.Time) ([]model.PowerCalculation, error) {
	query := r.DB.Where("status != ?", model.StatusDeleted).Where("status != ?", model.StatusDraft)

	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}
	if fromDate != nil {
		query = query.Where("formed_at >= ?", fromDate)
	}
	if toDate != nil {
		query = query.Where("formed_at <= ?", toDate)
	}

	var calcs []model.PowerCalculation
	err := query.Preload("User").Preload("Moderator").Find(&calcs).Error
	return calcs, err
}

func (r *CalculationRepository) UpdateDeviceQuantity(calcID, deviceID uint, quantity int) error {
	return r.DB.Model(&model.PowerCalculationDevice{}).
		Where("calculation_id = ? AND device_id = ?", calcID, deviceID).
		Update("quantity", quantity).Error
}

func (r *CalculationRepository) RemoveDeviceFromCalculation(calcID, deviceID uint) error {
	return r.DB.Where("calculation_id = ? AND device_id = ?", calcID, deviceID).
		Delete(&model.PowerCalculationDevice{}).Error
}

func (r *CalculationRepository) GetCalculationsByUser(userID uint, statuses []model.CalculationStatus, fromDate, toDate *time.Time) ([]model.PowerCalculation, error) {
	query := r.DB.Where("created_by = ? AND status != ? AND status != ?", userID, model.StatusDeleted, model.StatusDraft)

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if fromDate != nil {
		query = query.Where("formed_at >= ?", fromDate)
	}
	if toDate != nil {
		query = query.Where("formed_at <= ?", toDate)
	}

	var calcs []model.PowerCalculation
	err := query.Preload("User").Preload("Moderator").Find(&calcs).Error
	return calcs, err
}

func (r *CalculationRepository) GetPublicCalculations(statuses []model.CalculationStatus, fromDate, toDate *time.Time) ([]model.PowerCalculation, error) {
	allowed := []model.CalculationStatus{model.StatusCompleted, model.StatusRejected}
	query := r.DB.Where("status IN ?", allowed)

	if len(statuses) > 0 {
		filtered := []model.CalculationStatus{}
		for _, s := range statuses {
			for _, a := range allowed {
				if s == a {
					filtered = append(filtered, s)
					break
				}
			}
		}
		if len(filtered) > 0 {
			query = query.Where("status IN ?", filtered)
		}
	}
	if fromDate != nil {
		query = query.Where("formed_at >= ?", fromDate)
	}
	if toDate != nil {
		query = query.Where("formed_at <= ?", toDate)
	}

	var calcs []model.PowerCalculation
	err := query.Preload("User").Preload("Moderator").Find(&calcs).Error
	return calcs, err
}
