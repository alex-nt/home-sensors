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

var (
	pmStandardGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.ParticleMatterStandard.ID,
		Help: sensors.ParticleMatterStandard.Description,
	}, []string{string(sensors.ParticleConcentration), string(sensors.SensorName)})
	pmEnvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.ParticleMatterEnvironmental.ID,
		Help: sensors.ParticleMatterEnvironmental.Description,
	}, []string{string(sensors.ParticleConcentration), string(sensors.SensorName)})
	particlesCountGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.ParticleCount.ID,
		Help: sensors.ParticleCount.Description,
	}, []string{string(sensors.ParticleSize), string(sensors.SensorName)})
)

var (
	temperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.Temperature.ID,
		Help: sensors.Temperature.Description,
	}, []string{string(sensors.SensorName)})
	humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.Humidity.ID,
		Help: sensors.Humidity.Description,
	}, []string{string(sensors.SensorName)})
	co2Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.CarbonDioxide.ID,
		Help: sensors.CarbonDioxide.Description,
	}, []string{string(sensors.SensorName)})
	vocGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.VOC.ID,
		Help: sensors.VOC.Description,
	}, []string{string(sensors.SensorName)})
	noxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.NOx.ID,
		Help: sensors.NOx.Description,
	}, []string{string(sensors.SensorName)})
	pressureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.Pressure.ID,
		Help: sensors.Pressure.Description,
	}, []string{string(sensors.SensorName)})
	gasResistanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.GasResistance.ID,
		Help: sensors.GasResistance.Description,
	}, []string{string(sensors.SensorName)})
	iaqGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: sensors.AIQ.ID,
		Help: sensors.AIQ.Description,
	}, []string{string(sensors.SensorName)})
)

var measurementToGauge = map[sensors.Measurement]*prometheus.GaugeVec{
	sensors.Pressure:                    pressureGauge,
	sensors.Temperature:                 temperatureGauge,
	sensors.Humidity:                    humidityGauge,
	sensors.CarbonDioxide:               co2Gauge,
	sensors.GasResistance:               gasResistanceGauge,
	sensors.AIQ:                         iaqGauge,
	sensors.ParticleCount:               particlesCountGauge,
	sensors.ParticleMatterEnvironmental: pmEnvGauge,
	sensors.ParticleMatterStandard:      pmStandardGauge,
	sensors.NOx:                         noxGauge,
	sensors.VOC:                         vocGauge,
}

type PrometheusExporter struct {
}

func CreateExporter() exporters.Exporter {
	http.Handle("/metrics", promhttp.Handler())
	return &PrometheusExporter{}
}

func (pe *PrometheusExporter) Export(recordings []sensors.MeasurementRecording) {
	for _, metricRecording := range recordings {
		gauge := measurementToGauge[*metricRecording.Measure]
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
