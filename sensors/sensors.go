package sensors

import (
	"sync"
	"sync/atomic"

	"periph.io/x/conn/v3/i2c"
)

type Sensor interface {
	Initialize(bus i2c.Bus, addr uint16)
	Name() string
	Family(name string) bool
	Collect() []MeasurementRecording
}

// Sensors is the list of supported sensors.
var (
	sensorsMu     sync.Mutex
	atomicSensors atomic.Value
)

func RegisterSensor(sensor Sensor) {
	sensorsMu.Lock()
	sensors, _ := atomicSensors.Load().([]Sensor)
	atomicSensors.Store(append(sensors, sensor))
	sensorsMu.Unlock()
}

func Sniff(family string) *Sensor {
	sensors, _ := atomicSensors.Load().([]Sensor)
	for _, s := range sensors {
		if s.Family(family) {
			return &s
		}
	}
	return nil
}

func Supported() []string {
	sensorsMu.Lock()
	sensors, _ := atomicSensors.Load().([]Sensor)
	supportedList := make([]string, len(sensors))
	for _, sensor := range sensors {
		supportedList = append(supportedList, sensor.Name())
	}

	sensorsMu.Unlock()
	return supportedList
}
