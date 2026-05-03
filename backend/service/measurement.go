package service

import (
	"cmp"
	"healthgo/backend/domain"
	"healthgo/backend/repository"
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

	groups := s.groupByMinute(measurements)
	result := s.calculateAverages(groups)
	s.sortByTimestamp(result)

	return result, nil
}

func (s *MeasurementService) groupByMinute(measurements []domain.Measurement) map[time.Time][]float64 {
	groups := make(map[time.Time][]float64)
	for _, m := range measurements {
		minute := m.Timestamp.Truncate(time.Minute)
		groups[minute] = append(groups[minute], m.ValuePPM)
	}
	return groups
}

func (s *MeasurementService) calculateAverages(groups map[time.Time][]float64) []domain.SeriesData {
	result := make([]domain.SeriesData, 0, len(groups))
	for minute, values := range groups {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		result = append(result, domain.SeriesData{
			Timestamp: minute,
			ValuePPM:  sum / float64(len(values)),
		})
	}
	return result
}

func (s *MeasurementService) sortByTimestamp(data []domain.SeriesData) {
	slices.SortFunc(data, func(a, b domain.SeriesData) int {
		return cmp.Compare(a.Timestamp.UnixNano(), b.Timestamp.UnixNano())
	})
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
		health.IsStale = time.Since(last.Timestamp) > 5*time.Minute
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
