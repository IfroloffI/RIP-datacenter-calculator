package http

import (
	"net/http"
	"strconv"

	"datacenter-calc/internal/auth"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"

	"github.com/gin-gonic/gin"
)

type MMHandler struct {
	CalcRepo *repo.CalculationRepository
}

func NewMMHandler(calcRepo *repo.CalculationRepository) *MMHandler {
	return &MMHandler{CalcRepo: calcRepo}
}

// AddDeviceToCalculation godoc
// @Summary Добавить услугу в заявку
// @Description Добавляет оборудование в черновик заявки
// @Tags power_calculations-devices
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body AddDeviceRequest true "Данные услуги"
// @Success 201 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /power-calculations/{id}/devices [post]
func (h *MMHandler) AddDeviceToCalculation(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	calcID, _ := strconv.Atoi(c.Param("id"))

	calc, err := h.CalcRepo.GetCalculationWithDevices(uint(calcID))
	if err != nil || calc.CreatedBy != userID || calc.Status != model.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid calculation"})
		return
	}

	var req struct {
		DeviceID uint `json:"device_id" binding:"required"`
		Quantity int  `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.CalcRepo.AddDeviceToCalculation(uint(calcID), req.DeviceID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "added"})
}

// UpdateDeviceInCalculation godoc
// @Summary Изменить количество услуги в заявке
// @Description Обновляет количество оборудования в заявке
// @Tags power_calculations-devices
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param device_id path int true "ID услуги"
// @Param request body UpdateDeviceQuantityRequest true "Новое количество"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /power-calculations/{id}/devices/{device_id} [put]
func (h *MMHandler) UpdateDeviceInCalculation(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	calcID, _ := strconv.Atoi(c.Param("id"))
	deviceID, _ := strconv.Atoi(c.Param("device_id"))

	calc, err := h.CalcRepo.GetCalculationWithDevices(uint(calcID))
	if err != nil || calc.CreatedBy != userID || calc.Status != model.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid calculation"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.CalcRepo.UpdateDeviceQuantity(uint(calcID), uint(deviceID), req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// RemoveDeviceFromCalculation godoc
// @Summary Удалить услугу из заявки
// @Description Удаляет оборудование из заявки
// @Tags power_calculations-devices
// @Security Bearer
// @Produce json
// @Param id path int true "ID заявки"
// @Param device_id path int true "ID услуги"
// @Success 200 {object} SuccessMessage
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /power-calculations/{id}/devices/{device_id} [delete]
func (h *MMHandler) RemoveDeviceFromCalculation(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	calcID, _ := strconv.Atoi(c.Param("id"))
	deviceID, _ := strconv.Atoi(c.Param("device_id"))

	calc, err := h.CalcRepo.GetCalculationWithDevices(uint(calcID))
	if err != nil || calc.CreatedBy != userID || calc.Status != model.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid calculation"})
		return
	}

	err = h.CalcRepo.RemoveDeviceFromCalculation(uint(calcID), uint(deviceID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
