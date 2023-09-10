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

const (
	Label_Sensor                 = "sensor"
	Label_Particle_Size          = "particleSize"
	Label_Particle_Concentration = "particleConcentration"
)

var (
	pmStandardGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_pm_concentration_standard",
		Help: "Air quality. PM concentration in standard units",
	}, []string{Label_Sensor})
	pmEnvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_pm_concentration_env",
		Help: "Air quality. PM concentration in environmental units",
	}, []string{Label_Sensor})
	particlesCountGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_particles_count",
		Help: "Air quality. Particulate matter per 0.1L air.",
	}, []string{})
)

var (
	temperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_temperature",
		Help: "Ambient temperature in C",
	}, []string{})
	humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_humidity",
		Help: "Ambient relative humidity",
	}, []string{})
	co2Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_co2",
		Help: "CO2 in ppm",
	}, []string{})
	pressureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_pressure",
		Help: "Pressure Hpa",
	}, []string{})
	gasResistanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_gasResistance",
		Help: "Gas resistance in Ohm",
	}, []string{})
	iaqGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_iaq",
		Help: "Indoor Air Quality",
	}, []string{})
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
		extendedLabels[Label_Sensor] = metricRecording.Sensor
		for k, v := range metricRecording.Metadata {
			extendedLabels[k] = v
		}
		gauge.With(extendedLabels).Add(metricRecording.Value)
	}
}
