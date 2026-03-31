package repo

import (
	"errors"

	"gorm.io/gorm"

	"datacenter-calc/internal/model"
)

type UserRepo struct {
	DB *gorm.DB
}

func (r *UserRepo) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) CreateUser(user *model.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepo) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
