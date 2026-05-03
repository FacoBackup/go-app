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
	hh := handler.NewHealthHandler()

	mux := http.NewServeMux()

	handler.RegisterRoutes(mux, h, hh)

	// Métrica simples (log periódico)
	go func() {
		for range time.Tick(1 * time.Minute) {
			count, _ := repo.GetMeasurementsCount("", time.Time{}, time.Now())
			slog.Info("Metric: Total measurements in memory", "count", count)
		}
	}()

	slog.Info("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler.CORSMiddleware(mux)); err != nil {
		slog.Error("Error starting server", "error", err)
	}
}
