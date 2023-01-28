package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

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

var (
	ErrorLog *log.Logger
	InfoLog  *log.Logger
)

func init() {
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func recordMetrics(scd4x *SCD4X, PMSA003I *PMSA003I) {
	go func() {
		for {
			temperatureGauge.WithLabelValues("scd41").Set(scd4x.GetTemperature())
			humidityGauge.WithLabelValues("scd41").Set(scd4x.GetRelativeHumidity())
			co2Gauge.WithLabelValues("scd41").Set(float64(scd4x.GetCO2()))

			PMSA003I.Read()
			pmStandardGauge.WithLabelValues("pmsa003i", "10pm").Set(float64(PMSA003I.PM10Standard))
			pmStandardGauge.WithLabelValues("pmsa003i", "25pm").Set(float64(PMSA003I.PM25Standard))
			pmStandardGauge.WithLabelValues("pmsa003i", "100pm").Set(float64(PMSA003I.PM100Standard))

			pmEnvGauge.WithLabelValues("pmsa003i", "10pm").Set(float64(PMSA003I.PM10Env))
			pmEnvGauge.WithLabelValues("pmsa003i", "25pm").Set(float64(PMSA003I.PM25Env))
			pmEnvGauge.WithLabelValues("pmsa003i", "100pm").Set(float64(PMSA003I.PM100Env))

			particlesCountGauge.WithLabelValues("pmsa003i", "03um").Set(float64(PMSA003I.Particles03um))
			particlesCountGauge.WithLabelValues("pmsa003i", "05pm").Set(float64(PMSA003I.Particles05um))
			particlesCountGauge.WithLabelValues("pmsa003i", "10pm").Set(float64(PMSA003I.Particles10um))
			particlesCountGauge.WithLabelValues("pmsa003i", "25pm").Set(float64(PMSA003I.Particles25um))
			particlesCountGauge.WithLabelValues("pmsa003i", "50pm").Set(float64(PMSA003I.Particles50um))
			particlesCountGauge.WithLabelValues("pmsa003i", "100pm").Set(float64(PMSA003I.Particles100um))
			time.Sleep(15 * time.Second)
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
)

func main() {
	listenAddress := flag.String("web.listen-address", ":2112", "Define the address and port at which to listen")
	flag.Parse()

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Dev is a valid conn.Conn.
	scd41 := &i2c.Dev{Addr: 0x62, Bus: b}
	co2Sensor := NewSCD4X(scd41)
	co2Sensor.StartPeriodicMeasurement()

	PMSA003I := &i2c.Dev{Addr: 0x12, Bus: b}
	PMSA003ISensor := NewPMSA003I(PMSA003I)

	recordMetrics(&co2Sensor, &PMSA003ISensor)

	InfoLog.Printf("Started sensor collection service at %s \n", *listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*listenAddress, nil)
}
