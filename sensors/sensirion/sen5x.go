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
	SEN5X_RESET         = Command{code: 0xD304, description: "Reset device", delay: time.Duration(0.02 * float64(time.Second)), size: 2}
	SEN5X_SERIAL_NUMBER = Command{code: 0xD033, description: "Serial number", delay: time.Duration(0.05 * float64(time.Second)), size: 32}
	SEN5X_PRODUCT_NAME  = Command{code: 0xD014, description: "Product name", delay: time.Duration(0.05 * float64(time.Second)), size: 32}
	SEN5X_VERSION       = Command{code: 0xD100, description: "Versions", delay: time.Duration(0.02 * float64(time.Second)), size: 8}
	SEN5X_READ_STATUS   = Command{code: 0xD206, description: "Read status", delay: time.Duration(0.02 * float64(time.Second)), size: 4}
)

type SEN5X struct {
	device *i2c.Dev
	mu     sync.Mutex

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
	Status: %d
	FirmwareDebug: %t
	Versions:
		Firmware: %d.%d
		Hardware: %d.%d
		Protocol: %d.%d`,
		sen5x.productName, sen5x.status, sen5x.firmwareDebug,
		sen5x.firmwareMajorVersion, sen5x.firmwareMinorVersion,
		sen5x.hardwareMajorVersion, sen5x.hardwareMinorVersion,
		sen5x.protocolMajorVersion, sen5x.protocolMinorVersion)
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
	sen5x.serialNumber = string(data)
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
	sen5x.productName = string(data)
	return nil
}

func (sen5x *SEN5X) Versions() error {
	data, err := SEN5X_VERSION.Read(sen5x.device, &sen5x.mu)
	if err != nil {
		return fmt.Errorf("failed to read versions number: %q", err)
	}
	sen5x.firmwareMajorVersion = data[0]
	sen5x.firmwareMinorVersion = data[1]
	sen5x.firmwareDebug = data[2] == 1
	sen5x.hardwareMajorVersion = data[3]
	sen5x.hardwareMinorVersion = data[4]
	sen5x.protocolMajorVersion = data[5]
	sen5x.protocolMinorVersion = data[6]
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
	sen5x.status = binary.BigEndian.Uint32(data)
	return nil
}
