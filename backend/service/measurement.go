package service

import (
	"healthgo/backend/domain"
	"healthgo/backend/repository"
	"maps"
	"slices"
	"time"
)

type MeasurementService struct {
	repo repository.MeasurementRepository
}

func NewMeasurementService(repo repository.MeasurementRepository) *MeasurementService {
	return &MeasurementService{repo: repo}
}

func (s *MeasurementService) AddMeasurements(measurements []domain.Measurement) error {
	return s.repo.SaveBatch(measurements)
}

func (s *MeasurementService) GetSeries(deviceID string, gas domain.GasType, start, end time.Time) ([]domain.SeriesData, error) {
	measurements, err := s.repo.GetSeries(deviceID, gas, start, end)
	if err != nil {
		return nil, err
	}

	// Agrupar por minuto
	groups := make(map[time.Time][]float64)
	for _, m := range measurements {
		minute := m.Timestamp.Truncate(time.Minute)
		groups[minute] = append(groups[minute], m.ValuePPM)
	}

	result := make([]domain.SeriesData, 0, len(groups))
	for minute := range maps.Keys(groups) {
		values := groups[minute]
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		result = append(result, domain.SeriesData{
			Timestamp: minute,
			ValuePPM:  sum / float64(len(values)),
		})
	}

	// Ordenar por timestamp
	slices.SortFunc(result, func(a, b domain.SeriesData) int {
		if a.Timestamp.Before(b.Timestamp) {
			return -1
		}
		if a.Timestamp.After(b.Timestamp) {
			return 1
		}
		return 0
	})

	return result, nil
}

func (s *MeasurementService) GetDeviceHealth(deviceID string) (*domain.DeviceHealth, error) {
	last, err := s.repo.GetLastMeasurement(deviceID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	hourAgo := now.Add(-time.Hour)
	count, err := s.repo.GetMeasurementsCount(deviceID, hourAgo, now)
	if err != nil {
		return nil, err
	}

	health := &domain.DeviceHealth{
		LastMeasurement: last,
		RateLastHour:    float64(count) / 60.0,
		IsStale:         false,
	}

	if last != nil {
		health.IsStale = now.Sub(last.Timestamp) > 5*time.Minute
	} else {
		health.IsStale = true
	}

	return health, nil
}

// Estrutura adicional para operações no frontend
func (s *MeasurementService) GetDevices() ([]string, error) {
	return s.repo.GetDevices()
}

// Estrutura adicional para operações no frontend
func (s *MeasurementService) DeleteDevice(deviceID string) error {
	return s.repo.DeleteDevice(deviceID)
}
