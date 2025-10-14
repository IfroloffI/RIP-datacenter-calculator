package http

import (
	"net/http"

	"datacenter-calc/internal/auth"
	"datacenter-calc/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Usecase *usecase.UserUsecase
}

func NewUserHandler(usecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{Usecase: usecase}
}

// RegisterUser godoc
// @Summary Регистрация нового пользователя
// @Description Создаёт нового пользователя с ролью "user"
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные регистрации"
// @Success 201 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Router /users/register [post]
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Usecase.Register(req.Username, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user registered"})
}

// LoginUser godoc
// @Summary Аутентификация
// @Description Возвращает JWT-токен для авторизации
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.Usecase.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, _ := auth.GenerateToken(user.ID, user.Role, h.Usecase.JWTExpMin)
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}

// LogoutUser godoc
// @Summary Выход из системы
// @Description Добавляет токен в blacklist (Redis)
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Router /users/logout [post]
func (h *UserHandler) LogoutUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing Authorization header"})
		return
	}
	token := authHeader[len("Bearer "):]
	if err := h.Usecase.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// GetMe godoc
// @Summary Получить профиль текущего пользователя
// @Description Возвращает данные авторизованного пользователя
// @Tags users
// @Security Bearer
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	user, err := h.Usecase.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// UpdateMe godoc
// @Summary Обновить профиль
// @Description Обновляет поля профиля (например, username)
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest false "Поля для обновления"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	var req struct {
		Email *string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	if err := h.Usecase.UpdateMe(userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
