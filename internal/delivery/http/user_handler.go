// internal/delivery/http/user_handler.go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"token": "mock-token"})
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id":      1,
		"username":     "user",
		"is_moderator": false,
	})
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
