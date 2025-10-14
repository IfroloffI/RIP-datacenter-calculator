package usecase

import (
	"context"
	"errors"
	"time"

	"datacenter-calc/internal/auth"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/redis"
	"datacenter-calc/internal/repo"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	UserRepo    *repo.UserRepo
	RedisClient *redis.RedisClient
	JWTExpMin   int
}

func NewUserUsecase(userRepo *repo.UserRepo, redisClient *redis.RedisClient, jwtExpMin int) *UserUsecase {
	return &UserUsecase{
		UserRepo:    userRepo,
		RedisClient: redisClient,
		JWTExpMin:   jwtExpMin,
	}
}

func (u *UserUsecase) Register(username, password string) error {
	existing, _ := u.UserRepo.GetUserByUsername(username)
	if existing != nil {
		return errors.New("username already exists")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Role:     model.RoleUser,
	}
	return u.UserRepo.CreateUser(user)
}

func (u *UserUsecase) Login(username, password string) (*model.User, error) {
	user, err := u.UserRepo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (u *UserUsecase) Logout(ctx context.Context, token string) error {
	claims, err := auth.ParseToken(token)
	if err != nil {
		return nil
	}
	ttl := time.Until(time.Unix(claims.ExpiresAt.Time.Unix(), 0))
	if ttl > 0 {
		return u.RedisClient.AddToBlacklist(ctx, token, ttl)
	}
	return nil
}

func (u *UserUsecase) GetMe(userID uint) (*model.User, error) {
	return u.UserRepo.GetUserByID(userID)
}

func (u *UserUsecase) UpdateMe(userID uint, updates map[string]interface{}) error {
	delete(updates, "id")
	delete(updates, "username")
	delete(updates, "role")
	return u.UserRepo.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
