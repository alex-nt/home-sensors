package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"azuremyst.org/go-home-sensors/exporters"
	"azuremyst.org/go-home-sensors/exporters/prometheus"
	"azuremyst.org/go-home-sensors/exporters/sqlite"
	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"

	_ "azuremyst.org/go-home-sensors/sensors/bosch"
	_ "azuremyst.org/go-home-sensors/sensors/plantower"
	_ "azuremyst.org/go-home-sensors/sensors/sensirion"

	"github.com/BurntSushi/toml"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func recordMetrics(interval time.Duration, sens []sensors.Sensor, exps []exporters.Exporter) {
	go func() {
		log.InfoLog.Println("Collecting sensor data")

		for {
			collectedMeasurements := make([]sensors.MeasurementRecording, 0)
			for _, sen := range sens {
				collectedMeasurements = append(collectedMeasurements, sen.Collect()...)
			}

			for _, exp := range exps {
				exp.Export(collectedMeasurements)
			}
			time.Sleep(interval)
		}
	}()
}

func initializeExporters(conf Config) []exporters.Exporter {
	initializeExporters := make([]exporters.Exporter, 0)
	if conf.Exporters.Prometheus.Enable {
		initializeExporters = append(initializeExporters, prometheus.CreateExporter())
	}

	if conf.Exporters.Sqlite.Enable {
		initializeExporters = append(initializeExporters, sqlite.CreateExporter(conf.Exporters.Sqlite.DB))
	}

	return initializeExporters
}

type (
	Config struct {
		Bus       string
		Sensors   map[string]SensorConfig
		Exporters MetricExporters
		Port      int
		Frequency time.Duration
	}

	SensorConfig struct {
		Enable   bool
		Register uint16
	}

	sqliteExporter struct {
		Enable bool
		DB     string
	}

	prometheusExporter struct {
		Enable bool
	}

	MetricExporters struct {
		Prometheus prometheusExporter
		Sqlite     sqliteExporter
	}
)

func main() {
	configLocation := flag.String("config.file", "config.toml", "Configuration in toml format")
	flag.Parse()

	var conf Config
	_, err := toml.DecodeFile(*configLocation, &conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Please provide a valid config file via `--config.file` parameter. Unable to read: %v\n", err)
		os.Exit(1)
	}

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.ErrorLog.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open(conf.Bus)
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	defer b.Close()

	log.InfoLog.Println("Supported sensors:")
	supported := sensors.Supported()
	for _, s := range supported {
		log.InfoLog.Printf("\t%s\n", s)
	}

	initializedExporters := initializeExporters(conf)
	if len(initializedExporters) == 0 {
		log.ErrorLog.Panicln("No exporter was configured!")
	}

	initializedSensors := make([]sensors.Sensor, 0)
	for senName, senConfig := range conf.Sensors {
		if !senConfig.Enable {
			log.InfoLog.Printf("Sensor %s is disabled.\n", senName)
			continue
		}
		sensorInstance := sensors.Sniff(senName)
		if nil == sensorInstance {
			log.ErrorLog.Println("Sensor " + senName + " not supported!")
		} else {
			sensor := *sensorInstance
			sensor.Initialize(b, senConfig.Register)
			initializedSensors = append(initializedSensors, sensor)
		}
	}

	recordMetrics(conf.Frequency, initializedSensors, initializedExporters)

	log.InfoLog.Printf("Started sensor collection service at %d \n", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil)
}
