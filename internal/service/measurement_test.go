package service

import (
	"healthgo/backend/internal/domain"
	"healthgo/backend/internal/repository"
	"testing"
	"time"
)

func TestMeasurementService_GetSeries_Aggregation(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	svc := NewMeasurementService(repo)

	deviceID := "dev1"
	gas := domain.GasH2
	ts1 := time.Date(2024, 1, 1, 12, 0, 10, 0, time.UTC)
	ts2 := time.Date(2024, 1, 1, 12, 0, 50, 0, time.UTC)
	ts3 := time.Date(2024, 1, 1, 12, 1, 10, 0, time.UTC)

	repo.SaveBatch([]domain.Measurement{
		{DeviceID: deviceID, Timestamp: ts1, Gas: gas, ValuePPM: 10.0},
		{DeviceID: deviceID, Timestamp: ts2, Gas: gas, ValuePPM: 20.0},
		{DeviceID: deviceID, Timestamp: ts3, Gas: gas, ValuePPM: 30.0},
	})

	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 12, 2, 0, 0, time.UTC)

	series, _ := svc.GetSeries(deviceID, gas, start, end)

	if len(series) != 2 {
		t.Fatalf("Expected 2 aggregated points, got %d", len(series))
	}

	// Primeiro minuto: (10+20)/2 = 15
	if series[0].ValuePPM != 15.0 {
		t.Errorf("Expected first point to be 15.0, got %f", series[0].ValuePPM)
	}

	// Segundo minuto: 30
	if series[1].ValuePPM != 30.0 {
		t.Errorf("Expected second point to be 30.0, got %f", series[1].ValuePPM)
	}
}
