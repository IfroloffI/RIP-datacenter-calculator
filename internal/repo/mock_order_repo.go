package repo

import "datacenter-calc/internal/model"

var Orders = []model.Order{
	{
		ID:        1,
		DeviceIDs: []int{1, 2},
		CreatedAt: "13.09.2025 12:00",
	},
}

type OrderRepository struct{}

func (r *OrderRepository) GetByID(id int) *model.Order {
	for _, o := range Orders {
		if o.ID == id {
			return &o
		}
	}
	return nil
}

func (r *OrderRepository) GetAll() []model.Order {
	return Orders
}

func (r *OrderRepository) GetCurrentOrder() *model.Order {
	return r.GetByID(1)
}
