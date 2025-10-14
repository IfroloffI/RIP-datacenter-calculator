package auth

import (
	"datacenter-calc/internal/model"

	"github.com/gin-gonic/gin"
)

func UserIDFromContext(c *gin.Context) uint {
	userID, _ := c.Get("userID")
	if userID == nil {
		return 0
	}
	return userID.(uint)
}

func UserRoleFromContext(c *gin.Context) model.UserRole {
	role, _ := c.Get("userRole")
	if role == nil {
		return ""
	}
	return role.(model.UserRole)
}
