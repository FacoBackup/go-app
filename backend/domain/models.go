package domain

import "time"

type GasType string

const (
	GasH2  GasType = "H2"
	GasCH4 GasType = "CH4"
	GasH2S GasType = "H2S"
)

type Measurement struct {
	DeviceID   string    `json:"device_id"`
	Timestamp  time.Time `json:"timestamp"`
	Gas        GasType   `json:"gas"`
	ValuePPM   float64   `json:"value_ppm"`
	PatientRef string    `json:"patient_ref"`
}

type SeriesData struct {
	Timestamp time.Time `json:"timestamp"`
	ValuePPM  float64   `json:"value_ppm"`
}

type DeviceHealth struct {
	LastMeasurement *Measurement `json:"last_measurement"`
	RateLastHour    float64      `json:"rate_last_hour"` // medições por minuto na última hora
	IsStale         bool         `json:"stale"`
}
