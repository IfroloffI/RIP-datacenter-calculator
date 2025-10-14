package http

import (
	"context"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"datacenter-calc/internal/minio"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/repo"

	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
)

type DeviceHandler struct {
	Repo     *repo.DeviceRepository
	MinIO    *minio.MinIOClient
	MinIOURL string
}

func NewDeviceHandler(repo *repo.DeviceRepository, minioClient *minio.MinIOClient, minioURL string) *DeviceHandler {
	return &DeviceHandler{Repo: repo, MinIO: minioClient, MinIOURL: minioURL}
}

// GET /api/devices
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	q := c.Query("q")
	devices := h.Repo.GetAllActive()
	if q != "" {
		filtered := []model.Device{}
		for _, d := range devices {
			if strings.Contains(strings.ToLower(d.Name), strings.ToLower(q)) ||
				strings.Contains(strings.ToLower(d.Category), strings.ToLower(q)) {
				filtered = append(filtered, d)
			}
		}
		devices = filtered
	}
	resp := make([]model.DeviceResponse, len(devices))
	for i, d := range devices {
		resp[i] = d.ToResponse(h.MinIOURL)
	}
	c.JSON(http.StatusOK, resp)
}

// POST /api/devices
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		PowerWatt   int    `json:"power_watt" binding:"required"`
		Description string `json:"description"`
		Category    string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	device := model.Device{
		Name:        req.Name,
		PowerWatt:   req.PowerWatt,
		Description: req.Description,
		Category:    req.Category,
		IsDeleted:   false,
	}
	result := h.Repo.DB.Create(&device)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusCreated, device.ToResponse(h.MinIOURL))
}

// GET /api/devices/:id
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	device := h.Repo.GetByID(uint(id))
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}
	c.JSON(http.StatusOK, device.ToResponse(h.MinIOURL))
}

// PUT /api/devices/:id
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name        *string `json:"name"`
		PowerWatt   *int    `json:"power_watt"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.PowerWatt != nil {
		updates["power_watt"] = *req.PowerWatt
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	result := h.Repo.DB.Model(&model.Device{}).Where("id = ? AND is_deleted = false", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found or deleted"})
		return
	}

	updated := h.Repo.GetByID(uint(id))
	c.JSON(http.StatusOK, updated.ToResponse(h.MinIOURL))
}

// DELETE /api/devices/:id
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	device := h.Repo.GetByID(uint(id))
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	if device.ImageURL != "" {
		if err := h.MinIO.DeleteFile(context.Background(), device.ImageURL); err != nil {
			// TODO: логи
		}
	}

	result := h.Repo.DB.Model(&model.Device{}).Where("id = ?", id).Update("is_deleted", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// POST /api/devices/:id/image
func (h *DeviceHandler) UploadDeviceImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	device := h.Repo.GetByID(uint(id))
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	// Генерируем имя
	ext := filepath.Ext(file.Filename)
	safeName := sanitizeFileName(file.Filename[:len(file.Filename)-len(ext)])
	uuidStr := uuid.New().String()
	objectName := safeName + "_" + uuidStr + ext

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	if device.ImageURL != "" {
		h.MinIO.DeleteFile(context.Background(), device.ImageURL)
	}

	err = h.MinIO.UploadFile(context.Background(), objectName, src, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload to MinIO"})
		return
	}

	result := h.Repo.DB.Model(&model.Device{}).Where("id = ?", id).Update("image_url", objectName)
	if result.Error != nil {
		// Откат
		h.MinIO.DeleteFile(context.Background(), objectName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_url": h.MinIOURL + "/" + objectName,
	})
}

// sanitizeFileName оставляет только латинские буквы, цифры, дефисы и подчёркивания
func sanitizeFileName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return reg.ReplaceAllString(name, "_")
}
