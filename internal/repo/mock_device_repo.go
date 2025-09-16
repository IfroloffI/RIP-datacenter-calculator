package repo

import "datacenter-calc/internal/model"

var Devices = []model.Device{
	{
		ID:          1,
		Name:        "Сервер Dell R760",
		PowerWatt:   850,
		Description: "Мощный сервер для критичных workloads. 2x Intel Xeon, до 3TB RAM.",
		ImageURL:    "dell-r760.jpg",
		Category:    "Сервер",
	},
	{
		ID:          2,
		Name:        "Коммутатор Dell N2024",
		PowerWatt:   30,
		Description: "Коммутатор уровня доступа, 24 порта 1G.",
		ImageURL:    "dell-n2024.jpg",
		Category:    "Коммутатор",
	},
	{
		ID:          3,
		Name:        "Коммутатор Dell N3248TE-ON",
		PowerWatt:   75,
		Description: "Коммутатор уровня агрегации, 48 портов 10G.",
		ImageURL:    "dell-n3248te.jpg", //TODO Image
		Category:    "Коммутатор",
	},
	{
		ID:          4,
		Name:        "Сервер Dell PowerEdge R660",
		PowerWatt:   650,
		Description: "Универсальный сервер для виртуализации. 2x Intel Xeon, до 2TB RAM.",
		ImageURL:    "dell-r660.jpg",
		Category:    "Сервер",
	},
	{
		ID:          5,
		Name:        "СХД Dell PowerVault ME5024",
		PowerWatt:   1400,
		Description: "Система хранения данных начального уровня, 24 дисковых слота.",
		ImageURL:    "dell-me5024.jpg",
		Category:    "СХД",
	},
	{
		ID:          6,
		Name:        "Сервер Dell PowerEdge R7625",
		PowerWatt:   1200,
		Description: "Сервер на базе AMD EPYC для высокопроизводительных вычислений.",
		ImageURL:    "dell-r7625.jpg",
		Category:    "Сервер",
	},
}

type DeviceRepository struct{}

func (r *DeviceRepository) GetAll() []model.Device {
	return Devices
}

func (r *DeviceRepository) GetByID(id int) *model.Device {
	for _, d := range Devices {
		if d.ID == id {
			return &d
		}
	}
	return nil
}
