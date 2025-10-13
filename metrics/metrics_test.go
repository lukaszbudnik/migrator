package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newMetrics(r *gin.Engine) Metrics {
	p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Namespace("migrator"),
		ginprom.Subsystem("gin"),
		ginprom.Path("/metrics"),
	)

	p.AddCustomGauge("gauge", "Test guage", []string{"type"})

	r.Use(p.Instrument())

	metrics := New(p)

	return metrics
}

func TestMetrics(t *testing.T) {
	r := gin.New()

	metrics := newMetrics(r)
	metrics.SetGaugeValue("gauge", []string{"first"}, 1)
	metrics.SetGaugeValue("gauge", []string{"second"}, 1)
	metrics.AddGaugeValue("gauge", []string{"first"}, 1)
	metrics.IncrementGaugeValue("gauge", []string{"second"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8; escaping=underscores", w.Result().Header.Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `migrator_gin_gauge{type="first"} 2`)
	assert.Contains(t, w.Body.String(), `migrator_gin_gauge{type="second"} 2`)
}
