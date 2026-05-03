package repository

import (
	"healthgo/backend/internal/domain"
	"sync"
	"time"
)

type MeasurementRepository interface {
	SaveBatch(measurements []domain.Measurement) error
	GetSeries(deviceID string, gas domain.GasType, start, end time.Time) ([]domain.Measurement, error)
	GetLastMeasurement(deviceID string) (*domain.Measurement, error)
	GetMeasurementsCount(deviceID string, start, end time.Time) (int, error)
}

type InMemoryRepository struct {
	mu           sync.RWMutex
	measurements []domain.Measurement
	// key: deviceID + timestamp + gas
	dedup map[string]bool
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		measurements: make([]domain.Measurement, 0),
		dedup:        make(map[string]bool),
	}
}

func (r *InMemoryRepository) SaveBatch(measurements []domain.Measurement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range measurements {
		key := m.DeviceID + m.Timestamp.Format(time.RFC3339) + string(m.Gas)
		if r.dedup[key] {
			continue
		}
		r.measurements = append(r.measurements, m)
		r.dedup[key] = true
	}
	return nil
}

func (r *InMemoryRepository) GetSeries(deviceID string, gas domain.GasType, start, end time.Time) ([]domain.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Measurement
	for _, m := range r.measurements {
		if m.DeviceID == deviceID && m.Gas == gas && (m.Timestamp.After(start) || m.Timestamp.Equal(start)) && m.Timestamp.Before(end) {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) GetLastMeasurement(deviceID string) (*domain.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var last *domain.Measurement
	for i := range r.measurements {
		m := &r.measurements[i]
		if m.DeviceID == deviceID {
			if last == nil || m.Timestamp.After(last.Timestamp) {
				last = m
			}
		}
	}
	return last, nil
}

func (r *InMemoryRepository) GetMeasurementsCount(deviceID string, start, end time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, m := range r.measurements {
		if (deviceID == "" || m.DeviceID == deviceID) && (m.Timestamp.After(start) || m.Timestamp.Equal(start)) && m.Timestamp.Before(end) {
			count++
		}
	}
	return count, nil
}
