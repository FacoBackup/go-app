package main

import (
	"healthgo/backend/handler"
	"healthgo/backend/repository"
	"healthgo/backend/service"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	// Configurar log estruturado (JSON)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	repo := repository.NewInMemoryRepository()
	svc := service.NewMeasurementService(repo)
	h := handler.NewMeasurementHandler(svc)

	mux := http.NewServeMux()

	// Middleware de CORS simplificado para desenvolvimento
	corsMiddleware := func(next http.Handler) http.Handler {
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

	mux.HandleFunc("POST /v1/measurements", h.PostMeasurements)
	mux.HandleFunc("GET /v1/devices", h.GetDevices)
	mux.HandleFunc("GET /v1/devices/{device_id}/series", h.GetSeries)
	mux.HandleFunc("GET /v1/devices/{device_id}/health", h.GetHealth)
	mux.HandleFunc("DELETE /v1/devices/{device_id}", h.DeleteDevice)

	// Endpoints de saúde
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		// Como o banco é em memória e local, sempre está pronto após iniciar
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// Métrica simples (log periódico)
	go func() {
		for range time.Tick(1 * time.Minute) {
			count, _ := repo.GetMeasurementsCount("", time.Time{}, time.Now())
			slog.Info("Metric: Total measurements in memory", "count", count)
		}
	}()

	slog.Info("Server starting on :8080")
	if err := http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		slog.Error("Error starting server", "error", err)
	}
}
