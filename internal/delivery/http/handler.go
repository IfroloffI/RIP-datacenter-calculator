package http

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"datacenter-calc/internal/model"
	"datacenter-calc/internal/usecase"
)

type Handler struct {
	Calculator *usecase.PowerCalculator
	MinIOURL   string
}

func NewHandler(calculator *usecase.PowerCalculator) *Handler {
	return &Handler{
		Calculator: calculator,
		MinIOURL:   "http://127.0.0.1:9000/devices",
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getCommonData() map[string]interface{} {
	userID := uint(1)
	totalItems := h.Calculator.OrderRepo.GetTotalItemsInDraft(userID)
	return map[string]interface{}{
		"TotalItems": totalItems,
		"MinIOURL":   h.MinIOURL,
	}
}

func (h *Handler) Devices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	all := h.Calculator.DeviceRepo.GetAllActive()

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

	data := h.getCommonData()
	data["Devices"] = filtered
	data["SearchQuery"] = query

	h.render(w, "devices.html", data)
}

func (h *Handler) DeviceDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/device/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	device := h.Calculator.DeviceRepo.GetByID(uint(id))
	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	data := h.getCommonData()
	data["Device"] = device

	h.render(w, "device.html", data)
}

func (h *Handler) PowerCalc(w http.ResponseWriter, r *http.Request) {
	userID := uint(1)
	draft := h.Calculator.OrderRepo.GetDraftByUser(userID)
	if draft == nil {
		data := h.getCommonData()
		data["Order"] = model.Order{Status: model.StatusDraft, CreatedBy: userID}
		data["BasePower"] = 0
		data["PUEPower"] = 0
		data["Devices"] = []model.DeviceWithQuantityAndIPW{}
		data["PUE"] = 1.5
		h.render(w, "calc.html", data)
		return
	}

	order, err := h.Calculator.OrderRepo.GetOrderWithDevices(draft.ID)
	if err != nil {
		http.Error(w, "Order load error", http.StatusInternalServerError)
		return
	}

	devices := h.Calculator.GetDevicesInOrder(*order)
	base, calculated := h.Calculator.CalculateTotalPower(devices, 1.5)

	data := h.getCommonData()
	data["Order"] = *order
	data["BasePower"] = base
	data["PUEPower"] = calculated
	data["Devices"] = devices
	data["PUE"] = 1.5

	h.render(w, "calc.html", data)
}

func (h *Handler) AddDeviceToOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID, err := strconv.Atoi(r.FormValue("device_id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID := uint(1)
	draft := h.Calculator.OrderRepo.GetDraftByUser(userID)
	if draft == nil {
		draft = h.Calculator.OrderRepo.CreateDraft(userID)
	}

	h.Calculator.OrderRepo.AddDeviceToOrder(draft.ID, uint(deviceID), 1)

	http.Redirect(w, r, "/power-calc", http.StatusSeeOther)
}

func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := uint(1)
	draft := h.Calculator.OrderRepo.GetDraftByUser(userID)
	if draft != nil {
		h.Calculator.OrderRepo.SoftDeleteOrderSQL(draft.ID)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
