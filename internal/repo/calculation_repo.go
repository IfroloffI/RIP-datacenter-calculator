package repo

import (
	"datacenter-calc/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CalculationRepository struct {
	DB *gorm.DB
}

func (r *CalculationRepository) GetDraftByUser(userID uint) *model.Calculation {
	var calc model.Calculation
	if r.DB.Where("created_by = ? AND status = ?", userID, model.StatusDraft).First(&calc).Error != nil {
		return nil
	}
	return &calc
}

func (r *CalculationRepository) CreateDraft(userID uint) *model.Calculation {
	calc := &model.Calculation{
		Status:    model.StatusDraft,
		CreatedBy: userID,
	}
	r.DB.Create(calc)
	return calc
}

func (r *CalculationRepository) AddDeviceToCalculation(calcID, deviceID uint, quantity int) error {
	cd := model.CalculationDevice{
		CalculationID: calcID,
		DeviceID:      deviceID,
		Quantity:      quantity,
	}
	return r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "calculation_id"}, {Name: "device_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity": gorm.Expr("calculation_devices.quantity + ?", quantity),
		}),
	}).Create(&cd).Error
}

func (r *CalculationRepository) GetCalculationWithDevices(calcID uint) (*model.Calculation, error) {
	var calc model.Calculation
	err := r.DB.Preload("Devices").Preload("User").First(&calc, calcID).Error
	if err != nil {
		return nil, err
	}
	return &calc, nil
}

func (r *CalculationRepository) SoftDeleteCalculationSQL(calcID uint) error {
	return r.DB.Exec("UPDATE calculations SET status = 'deleted' WHERE id = ?", calcID).Error
}

func (r *CalculationRepository) GetTotalItemsInDraft(userID uint) int {
	var total int64
	err := r.DB.Table("calculation_devices").
		Select("COALESCE(SUM(quantity), 0)").
		Joins("JOIN calculations ON calculations.id = calculation_devices.calculation_id").
		Where("calculations.created_by = ? AND calculations.status = ?", userID, model.StatusDraft).
		Scan(&total).Error
	if err != nil {
		return 0
	}
	return int(total)
}
