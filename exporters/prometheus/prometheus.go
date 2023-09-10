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
		Name: "room_air_quality_pm_concentration_standard",
		Help: "Air quality. PM concentration in standard units",
	}, []string{string(sensors.ParticleConcentration), string(sensors.SensorName)})
	pmEnvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_pm_concentration_env",
		Help: "Air quality. PM concentration in environmental units",
	}, []string{string(sensors.ParticleConcentration), string(sensors.SensorName)})
	particlesCountGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_particles_count",
		Help: "Air quality. Particulate matter per 0.1L air.",
	}, []string{string(sensors.ParticleSize), string(sensors.SensorName)})
)

var (
	temperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_temperature",
		Help: "Ambient temperature in C",
	}, []string{string(sensors.SensorName)})
	humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_humidity",
		Help: "Ambient relative humidity",
	}, []string{string(sensors.SensorName)})
	co2Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_co2",
		Help: "CO2 in ppm",
	}, []string{string(sensors.SensorName)})
	pressureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_pressure",
		Help: "Pressure Hpa",
	}, []string{string(sensors.SensorName)})
	gasResistanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_gasResistance",
		Help: "Gas resistance in Ohm",
	}, []string{string(sensors.SensorName)})
	iaqGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_iaq",
		Help: "Indoor Air Quality",
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
		gauge.With(extendedLabels).Add(metricRecording.Value)
	}
}
