package handler

import (
	"encoding/json"
	"healthgo/backend/domain"
	"healthgo/backend/service"
	"log/slog"
	"net/http"
	"time"
)

type MeasurementHandler struct {
	service *service.MeasurementService
}

func NewMeasurementHandler(service *service.MeasurementService) *MeasurementHandler {
	return &MeasurementHandler{service: service}
}

func (h *MeasurementHandler) PostMeasurements(w http.ResponseWriter, r *http.Request) {
	var batch []domain.Measurement
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		slog.Error("Failed to decode batch", "error", err)
		http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(batch) == 0 || len(batch) > 500 {
		slog.Warn("Invalid batch size", "size", len(batch))
		http.Error(w, "Batch size must be between 1 and 500", http.StatusBadRequest)
		return
	}

	for i, m := range batch {
		if m.ValuePPM < 0 {
			slog.Warn("Negative value_ppm", "index", i, "value", m.ValuePPM)
			http.Error(w, "value_ppm cannot be negative", http.StatusBadRequest)
			return
		}
		if m.DeviceID == "" || m.Gas == "" || m.Timestamp.IsZero() {
			slog.Warn("Missing required fields", "index", i)
			http.Error(w, "device_id, gas and timestamp are required for all measurements", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.AddMeasurements(batch); err != nil {
		slog.Error("Failed to add measurements", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Batch processed successfully", "count", len(batch))
	w.WriteHeader(http.StatusAccepted)
}

func (h *MeasurementHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	gas := domain.GasType(r.URL.Query().Get("gas"))
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if deviceID == "" || gas == "" || startStr == "" || endStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "Invalid start timestamp", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, "Invalid end timestamp", http.StatusBadRequest)
		return
	}

	series, err := h.service.GetSeries(deviceID, gas, start, end)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(series)
}

func (h *MeasurementHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	if deviceID == "" {
		http.Error(w, "Missing device_id", http.StatusBadRequest)
		return
	}

	health, err := h.service.GetDeviceHealth(deviceID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Estrutura adicional para operações no frontend
func (h *MeasurementHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.service.GetDevices()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// Estrutura adicional para operações no frontend
func (h *MeasurementHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	if deviceID == "" {
		http.Error(w, "Missing device_id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteDevice(deviceID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
