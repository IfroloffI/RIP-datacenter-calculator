package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"datacenter-calc/internal/auth"
	"datacenter-calc/internal/dto"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CalculationHandler struct {
	Calculator *usecase.PowerCalculator
	MinIOURL   string
	AsyncToken string
}

func NewCalculationHandler(calc *usecase.PowerCalculator, minioURL string, asyncToken string) *CalculationHandler {
	return &CalculationHandler{Calculator: calc, MinIOURL: minioURL, AsyncToken: asyncToken}
}

// GetCalculations godoc
// @Summary Получить список заявок
// @Description Для гостей — только completed/rejected. Для пользователей — только свои. Для модераторов — все.
// @Tags power-calculations
// @Produce json
// @Param status query string false "Фильтр по статусу (через запятую)"
// @Param from_date query string false "Дата формирования от (формат: 2025-01-01)"
// @Param to_date query string false "Дата формирования до (формат: 2025-12-31)"
// @Success 200 {array} PowerCalculationResponse
// @Router /power-calculations [get]
func (h *CalculationHandler) GetCalculations(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	role := auth.UserRoleFromContext(c)

	var statuses []model.CalculationStatus
	if s := c.Query("status"); s != "" {
		for _, st := range strings.Split(s, ",") {
			statuses = append(statuses, model.CalculationStatus(st))
		}
	}

	var fromDate, toDate *time.Time
	if f := c.Query("from_date"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			fromDate = &t
		}
	}
	if t := c.Query("to_date"); t != "" {
		if tm, err := time.Parse("2006-01-02", t); err == nil {
			toDate = &tm
		}
	}

	var calcs []model.PowerCalculation
	var err error

	if role == model.RoleModerator {
		calcs, err = h.Calculator.CalculationRepo.GetCalculationsFiltered(statuses, fromDate, toDate)
	} else if userID != 0 {
		calcs, err = h.Calculator.CalculationRepo.GetCalculationsByUser(userID, statuses, fromDate, toDate)
	} else {
		calcs, err = h.Calculator.CalculationRepo.GetPublicCalculations(statuses, fromDate, toDate)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	resp := make([]model.PowerCalculationResponse, len(calcs))
	for i, calc := range calcs {
		resp[i] = calc.ToResponse(h.MinIOURL)
	}
	c.JSON(http.StatusOK, resp)
}

// GetCalculation godoc
// @Summary Получить заявку по ID
// @Description Возвращает полную информацию о заявке с услугами и количеством
// @Tags power-calculations
// @Security Bearer
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} PowerCalculationResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /power-calculations/{id} [get]
func (h *CalculationHandler) GetCalculation(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := auth.UserIDFromContext(c)
	role := auth.UserRoleFromContext(c)

	calc, err := h.Calculator.CalculationRepo.GetCalculationWithDevices(uint(id))
	if err != nil || calc.Status == model.StatusDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if role != model.RoleModerator && (userID == 0 || calc.CreatedBy != userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	devicesResp := make([]model.DeviceWithQuantity, len(calc.Devices))
	for i, d := range calc.Devices {
		img := ""
		if d.ImageURL != "" {
			img = h.MinIOURL + "/" + d.ImageURL
		}
		devicesResp[i] = model.DeviceWithQuantity{
			ID:          d.ID,
			Name:        d.Name,
			PowerWatt:   d.PowerWatt,
			Description: d.Description,
			ImageURL:    img,
			Category:    d.Category,
			Quantity:    calc.DeviceQuantities[d.ID],
		}
	}

	creator := calc.User.Username
	var moderator *string
	if calc.Moderator != nil {
		m := calc.Moderator.Username
		moderator = &m
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             calc.ID,
		"status":         calc.Status,
		"created_at":     calc.CreatedAt,
		"formed_at":      calc.FormedAt,
		"completed_at":   calc.CompletedAt,
		"total_power":    calc.TotalPower,
		"creator":        creator,
		"moderator":      moderator,
		"dc_name":        calc.DCName,
		"dc_description": calc.DCDescription,
		"pue":            calc.PUE,
		"devices":        devicesResp,
	})
}

// FormCalculation godoc
// @Summary Сформировать заявку
// @Description Переводит черновик в статус "formed" (только создатель)
// @Tags power-calculations
// @Security Bearer
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /power-calculations/{id}/form [put]
func (h *CalculationHandler) FormCalculation(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := auth.UserIDFromContext(c)

	calc, err := h.Calculator.CalculationRepo.GetCalculationWithDevices(uint(id))
	if err != nil || calc.CreatedBy != userID || calc.Status != model.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{"error": "only creator can form draft"})
		return
	}

	err = h.Calculator.FormCalculation(uint(id), true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "formed"})
}

// CompleteCalculation godoc
// @Summary Завершить заявку
// @Description Вычисляет итоговую мощность, завершает заявку (только модератор)
// @Tags power-calculations
// @Security Bearer
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /power-calculations/{id}/complete [put]
func (h *CalculationHandler) CompleteCalculation(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.Calculator.CompleteCalculation(uint(id), auth.UserIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

// RejectCalculation godoc
// @Summary Отклонить заявку
// @Description Отклоняет сформированную заявку (только модератор)
// @Tags power-calculations
// @Security Bearer
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /power-calculations/{id}/reject [put]
func (h *CalculationHandler) RejectCalculation(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.Calculator.RejectCalculation(uint(id), auth.UserIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

// DeleteCalculation godoc
// @Summary Удалить заявку
// @Description Удаляет черновик (только создатель)
// @Tags power-calculations
// @Security Bearer
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} SuccessMessage
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /power-calculations/{id} [delete]
func (h *CalculationHandler) DeleteCalculation(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := auth.UserIDFromContext(c)

	calc, err := h.Calculator.CalculationRepo.GetCalculationWithDevices(uint(id))
	if err != nil || calc.CreatedBy != userID || calc.Status != model.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{"error": "only creator can delete draft"})
		return
	}

	h.Calculator.CalculationRepo.SoftDeleteCalculationSQL(uint(id))
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// UpdateCalculationFields godoc
// @Summary Обновить поля заявки
// @Description Изменяет параметры ЦОД (только для черновиков)
// @Tags power-calculations
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body UpdateCalculationRequest false "Поля для обновления"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /power-calculations/{id} [put]
func (h *CalculationHandler) UpdateCalculationFields(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID := auth.UserIDFromContext(c)
	calc, err := h.Calculator.CalculationRepo.GetCalculationWithDevices(uint(id))
	if err != nil || calc.Status == model.StatusDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if calc.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only creator can edit"})
		return
	}

	if calc.Status != model.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{"error": "only draft can be edited"})
		return
	}

	var req struct {
		DCName        *string  `json:"dc_name"`
		DCDescription *string  `json:"dc_description"`
		PUE           *float64 `json:"pue"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.DCName != nil {
		updates["dc_name"] = *req.DCName
	}
	if req.DCDescription != nil {
		updates["dc_description"] = *req.DCDescription
	}
	if req.PUE != nil {
		updates["pue"] = *req.PUE
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	result := h.Calculator.CalculationRepo.DB.Model(&model.PowerCalculation{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ProcessAsyncResult godoc
// @Summary Применить результат асинхронного расчёта
// @Description Внутренний вебхук, вызывается асинхронным сервисом после завершения расчёта. НЕ использовать из пользовательского/админского интерфейса.
// @Tags power-calculations
// @Accept json
// @Produce json
// @Param X-Async-Token header string true "Секретный токен асинхронного сервиса"
// @Param request body dto.AsyncResultRequest true "Результат асинхронного расчёта"
// @Success 200 {object} SuccessMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse "Неверный токен"
// @Router /power-calculations/process-result [put]
func (h *CalculationHandler) ProcessAsyncResult(c *gin.Context) {
	var req dto.AsyncResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.GetHeader("X-Async-Token")
	if token != h.AsyncToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "bad async token"})
		return
	}

	if err := h.Calculator.ProcessAsyncResult(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "async result applied"})
}
