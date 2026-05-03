package repository

import (
	"healthgo/backend/domain"
	"time"
)

type MeasurementRepository interface {
	SaveBatch(measurements []domain.Measurement) error
	GetSeries(deviceID string, gas domain.GasType, start, end time.Time) ([]domain.Measurement, error)
	GetLastMeasurement(deviceID string) (*domain.Measurement, error)
	GetMeasurementsCount(deviceID string, start, end time.Time) (int, error)
	// Estrutura adicional para operações no frontend
	GetDevices() ([]string, error)
	// Estrutura adicional para operações no frontend
	DeleteDevice(deviceID string) error
}
