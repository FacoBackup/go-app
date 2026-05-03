package domain

import "time"

type SeriesData struct {
	Timestamp time.Time `json:"timestamp"`
	ValuePPM  float64   `json:"value_ppm"`
}
