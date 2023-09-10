package plantower

import (
	"encoding/binary"
	"fmt"
	"strings"

	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"
	"periph.io/x/conn/v3/i2c"
)

type PMSA003I struct {
	device         *i2c.Dev
	PM1Standard    uint16
	PM2_5Standard  uint16
	PM10Standard   uint16
	PM1Env         uint16
	PM2_5Env       uint16
	PM10Env        uint16
	Particles0_3um uint16
	Particles0_5um uint16
	Particles1um   uint16
	Particles2_5um uint16
	Particles5um   uint16
	Particles10um  uint16
}

func init() {
	sensors.RegisterSensor(&PMSA003I{})
}

func (pmsa *PMSA003I) Initialize(bus i2c.Bus, addr uint16) {
	pmsa.device = &i2c.Dev{Addr: addr, Bus: bus}
}

func (pmsa *PMSA003I) Name() string {
	return "pmsa003i"
}

func (pmsa *PMSA003I) Family(name string) bool {
	return len(name) == 8 && strings.EqualFold(pmsa.Name(), name)
}

func (pmsa *PMSA003I) Collect() []sensors.MeasurementRecording {
	pmsa.read()
	measurements := make([]sensors.MeasurementRecording, 0)
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterStandard,
		Value:    float64(pmsa.PM1Standard),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleConcentration": "1.0pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterStandard,
		Value:    float64(pmsa.PM2_5Standard),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleConcentration": "2.5pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterStandard,
		Value:    float64(pmsa.PM10Standard),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleConcentration": "10pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    float64(pmsa.PM1Env),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleConcentration": "1.0pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    float64(pmsa.PM2_5Env),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleConcentration": "2.5pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    float64(pmsa.PM10Env),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleConcentration": "10pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleCount,
		Value:    float64(pmsa.Particles0_3um),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleSize": "0.3um"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleCount,
		Value:    float64(pmsa.Particles0_5um),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleSize": "0.5um"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleCount,
		Value:    float64(pmsa.Particles1um),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleSize": "1um"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleCount,
		Value:    float64(pmsa.Particles2_5um),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleSize": "2.5um"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleCount,
		Value:    float64(pmsa.Particles5um),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleSize": "5.0um"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleCount,
		Value:    float64(pmsa.Particles10um),
		Sensor:   pmsa.Name(),
		Metadata: map[string]string{"particleSize": "10um"},
	})
	return measurements
}

func (pmsa *PMSA003I) read() {
	response := make([]byte, 32)
	var command []byte
	if err := pmsa.device.Tx(command, response); err != nil {
		log.ErrorLog.Printf("error while reading from device")
		return
	}

	frameLength := binary.BigEndian.Uint16(response[2:4])
	if frameLength != 28 {
		log.ErrorLog.Printf("invalid PM2.5 frame length")
		return
	}
	if err := pmsa.crc(response); err != nil {
		log.ErrorLog.Printf("crc check failed %v", err)
		return
	}

	pmsa.PM1Standard = binary.BigEndian.Uint16(response[4:6])
	pmsa.PM2_5Standard = binary.BigEndian.Uint16(response[6:8])
	pmsa.PM10Standard = binary.BigEndian.Uint16(response[8:10])
	pmsa.PM1Env = binary.BigEndian.Uint16(response[10:12])
	pmsa.PM2_5Env = binary.BigEndian.Uint16(response[12:14])
	pmsa.PM10Env = binary.BigEndian.Uint16(response[14:16])
	pmsa.Particles0_3um = binary.BigEndian.Uint16(response[16:18])
	pmsa.Particles0_5um = binary.BigEndian.Uint16(response[18:20])
	pmsa.Particles1um = binary.BigEndian.Uint16(response[20:22])
	pmsa.Particles2_5um = binary.BigEndian.Uint16(response[22:24])
	pmsa.Particles5um = binary.BigEndian.Uint16(response[24:26])
	pmsa.Particles10um = binary.BigEndian.Uint16(response[26:28])
}

func (pmsa *PMSA003I) crc(data []byte) error {
	checksum := binary.BigEndian.Uint16(data[30:32])
	check := sumUINT8(data[0:30])
	if check != checksum {
		return fmt.Errorf("invalid PM2.5 checksum")
	}
	return nil
}

func sumUINT8(array []uint8) uint16 {
	var result uint16
	for _, v := range array {
		result += uint16(v)
	}
	return result
}
