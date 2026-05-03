package handler

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *MeasurementHandler, hh *HealthHandler) {
	mux.HandleFunc("POST /v1/measurements", h.PostMeasurements)
	mux.HandleFunc("GET /v1/devices", h.GetDevices)
	mux.HandleFunc("GET /v1/devices/{device_id}/series", h.GetSeries)
	mux.HandleFunc("GET /v1/devices/{device_id}/health", h.GetHealth)
	mux.HandleFunc("DELETE /v1/devices/{device_id}", h.DeleteDevice)

	mux.HandleFunc("GET /healthz", hh.Healthz)
	mux.HandleFunc("GET /readyz", hh.Readyz)
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
