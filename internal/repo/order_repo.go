package repo

import (
	"datacenter-calc/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	DB *gorm.DB
}

func (r *OrderRepository) GetDraftByUser(userID uint) *model.Order {
	var order model.Order
	if r.DB.Where("created_by = ? AND status = ?", userID, model.StatusDraft).First(&order).Error != nil {
		return nil
	}
	return &order
}

func (r *OrderRepository) CreateDraft(userID uint) *model.Order {
	order := &model.Order{
		Status:    model.StatusDraft,
		CreatedBy: userID,
	}
	r.DB.Create(order)
	return order
}

func (r *OrderRepository) AddDeviceToOrder(orderID, deviceID uint, quantity int) error {
	od := model.OrderDevice{
		OrderID:  orderID,
		DeviceID: deviceID,
		Quantity: quantity,
	}
	return r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "order_id"}, {Name: "device_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity": gorm.Expr("order_devices.quantity + ?", quantity),
		}),
	}).Create(&od).Error
}

func (r *OrderRepository) GetOrderWithDevices(orderID uint) (*model.Order, error) {
	var order model.Order
	err := r.DB.Preload("Devices").First(&order, orderID).Error
	return &order, err
}

func (r *OrderRepository) SoftDeleteOrderSQL(orderID uint) error {
	return r.DB.Exec("UPDATE orders SET status = 'deleted' WHERE id = ?", orderID).Error
}

func (r *OrderRepository) GetTotalItemsInDraft(userID uint) int {
	var count int64
	r.DB.Table("order_devices").
		Joins("JOIN orders ON orders.id = order_devices.order_id").
		Where("orders.created_by = ? AND orders.status = ?", userID, model.StatusDraft).
		Count(&count)
	return int(count)
}
