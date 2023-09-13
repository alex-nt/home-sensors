package prometheus

import (
	"net/http"

	"azuremyst.org/go-home-sensors/exporters"
	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusExporter struct {
	gauges map[string]*prometheus.GaugeVec
}

func CreateExporter() exporters.Exporter {
	gauges := make(map[string]*prometheus.GaugeVec, 0)
	for _, measurement := range sensors.Measurements {
		gauges[measurement.ID] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: measurement.ID,
			Help: measurement.Description,
		}, measurement.Labels)
	}
	http.Handle("/metrics", promhttp.Handler())
	return &PrometheusExporter{gauges: gauges}
}

func (pe *PrometheusExporter) Export(recordings []sensors.MeasurementRecording) {
	for _, metricRecording := range recordings {
		gauge := pe.gauges[metricRecording.Measure.ID]
		if nil == gauge {
			log.ErrorLog.Printf("No gauge found for metric %s[%s %s]\n",
				metricRecording.Measure.ID, metricRecording.Measure.Unit, metricRecording.Measure.Description)
			continue
		}

		extendedLabels := make(map[string]string, len(metricRecording.Metadata)+1)
		extendedLabels[string(sensors.SensorName)] = metricRecording.Sensor
		for k, v := range metricRecording.Metadata {
			extendedLabels[string(k)] = v
		}
		gauge.With(extendedLabels).Set(metricRecording.Value)
	}
}
