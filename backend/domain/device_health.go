package domain

type DeviceHealth struct {
	LastMeasurement *Measurement `json:"last_measurement"`
	RateLastHour    float64      `json:"rate_last_hour"` // medições por minuto na última hora
	IsStale         bool         `json:"stale"`
}
