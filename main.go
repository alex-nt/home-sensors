package main

import (
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

func recordMetrics(scd4x *SCD4X, pmsA0003i *PMSA0003I) {
	go func() {
		for {
			temperatureGauge.Set(scd4x.GetTemperature())
			humidityGauge.Set(scd4x.GetRelativeHumidity())
			co2Gauge.Set(float64(scd4x.GetCO2()))

			pmsA0003i.Read()
			pm10StandardGauge.Set(float64(pmsA0003i.PM10Standard))
			pm25StandardGauge.Set(float64(pmsA0003i.PM25Standard))
			pm100StandardGauge.Set(float64(pmsA0003i.PM100Standard))
			pm10EnvGauge.Set(float64(pmsA0003i.PM10Env))
			pm25EnvGauge.Set(float64(pmsA0003i.PM25Env))
			pm100EnvGauge.Set(float64(pmsA0003i.PM100Env))
			particles03umGauge.Set(float64(pmsA0003i.Particles03um))
			particles05umGauge.Set(float64(pmsA0003i.Particles05um))
			particles10umGauge.Set(float64(pmsA0003i.Particles10um))
			particles25umGauge.Set(float64(pmsA0003i.Particles25um))
			particles50umGauge.Set(float64(pmsA0003i.Particles50um))
			particles100umGauge.Set(float64(pmsA0003i.Particles100um))
			time.Sleep(5 * time.Second)
		}
	}()
}

var (
	pm10StandardGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_pm10_standard",
		Help: "Air quality PM10 Standard",
	})
	pm25StandardGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_pm25_standard",
		Help: "Air quality PM25 Standard",
	})
	pm100StandardGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_pm10_standard",
		Help: "Air quality PM10 Standard",
	})
	pm10EnvGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_pm10_env",
		Help: "Air quality PM10 Environmental",
	})
	pm25EnvGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_pm10_env",
		Help: "Air quality PM25 Environmental",
	})
	pm100EnvGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_pm100_env",
		Help: "Air quality PM100 Environmental",
	})
	particles03umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_particles03um",
		Help: "Air quality Particles 03um",
	})
	particles05umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_particles05um",
		Help: "Air quality Particles 05um",
	})
	particles10umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_particles10um",
		Help: "Air quality Particles 10um",
	})
	particles25umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_particles25um",
		Help: "Air quality Particles 25um",
	})
	particles50umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_particles50um",
		Help: "Air quality Particles 50um",
	})
	particles100umGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_air_quality_particles100um",
		Help: "Air quality Particles 100um",
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

	pmsA0003i := &i2c.Dev{Addr: 0x12, Bus: b}
	pmsA0003iSensor := NewPMSA0003I(pmsA0003i)

	recordMetrics(&co2Sensor, &pmsA0003iSensor)

	InfoLog.Println("Started sensor collection service")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
