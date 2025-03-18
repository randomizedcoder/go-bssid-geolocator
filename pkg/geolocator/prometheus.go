package geolocator

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	quantileError    = 0.05
	summaryVecMaxAge = 5 * time.Minute
)

func (g *GeoLocator) InitPromethus() {

	g.pC = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "geolocator",
			Name:      "counts",
			Help:      "geolocator counts",
		},
		[]string{"function", "variable", "type"},
	)

	g.pH = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: "geolocator",
			Name:      "histograms",
			Help:      "geolocator historgrams",
			Objectives: map[float64]float64{
				0.1:  quantileError,
				0.5:  quantileError,
				0.99: quantileError,
			},
			MaxAge: summaryVecMaxAge,
		},
		[]string{"function", "variable", "type"},
	)

	g.pG = promauto.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "geolocator",
			Name:      "gauge",
			Help:      "geolocator network namespace gauge",
		},
	)

	// workers

	g.pwC = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "worker",
			Name:      "counts",
			Help:      "worker counts",
		},
		[]string{"id", "function", "variable", "type"},
	)

	g.pwH = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: "worker",
			Name:      "histograms",
			Help:      "worker historgrams",
			Objectives: map[float64]float64{
				0.1:  quantileError,
				0.5:  quantileError,
				0.99: quantileError,
			},
			MaxAge: summaryVecMaxAge,
		},
		[]string{"id", "function", "variable", "type"},
	)

	g.pwG = promauto.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "worker",
			Name:      "gauge",
			Help:      "worker gauge",
		},
	)

}
