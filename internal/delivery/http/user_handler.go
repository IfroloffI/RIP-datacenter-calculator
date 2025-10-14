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
