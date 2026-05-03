package repository

import (
	"healthgo/backend/internal/domain"
	"testing"
	"time"
)

func TestInMemoryRepository_SaveBatch_Idempotency(t *testing.T) {
	repo := NewInMemoryRepository()

	ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	measurements := []domain.Measurement{
		{DeviceID: "dev1", Timestamp: ts, Gas: domain.GasH2, ValuePPM: 10.0},
		{DeviceID: "dev1", Timestamp: ts, Gas: domain.GasH2, ValuePPM: 20.0}, // Duplicata (mesmo dev+ts+gas)
	}

	err := repo.SaveBatch(measurements)
	if err != nil {
		t.Fatalf("Failed to save batch: %v", err)
	}

	series, _ := repo.GetSeries("dev1", domain.GasH2, ts, ts.Add(time.Minute))
	if len(series) != 1 {
		t.Errorf("Expected 1 measurement due to idempotency, got %d", len(series))
	}
}
