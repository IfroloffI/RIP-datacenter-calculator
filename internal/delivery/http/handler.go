package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"datacenter-calc/internal/model"
	"datacenter-calc/internal/usecase"
)

type Handler struct {
	Calculator *usecase.PowerCalculator
	Order      *model.Order
	MinIOURL   string
}

func NewHandler(calculator *usecase.PowerCalculator) *Handler {
	return &Handler{
		Calculator: calculator,
		Order: &model.Order{
			ID:        1,
			DeviceIDs: []int{1, 2},
			CreatedAt: "13.09.2025 12:00",
		},
		MinIOURL: "http://localhost:9000/devices",
	}
}

func (h *Handler) ServeStatic() http.Handler {
	return http.FileServer(http.Dir("static/"))
}

func (h *Handler) render(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("templates/base.html", "templates/"+tmpl)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var buf strings.Builder
	err = t.ExecuteTemplate(&buf, "base.html", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(buf.String()))
}

func (h *Handler) Devices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	all := h.Calculator.Repo.GetAll()

	var filtered []model.Device
	if query != "" {
		for _, d := range all {
			if strings.Contains(strings.ToLower(d.Name), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(d.Category), strings.ToLower(query)) {
				filtered = append(filtered, d)
			}
		}
	} else {
		filtered = all
	}

	data := map[string]interface{}{
		"Devices":     filtered,
		"Order":       h.Order,
		"TotalItems":  len(h.Order.DeviceIDs),
		"MinIOURL":    h.MinIOURL,
		"SearchQuery": query,
	}

	h.render(w, "devices.html", data)
}

func (h *Handler) DeviceDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/device/"):]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	device := h.Calculator.Repo.GetByID(id)
	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	data := map[string]interface{}{
		"Device":     device,
		"Order":      h.Order,
		"TotalItems": len(h.Order.DeviceIDs),
		"MinIOURL":   h.MinIOURL,
	}

	h.render(w, "device.html", data)
}

func (h *Handler) PowerCalc(w http.ResponseWriter, r *http.Request) {
	base, calculated := h.Calculator.CalculateTotalPower(*h.Order)
	devices := h.Calculator.GetDevicesInOrder(*h.Order)
	fmt.Println(h.Order)
	data := map[string]interface{}{
		"Order":      h.Order,
		"BasePower":  base,
		"PUEPower":   calculated,
		"Devices":    devices,
		"TotalItems": len(h.Order.DeviceIDs),
		"MinIOURL":   h.MinIOURL,
	}

	h.render(w, "calc.html", data)
}
