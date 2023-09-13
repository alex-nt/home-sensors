package sensirion

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
	"time"

	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"
	"periph.io/x/conn/v3/i2c"
)

var (
	SEN5X_RESET             = Command{code: 0xD304, description: "Reset device", delay: time.Duration(100 * time.Millisecond), size: 0}
	SEN5X_SERIAL_NUMBER     = Command{code: 0xD033, description: "Serial number", delay: time.Duration(20 * time.Millisecond), size: 32}
	SEN5X_PRODUCT_NAME      = Command{code: 0xD014, description: "Product name", delay: time.Duration(20 * time.Millisecond), size: 32}
	SEN5X_VERSION           = Command{code: 0xD100, description: "Versions", delay: time.Duration(20 * time.Millisecond), size: 8}
	SEN5X_READ_STATUS       = Command{code: 0xD206, description: "Read status", delay: time.Duration(20 * time.Millisecond), size: 4}
	SEN5X_START_MEASUREMENT = Command{code: 0x0021, description: "Start measurement", delay: time.Duration(50 * time.Millisecond), size: 0}
	SEN5X_READ_MEASUREMENTS = Command{code: 0x03C4, description: "Read measurements", delay: time.Duration(20 * time.Millisecond), size: 16}
)

type SEN5XDeviceInfo struct {
	serialNumber string
	productName  string

	firmwareMajorVersion uint8
	firmwareMinorVersion uint8
	firmwareDebug        bool
	hardwareMajorVersion uint8
	hardwareMinorVersion uint8
	protocolMajorVersion uint8
	protocolMinorVersion uint8

	status uint32
}

type SEN5XMeasurement struct {
	VOCIndex    float64
	NOxIndex    float64
	Humidity    float64
	Temperature float64
	PM1_0       float64
	PM2_5       float64
	PM4_0       float64
	PM10        float64
}

type SEN5X struct {
	device *i2c.Dev
	mu     sync.Mutex

	deviceInfo SEN5XDeviceInfo
	data       SEN5XMeasurement
}

func init() {
	sensors.RegisterSensor(&SEN5X{})
}

func (sen5x *SEN5X) Name() string {
	return "sen5x"
}

func (sen5x *SEN5X) Family(name string) bool {
	return len(name) == 5 && strings.HasPrefix(strings.ToLower(name), "sen5")
}

func (sen5x *SEN5X) Collect() []sensors.MeasurementRecording {
	measurements := make([]sensors.MeasurementRecording, 0)

	data, err := SEN5X_READ_MEASUREMENTS.Read(sen5x.device, &sen5x.mu)
	if err != nil {
		log.ErrorLog.Printf("Failed to read measurements: %q", err)
		return measurements
	}

	sen5x.data.PM1_0 = float64(binary.BigEndian.Uint16(data[0:2])) / 10
	sen5x.data.PM2_5 = float64(binary.BigEndian.Uint16(data[2:4])) / 10
	sen5x.data.PM4_0 = float64(binary.BigEndian.Uint16(data[4:6])) / 10
	sen5x.data.PM10 = float64(binary.BigEndian.Uint16(data[6:8])) / 10
	sen5x.data.Humidity = float64(int16(binary.BigEndian.Uint16(data[8:10]))) / 200
	sen5x.data.Temperature = float64(int16(binary.BigEndian.Uint16(data[10:12]))) / 100
	sen5x.data.VOCIndex = float64(int16(binary.BigEndian.Uint16(data[12:14]))) / 10
	sen5x.data.NOxIndex = float64(int16(binary.BigEndian.Uint16(data[14:16]))) / 10

	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.Humidity,
		Value:   sen5x.data.Humidity,
		Sensor:  sen5x.deviceInfo.productName,
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.Temperature,
		Value:   sen5x.data.Temperature,
		Sensor:  sen5x.deviceInfo.productName,
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.NOx,
		Value:   sen5x.data.NOxIndex,
		Sensor:  sen5x.deviceInfo.productName,
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.VOC,
		Value:   sen5x.data.VOCIndex,
		Sensor:  sen5x.deviceInfo.productName,
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    sen5x.data.PM1_0,
		Sensor:   sen5x.deviceInfo.productName,
		Metadata: map[sensors.Metadata]string{sensors.ParticleConcentration: "1.0pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    sen5x.data.PM2_5,
		Sensor:   sen5x.deviceInfo.productName,
		Metadata: map[sensors.Metadata]string{sensors.ParticleConcentration: "2.5pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    sen5x.data.PM4_0,
		Sensor:   sen5x.deviceInfo.productName,
		Metadata: map[sensors.Metadata]string{sensors.ParticleConcentration: "4.0pm"},
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure:  &sensors.ParticleMatterEnvironmental,
		Value:    sen5x.data.PM10,
		Sensor:   sen5x.deviceInfo.productName,
		Metadata: map[sensors.Metadata]string{sensors.ParticleConcentration: "10pm"},
	})
	return measurements
}

func (sen5x *SEN5X) Initialize(bus i2c.Bus, addr uint16) {
	sen5x.device = &i2c.Dev{Addr: addr, Bus: bus}
	if err := sen5x.Reset(); err != nil {
		log.ErrorLog.Printf("Failed to reset device: %q", err)
		return
	}
	if err := sen5x.Versions(); err != nil {
		log.ErrorLog.Printf("Failed to retrieve device versions: %q", err)
		return
	}
	if err := sen5x.ProductName(); err != nil {
		log.ErrorLog.Printf("Failed to retrieve product name: %q", err)
		return
	}
	if err := sen5x.SerialNumber(); err != nil {
		log.ErrorLog.Printf("Failed to retrieve serial number: %q", err)
		return
	}
	if err := sen5x.SerialNumber(); err != nil {
		log.ErrorLog.Printf("Failed to retrieve serial number: %q", err)
		return
	}
	if err := sen5x.Status(); err != nil {
		log.ErrorLog.Printf("Failed to retrieve status: %q", err)
		return
	}
	log.InfoLog.Printf(`Sensirion SEN5x
	ProductName: %s
	SerialNumber: %s
	Status: %d
	FirmwareDebug: %t
	Versions:
		Firmware: %d.%d
		Hardware: %d.%d
		Protocol: %d.%d`,
		sen5x.deviceInfo.productName, sen5x.deviceInfo.serialNumber,
		sen5x.deviceInfo.status, sen5x.deviceInfo.firmwareDebug,
		sen5x.deviceInfo.firmwareMajorVersion, sen5x.deviceInfo.firmwareMinorVersion,
		sen5x.deviceInfo.hardwareMajorVersion, sen5x.deviceInfo.hardwareMinorVersion,
		sen5x.deviceInfo.protocolMajorVersion, sen5x.deviceInfo.protocolMinorVersion)

	if err := SEN5X_START_MEASUREMENT.Write(sen5x.device, &sen5x.mu); err != nil {
		log.ErrorLog.Printf("Failed to start measurements: %q", err)
		return
	}
}

func (sen5x *SEN5X) Reset() error {
	return SEN5X_RESET.Write(sen5x.device, &sen5x.mu)
}

func (sen5x *SEN5X) SerialNumber() error {
	err := SEN5X_SERIAL_NUMBER.Write(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to write serial number command: %q", err)
	}
	data, err := SEN5X_SERIAL_NUMBER.Read(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to read serial number: %q", err)
	}
	sen5x.deviceInfo.serialNumber = string(data)
	return nil
}

func (sen5x *SEN5X) ProductName() error {
	err := SEN5X_PRODUCT_NAME.Write(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to write product name command: %q", err)
	}
	data, err := SEN5X_PRODUCT_NAME.Read(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to read product name number: %q", err)
	}
	sen5x.deviceInfo.productName = string(data)
	return nil
}

func (sen5x *SEN5X) Versions() error {
	err := SEN5X_VERSION.Write(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to write versions command: %q", err)
	}
	data, err := SEN5X_VERSION.Read(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to read versions number: %q", err)
	}
	sen5x.deviceInfo.firmwareMajorVersion = data[0]
	sen5x.deviceInfo.firmwareMinorVersion = data[1]
	sen5x.deviceInfo.firmwareDebug = data[2] == 1
	sen5x.deviceInfo.hardwareMajorVersion = data[3]
	sen5x.deviceInfo.hardwareMinorVersion = data[4]
	sen5x.deviceInfo.protocolMajorVersion = data[5]
	sen5x.deviceInfo.protocolMinorVersion = data[6]
	return nil
}

func (sen5x *SEN5X) Status() error {
	err := SEN5X_READ_STATUS.Write(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to write status command: %q", err)
	}
	data, err := SEN5X_READ_STATUS.Read(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to read status: %q", err)
	}
	sen5x.deviceInfo.status = binary.BigEndian.Uint32(data)
	return nil
}
