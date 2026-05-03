package domain

import "time"

type Measurement struct {
	DeviceID   string    `json:"device_id"`
	Timestamp  time.Time `json:"timestamp"`
	Gas        GasType   `json:"gas"`
	ValuePPM   float64   `json:"value_ppm"`
	PatientRef string    `json:"patient_ref"`
}
