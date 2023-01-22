package main

import (
	"log"
	"net/http"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics(scd4x *SCD4X) {
	go func() {
		for {
			temperatureGauge.Set(scd4x.GetTemperature())
			humidityGauge.Set(scd4x.GetRelativeHumidity())
			co2Gauge.Set(float64(scd4x.GetCO2()))
			time.Sleep(5 * time.Second)
		}
	}()
}

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
	d := &i2c.Dev{Addr: 0x62, Bus: b}
	co2Sensor := NewSCD4X(d)
	co2Sensor.StartPeriodicMeasurement()

	recordMetrics(&co2Sensor)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
