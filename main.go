package main

import (
	"flag"
	"net/http"
	"time"

	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Label_Sensor        = "sensor"
	Label_Particle_Size = "particleSize"
)

func recordMetrics(bme68x *sensors.BME68X) {
	go func() {
		log.InfoLog.Println("Collecting sensor data")

		for {
			// temperatureGauge.WithLabelValues("scd41").Set(scd4x.GetTemperature())
			// humidityGauge.WithLabelValues("scd41").Set(scd4x.GetRelativeHumidity())
			// co2Gauge.WithLabelValues("scd41").Set(float64(scd4x.GetCO2()))

			// PMSA003I.Read()
			// pmStandardGauge.WithLabelValues("pmsa003i", "10pm").Set(float64(PMSA003I.PM10Standard))
			// pmStandardGauge.WithLabelValues("pmsa003i", "25pm").Set(float64(PMSA003I.PM25Standard))
			// pmStandardGauge.WithLabelValues("pmsa003i", "100pm").Set(float64(PMSA003I.PM100Standard))

			// pmEnvGauge.WithLabelValues("pmsa003i", "10pm").Set(float64(PMSA003I.PM10Env))
			// pmEnvGauge.WithLabelValues("pmsa003i", "25pm").Set(float64(PMSA003I.PM25Env))
			// pmEnvGauge.WithLabelValues("pmsa003i", "100pm").Set(float64(PMSA003I.PM100Env))

			// particlesCountGauge.WithLabelValues("pmsa003i", "03um").Set(float64(PMSA003I.Particles03um))
			// particlesCountGauge.WithLabelValues("pmsa003i", "05pm").Set(float64(PMSA003I.Particles05um))
			// particlesCountGauge.WithLabelValues("pmsa003i", "10pm").Set(float64(PMSA003I.Particles10um))
			// particlesCountGauge.WithLabelValues("pmsa003i", "25pm").Set(float64(PMSA003I.Particles25um))
			// particlesCountGauge.WithLabelValues("pmsa003i", "50pm").Set(float64(PMSA003I.Particles50um))
			// particlesCountGauge.WithLabelValues("pmsa003i", "100pm").Set(float64(PMSA003I.Particles100um))

			bme68x.GetSensorData()
			temperatureGauge.WithLabelValues("bme68x").Set(float64(bme68x.Data.Temperature))
			humidityGauge.WithLabelValues("bme68x").Set(float64(bme68x.Data.Humidity))
			pressureGauge.WithLabelValues("bme68x").Set(float64(bme68x.Data.Pressure))
			gasResistanceGauge.WithLabelValues("bme68x").Set(float64(bme68x.Data.GasResistance))
			iaqGauge.WithLabelValues("bme68x").Set(float64(bme68x.Data.IAQ))
			time.Sleep(5 * time.Second)
		}
	}()
}

var (
	pmStandardGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_pm_concentration_standard",
		Help: "Air quality. PM concentration in standard units",
	}, []string{Label_Sensor, Label_Particle_Size})
	pmEnvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_pm_concentration_env",
		Help: "Air quality. PM concentration in environmental units",
	}, []string{Label_Sensor, Label_Particle_Size})
	particlesCountGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_air_quality_particles_count",
		Help: "Air quality. Particulate matter per 0.1L air.",
	}, []string{Label_Sensor, Label_Particle_Size})
)

var (
	temperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_temperature",
		Help: "Ambient temperature in C",
	}, []string{Label_Sensor})
	humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_humidity",
		Help: "Ambient relative humidity",
	}, []string{Label_Sensor})
	co2Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_co2",
		Help: "CO2 in ppm",
	}, []string{Label_Sensor})
	pressureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_pressure",
		Help: "Pressure Hpa",
	}, []string{Label_Sensor})
	gasResistanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_gasResistance",
		Help: "Gas resistance in Ohm",
	}, []string{Label_Sensor})
	iaqGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_iaq",
		Help: "Indoor Air Quality",
	}, []string{Label_Sensor})
)

func main() {
	listenAddress := flag.String("web.listen-address", ":2112", "Define the address and port at which to listen")
	flag.Parse()

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.ErrorLog.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	defer b.Close()

	// Dev is a valid conn.Conn.
	// co2Sensor := sensors.NewSCD4X(&i2c.Dev{Addr: 0x62, Bus: b})
	// co2Sensor.StartPeriodicMeasurement()

	// PMSA003ISensor := sensors.NewPMSA003I(&i2c.Dev{Addr: 0x12, Bus: b})

	bme68x := sensors.NewBME68X(&i2c.Dev{Addr: 0x76, Bus: b})
	bme68x.Init()

	recordMetrics(&bme68x)

	log.InfoLog.Printf("Started sensor collection service at %s \n", *listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*listenAddress, nil)
}
