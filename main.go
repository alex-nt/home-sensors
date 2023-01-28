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
			temperatureGauge.Set(scd4x.GetTemperature())
			humidityGauge.Set(scd4x.GetRelativeHumidity())
			co2Gauge.Set(float64(scd4x.GetCO2()))

			PMSA003I.Read()
			pm10StandardGauge.Set(float64(PMSA003I.PM10Standard))
			pm25StandardGauge.Set(float64(PMSA003I.PM25Standard))
			pm100StandardGauge.Set(float64(PMSA003I.PM100Standard))
			pm10EnvGauge.Set(float64(PMSA003I.PM10Env))
			pm25EnvGauge.Set(float64(PMSA003I.PM25Env))
			pm100EnvGauge.Set(float64(PMSA003I.PM100Env))
			particles03umGauge.Set(float64(PMSA003I.Particles03um))
			particles05umGauge.Set(float64(PMSA003I.Particles05um))
			particles10umGauge.Set(float64(PMSA003I.Particles10um))
			particles25umGauge.Set(float64(PMSA003I.Particles25um))
			particles50umGauge.Set(float64(PMSA003I.Particles50um))
			particles100umGauge.Set(float64(PMSA003I.Particles100um))
			time.Sleep(5 * time.Second)
		}
	}()
}

var (
	pm10StandardGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_standard_pm10",
		Help:        "Air quality PM10 Standard",
		ConstLabels: map[string]string{"diameter": "10pm"},
	})
	pm25StandardGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_standard_pm25",
		Help:        "Air quality PM25 Standard",
		ConstLabels: map[string]string{"diameter": "25pm"},
	})
	pm100StandardGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_standard_pm100",
		Help:        "Air quality PM100 Standard",
		ConstLabels: map[string]string{"diameter": "100pm"},
	})
	pm10EnvGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_env_pm10",
		Help:        "Air quality PM10 Environmental",
		ConstLabels: map[string]string{"diameter": "10pm"},
	})
	pm25EnvGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_env_pm25",
		Help:        "Air quality PM25 Environmental",
		ConstLabels: map[string]string{"diameter": "25pm"},
	})
	pm100EnvGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_env_pm100",
		Help:        "Air quality PM100 Environmental",
		ConstLabels: map[string]string{"diameter": "100pm"},
	})
	particles03umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_particles_03um",
		Help:        "Air quality Particles 03um",
		ConstLabels: map[string]string{"size": "03um"},
	})
	particles05umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_particles_05um",
		Help:        "Air quality Particles 05um",
		ConstLabels: map[string]string{"size": "05um"},
	})
	particles10umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_particles_10um",
		Help:        "Air quality Particles 10um",
		ConstLabels: map[string]string{"size": "10um"},
	})
	particles25umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_particles_25um",
		Help:        "Air quality Particles 25um",
		ConstLabels: map[string]string{"size": "25pm"},
	})
	particles50umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_particles_50um",
		Help:        "Air quality Particles 50um",
		ConstLabels: map[string]string{"size": "50um"},
	})
	particles100umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "room_air_quality_particles_100um",
		Help:        "Air quality Particles 100um",
		ConstLabels: map[string]string{"size": "100um"},
	})
)

var (
	temperatureGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_temperature",
		Help: "Ambient temperature in C",
	})
	humidityGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_humidity",
		Help: "Ambient relative humidity",
	})
	co2Gauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_co2",
		Help: "CO2 in ppm",
	})
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

	InfoLog.Printf("Started sensor collection service at %v \n", listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*listenAddress, nil)
}
