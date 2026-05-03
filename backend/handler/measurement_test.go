package handler

import (
	"bytes"
	"encoding/json"
	"healthgo/backend/domain"
	"healthgo/backend/repository"
	"healthgo/backend/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMeasurementHandler_PostMeasurements(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	svc := service.NewMeasurementService(repo)
	h := NewMeasurementHandler(svc)

	t.Run("Happy path", func(t *testing.T) {
		ts := time.Now().UTC()
		batch := []domain.Measurement{
			{DeviceID: "dev1", Timestamp: ts, Gas: domain.GasH2, ValuePPM: 10.5, PatientRef: "p1"},
		}
		body, _ := json.Marshal(batch)
		req := httptest.NewRequest("POST", "/v1/measurements", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.PostMeasurements(rr, req)

		if rr.Code != http.StatusAccepted {
			t.Errorf("Expected status 202, got %d", rr.Code)
		}

		// Verificar se foi salvo
		count, _ := repo.GetMeasurementsCount("dev1", ts.Add(-time.Second), ts.Add(time.Second))
		if count != 1 {
			t.Errorf("Expected 1 measurement saved, got %d", count)
		}
	})

	t.Run("Negative value", func(t *testing.T) {
		batch := []domain.Measurement{
			{DeviceID: "dev1", Timestamp: time.Now(), Gas: domain.GasH2, ValuePPM: -1.0},
		}
		body, _ := json.Marshal(batch)
		req := httptest.NewRequest("POST", "/v1/measurements", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.PostMeasurements(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/measurements", bytes.NewBufferString("invalid"))
		rr := httptest.NewRecorder()

		h.PostMeasurements(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})

	t.Run("Missing fields", func(t *testing.T) {
		batch := []domain.Measurement{
			{DeviceID: "", Timestamp: time.Now(), Gas: domain.GasH2, ValuePPM: 10.0},
		}
		body, _ := json.Marshal(batch)
		req := httptest.NewRequest("POST", "/v1/measurements", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.PostMeasurements(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})
}
