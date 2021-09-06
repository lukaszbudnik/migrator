package metrics

import (
	"github.com/Depado/ginprom"
)

// Metrics interface abstracts all metrics operations performed by migrator
type Metrics interface {
	SetGaugeValue(name string, labelValues []string, value float64) error
	AddGaugeValue(name string, labelValues []string, value float64) error
	IncrementGaugeValue(name string, labelValues []string) error
}

// New returns new instance of Metrics, currently Prometheus is available
func New(prometheus *ginprom.Prometheus) Metrics {
	return &prometheusMetrics{prometheus}
}

// prometheusMetrics is struct for implementing Prometheus metrics
type prometheusMetrics struct {
	prometheus *ginprom.Prometheus
}

// SetGaugeValue sets guage to a value
func (m *prometheusMetrics) SetGaugeValue(name string, labelValues []string, value float64) error {
	return m.prometheus.SetGaugeValue(name, labelValues, value)
}

// AddGaugeValue adds value to guage
func (m *prometheusMetrics) AddGaugeValue(name string, labelValues []string, value float64) error {
	return m.prometheus.AddGaugeValue(name, labelValues, value)
}

// IncrementGaugeValue increments guage
func (m *prometheusMetrics) IncrementGaugeValue(name string, labelValues []string) error {
	return m.prometheus.IncrementGaugeValue(name, labelValues)
}
