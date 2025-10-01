package http

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"datacenter-calc/config"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Calculator    *usecase.PowerCalculator
	MinIOURL      string
	CurrentUserID uint
}

func NewHandler(calculator *usecase.PowerCalculator, cfg *config.Config) *Handler {
	return &Handler{
		Calculator:    calculator,
		MinIOURL:      cfg.MinIO.URL + "/" + cfg.MinIO.Bucket,
		CurrentUserID: 1,
	}
}

func (h *Handler) render(c *gin.Context, tmpl string, data gin.H) {
	t, err := template.ParseFiles("templates/base.html", "templates/"+tmpl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Template error: " + err.Error()})
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	err = t.ExecuteTemplate(c.Writer, "base.html", data)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Render error: " + err.Error()})
	}
}

func (h *Handler) getCommonData() gin.H {
	draft := h.Calculator.CalculationRepo.GetDraftByUser(h.CurrentUserID)
	draftID := uint(0)
	if draft != nil {
		draftID = draft.ID
	}
	totalItems := h.Calculator.CalculationRepo.GetTotalItemsInDraft(h.CurrentUserID)
	return gin.H{
		"TotalItems":         totalItems,
		"DraftCalculationID": draftID,
		"MinIOURL":           h.MinIOURL,
	}
}

func (h *Handler) Devices(c *gin.Context) {
	query := c.Query("q")
	all := h.Calculator.DeviceRepo.GetAllActive()
	var filtered []model.Device
	if query != "" {
		for _, d := range all {
			if containsCI(d.Name, query) || containsCI(d.Category, query) {
				filtered = append(filtered, d)
			}
		}
	} else {
		filtered = all
	}
	data := h.getCommonData()
	data["Devices"] = filtered
	data["SearchQuery"] = query
	h.render(c, "devices.html", data)
}

func (h *Handler) DeviceDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	device := h.Calculator.DeviceRepo.GetByID(uint(id))
	if device == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	data := h.getCommonData()
	data["Device"] = device
	h.render(c, "device.html", data)
}

func (h *Handler) CalculationDetail(c *gin.Context) {
	idStr := c.Param("id")
	calcID, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	calc, err := h.Calculator.CalculationRepo.GetCalculationWithDevices(uint(calcID))
	if err != nil || calc.Status == model.StatusDeleted {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if calc.CreatedBy != h.CurrentUserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	devices := h.Calculator.GetDevicesInCalculation(*calc)
	base, calculated := h.Calculator.CalculateTotalPower(devices, 1.5)
	data := gin.H{
		"Calculation": *calc,
		"BasePower":   base,
		"PUEPower":    calculated,
		"Devices":     devices,
		"PUE":         1.5,
		"MinIOURL":    h.MinIOURL,
		"TotalItems":  len(devices),
	}
	h.render(c, "calc.html", data)
}

func (h *Handler) AddToCalc(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.PostForm("device_id"))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	// Сохраняем параметр поиска
	searchQuery := c.PostForm("search_query")

	draft := h.Calculator.CalculationRepo.GetDraftByUser(h.CurrentUserID)
	if draft == nil {
		draft = h.Calculator.CalculationRepo.CreateDraft(h.CurrentUserID)
	}
	h.Calculator.CalculationRepo.AddDeviceToCalculation(draft.ID, uint(deviceID), 1)

	redirectURL := "/"
	if searchQuery != "" {
		redirectURL = "/?q=" + url.QueryEscape(searchQuery)
	}

	c.Redirect(http.StatusSeeOther, redirectURL)
}

func (h *Handler) AddDeviceToCalculation(c *gin.Context) {
	calcID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	deviceID, err := strconv.Atoi(c.PostForm("device_id"))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/calculation/"+c.Param("id"))
		return
	}
	h.Calculator.CalculationRepo.AddDeviceToCalculation(uint(calcID), uint(deviceID), 1)
	c.Redirect(http.StatusSeeOther, "/calculation/"+c.Param("id"))
}

func (h *Handler) DeleteCalculation(c *gin.Context) {
	calcID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	calc, err := h.Calculator.CalculationRepo.GetCalculationWithDevices(uint(calcID))
	if err != nil || calc.Status != model.StatusDraft || calc.CreatedBy != h.CurrentUserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	h.Calculator.CalculationRepo.SoftDeleteCalculationSQL(uint(calcID))
	c.Redirect(http.StatusSeeOther, "/")
}

func containsCI(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
