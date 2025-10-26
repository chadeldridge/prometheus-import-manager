package router

import (
	"net/http"
	"time"

	"github.com/chadeldridge/prometheus-import-manager/core"
)

var metrics *Metrics

type Metrics struct {
	Start    time.Time
	Requests int
	Errors   int
	Duration []time.Duration
}

type MetricsReport struct {
	Start       time.Time `json:"start"`        // Time metrics started being gathered
	End         time.Time `json:"end"`          // Time metrics were gathered
	DurationSec float64   `json:"duration_sec"` // Duration metrics where gathered in Seconds
	Requests    int       `json:"requests"`     // Total Requests
	Errors      int       `json:"errors"`       // Total Errors
	RequestsPS  float64   `json:"requests_ps"`  // Requests Per Second
	ErrorsPS    float64   `json:"errors_ps"`    // Errors Per Second
	MinDuration float64   `json:"min_duration"` // Minimum Duration
	AvgDuration float64   `json:"avg_duration"` // Average Duration
	MaxDuration float64   `json:"max_duration"` // Maximum Duration
}

func init() {
	metrics = NewMetrics()
}

func NewMetrics() *Metrics {
	return &Metrics{
		Start:    time.Now(),
		Requests: 0,
		Errors:   0,
		Duration: []time.Duration{},
	}
}

func RecordRequest(code int, d time.Duration) {
	metrics.Requests++
	if code >= 400 {
		metrics.Errors++
	}
	metrics.Duration = append(metrics.Duration, d)
}

func Report() MetricsReport {
	end := time.Now()

	// Copy the metrics and reset the metrics.
	m := *metrics
	metrics = NewMetrics()

	// Calculate the duration since metrics started being collected.
	d := end.Sub(m.Start)

	// Calculate the min, avg, and max durations.
	min, avg, max := getDuration(m.Duration)

	rps := 0.0
	if m.Requests != 0 && d.Seconds() != 0 {
		rps = float64(m.Requests) / d.Seconds()
	}

	eps := 0.0
	if m.Errors != 0 && d.Seconds() != 0 {
		eps = float64(m.Errors) / d.Seconds()
	}

	return MetricsReport{
		Start:       m.Start,
		End:         end,
		DurationSec: d.Seconds(),
		Requests:    m.Requests,
		Errors:      m.Errors,
		RequestsPS:  rps,
		ErrorsPS:    eps,
		MinDuration: min,
		AvgDuration: avg,
		MaxDuration: max,
	}
}

func getDuration(durations []time.Duration) (float64, float64, float64) {
	count := len(durations)
	if count == 0 {
		return 0, 0, 0
	}

	var total time.Duration
	min := time.Duration(time.Hour)
	max := time.Duration(0)

	for _, d := range durations {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	avg := total / time.Duration(count)
	return min.Seconds(), avg.Seconds(), max.Seconds()
}

func HandleMetrics(logger *core.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			report := Report()
			err := RenderJSON(w, http.StatusOK, report)
			if err != nil {
				logger.Printf("handleMetrics: %v\n", err)
			}
		})
}
