package sensirion

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"
	"periph.io/x/conn/v3/i2c"
)

var (
	SCD4XX_REINIT                          = Command{code: 0x3646, description: "Reinit", delay: time.Duration(30 * time.Millisecond), size: 2}
	SCD4X_FACTORYRESET                     = Command{code: 0x3632, description: "Factory reset", delay: time.Duration(1200 * time.Millisecond), size: 0}
	SCD4X_FORCEDRECAL                      = Command{code: 0x362F, description: "Force recal", delay: time.Duration(400 * time.Millisecond), size: 2}
	SCD4X_SELFTEST                         = Command{code: 0x3639, description: "Self test", delay: time.Duration(10000 * time.Millisecond), size: 2}
	SCD4X_DATAREADY                        = Command{code: 0xE4B8, description: "Data Ready", delay: time.Duration(1 * time.Millisecond), size: 2}
	SCD4X_STOPPERIODICMEASUREMENT          = Command{code: 0x3F86, description: "Stop periodic measurement", delay: time.Duration(500 * time.Millisecond), size: 0}
	SCD4X_STARTPERIODICMEASUREMENT         = Command{code: 0x21B1, description: "Start periodic measurement", delay: time.Duration(0), size: 0}
	SCD4X_STARTLOWPOWERPERIODICMEASUREMENT = Command{code: 0x21AC, description: "Start low power periodic measurement", delay: time.Duration(0), size: 0}
	SCD4X_READMEASUREMENT                  = Command{code: 0xEC05, description: "Read measurement", delay: time.Duration(1 * time.Millisecond), size: 6}
	SCD4X_SERIALNUMBER                     = Command{code: 0x3682, description: "Serial number", delay: time.Duration(1 * time.Millisecond), size: 6}
	SCD4X_GETTEMPOFFSET                    = Command{code: 0x2318, description: "Get temp offset", delay: time.Duration(1 * time.Millisecond), size: 2}
	SCD4X_SETTEMPOFFSET                    = Command{code: 0x241D, description: "Set temp offset", delay: time.Duration(1 * time.Millisecond), size: 0}
	SCD4X_GETALTITUDE                      = Command{code: 0x2322, description: "Get altitude", delay: time.Duration(1 * time.Millisecond), size: 2}
	SCD4X_SETALTITUDE                      = Command{code: 0x2427, description: "Set altitude", delay: time.Duration(1 * time.Millisecond), size: 0}
	SCD4X_SETPRESSURE                      = Command{code: 0xE000, description: "Set pressure", delay: time.Duration(1 * time.Millisecond), size: 0}
	SCD4X_PERSISTSETTINGS                  = Command{code: 0x3615, description: "Persist settings", delay: time.Duration(800 * time.Millisecond), size: 0}
	SCD4X_GETASCE                          = Command{code: 0x2313, description: "Get asce", delay: time.Duration(1 * time.Millisecond), size: 2}
	SCD4X_SETASCE                          = Command{code: 0x2416, description: "Set asce", delay: time.Duration(1 * time.Millisecond), size: 0}
	SCD4X_WAKEUP                           = Command{code: 0x36f6, description: "Wake up", delay: time.Duration(30 * time.Millisecond), size: 0}
)

type SCD4XMeasurement struct {
	Humidity    float64
	Temperature float64
	CO2         float64
}

type SCD4XDeviceInfo struct {
	serialNumber string
}

type SCD4X struct {
	device *i2c.Dev
	mu     sync.Mutex

	deviceInfo SCD4XDeviceInfo
	data       SCD4XMeasurement
}

func init() {
	sensors.RegisterSensor(&SCD4X{})
}

func (scd4x *SCD4X) Initialize(bus i2c.Bus, addr uint16) {
	scd4x.device = &i2c.Dev{Addr: addr, Bus: bus}
	if err := SCD4X_WAKEUP.Write(scd4x.device, &scd4x.mu); err != nil {
		log.ErrorLog.Printf("Failed to wakeup: %q", err)
		return
	}
	if err := SCD4X_STOPPERIODICMEASUREMENT.Write(scd4x.device, &scd4x.mu); err != nil {
		log.ErrorLog.Printf("Failed to stop measurements: %q", err)
		return
	}
	if err := SCD4XX_REINIT.Write(scd4x.device, &scd4x.mu); err != nil {
		log.ErrorLog.Printf("Failed to reinit device: %q", err)
		return
	}
	if err := scd4x.SerialNumber(); err != nil {
		log.ErrorLog.Printf("Failed to read SN: %q", err)
	}
	log.InfoLog.Printf(`Sensirion SCD4X
	SerialNumber: %s`, scd4x.deviceInfo.serialNumber)
	scd4x.StartPeriodicMeasurement()
}

func (scd4x *SCD4X) Name() string {
	return "scd4x"
}

func (scd4x *SCD4X) Family(name string) bool {
	return len(name) == 5 && strings.HasPrefix(strings.ToLower(name), "scd4")
}

func (scd4x *SCD4X) Collect() []sensors.MeasurementRecording {
	if scd4x.dataReady() {
		scd4x.readData()
	}

	measurements := make([]sensors.MeasurementRecording, 0)
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.Temperature,
		Value:   scd4x.data.Temperature,
		Sensor:  scd4x.Name(),
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.Humidity,
		Value:   scd4x.data.Humidity,
		Sensor:  scd4x.Name(),
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.CarbonDioxide,
		Value:   scd4x.data.CO2,
		Sensor:  scd4x.Name(),
	})
	return measurements
}

func (scd4x *SCD4X) dataReady() bool {
	response, err := SCD4X_DATAREADY.Read(scd4x.device, &scd4x.mu)
	if err != nil {
		log.ErrorLog.Printf("Data not ready %q", err)
		return false
	}
	return !((response[0]&0x07 == 0) && (response[1] == 0))
}

func (scd4x *SCD4X) readData() error {
	response, err := SCD4X_READMEASUREMENT.Read(scd4x.device, &scd4x.mu)
	scd4x.data.CO2 = float64(binary.BigEndian.Uint16(response[0:2]))
	scd4x.data.Temperature = (-45 + 175*(float64(binary.BigEndian.Uint16(response[2:4]))/math.Pow(2, 16)))
	scd4x.data.Humidity = 100 * (float64(binary.BigEndian.Uint16(response[4:6])) / math.Pow(2, 16))
	return err
}

func (scd4x *SCD4X) StartPeriodicMeasurement() {
	SCD4X_STARTPERIODICMEASUREMENT.Write(scd4x.device, &scd4x.mu)
}

func (scd4x *SCD4X) FactoryReset() {
	SCD4X_STOPPERIODICMEASUREMENT.Write(scd4x.device, &scd4x.mu)
	SCD4X_FACTORYRESET.Write(scd4x.device, &scd4x.mu)
}

func (scd4x *SCD4X) IsCalibrationEnabled() bool {
	response, err := SCD4X_GETASCE.Read(scd4x.device, &scd4x.mu)
	if err != nil {
		log.ErrorLog.Printf("Calibration status could not be read %q", err)
		return false
	}
	return response[1] == 1
}

func (scd4x *SCD4X) ToggleCalibration(enable bool) {
	var value uint16
	if enable {
		value = 1
	}
	SCD4X_SETASCE.WriteUint16(scd4x.device, &scd4x.mu, value)
}

func (scd4x *SCD4X) ForceCalibration(targetCO2 uint16) error {
	SCD4X_STOPPERIODICMEASUREMENT.Write(scd4x.device, &scd4x.mu)
	response, err := SCD4X_FORCEDRECAL.ReadUint16(scd4x.device, &scd4x.mu, targetCO2)
	if err != nil {
		return err
	}
	var unpackedData uint16
	if err = binary.Read(bytes.NewReader(response[0:2]), binary.BigEndian, &unpackedData); err != nil {
		return err
	}

	if unpackedData == 0xFFFF {
		return fmt.Errorf("force recalibration failed, please make sure sensor is active for 3m first")
	}
	return nil
}

func (scd4x *SCD4X) Test() error {
	SCD4X_STOPPERIODICMEASUREMENT.Write(scd4x.device, &scd4x.mu)
	response, err := SCD4X_SELFTEST.Read(scd4x.device, &scd4x.mu)
	if err != nil {
		return err
	}
	if response[0] != 0 || response[1] != 0 {
		return fmt.Errorf("self test failed")
	}
	return nil
}

func (scd4x *SCD4X) SerialNumber() error {
	response, err := SCD4X_SERIALNUMBER.Read(scd4x.device, &scd4x.mu)
	if err != nil {
		return err
	}

	scd4x.deviceInfo.serialNumber = bytesToString(response)
	return nil
}

func (scd4x *SCD4X) StartLowPeriodicMeasurement() {
	SCD4X_STARTLOWPOWERPERIODICMEASUREMENT.Write(scd4x.device, &scd4x.mu)
}

func (scd4x *SCD4X) PersistSettings() {
	SCD4X_PERSISTSETTINGS.Write(scd4x.device, &scd4x.mu)
}

func (scd4x *SCD4X) SetAmbientPressure(ambientPressure uint16) error {
	return SCD4X_SETPRESSURE.WriteUint16(scd4x.device, &scd4x.mu, ambientPressure)
}

/*
Specifies the offset to be added to the reported measurements to account for a bias in
the measured signal. Value is in degrees Celsius with a resolution of 0.01 degrees and a
maximum value of 374 C
.. note::

	This value will NOT be saved and will be reset on boot unless saved with
	persist_settings().
*/
func (scd4x *SCD4X) GetTempetatureOffset() (float64, error) {
	response, err := SCD4X_GETTEMPOFFSET.Read(scd4x.device, &scd4x.mu)
	if err != nil {
		return 0, err
	}
	unwrappedvalue := binary.BigEndian.Uint16(response[0:2])
	return 175.0 * float64(unwrappedvalue) / math.Pow(2, 16), nil
}

func (scd4x *SCD4X) SetTemperatureOffset(value uint16) error {
	if value > 374 {
		return fmt.Errorf("offset value must be less than or equal to 374 degrees Celsius")
	}
	temp := value * uint16(math.Pow(2, 16)) / 175
	return SCD4X_SETTEMPOFFSET.WriteUint16(scd4x.device, &scd4x.mu, temp)
}

/*
Specifies the altitude at the measurement location in meters above sea level. Setting
this value adjusts the CO2 measurement calculations to account for the air pressure's effect
on readings.
.. note::

	This value will NOT be saved and will be reset on boot unless saved with
	persist_settings().
*/
func (scd4x *SCD4X) GetAltitude() (uint16, error) {
	response, err := SCD4X_GETALTITUDE.Read(scd4x.device, &scd4x.mu)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(response[0:2]), nil
}

func (scd4x *SCD4X) SetAltitude(altitude uint16) error {
	return SCD4X_SETALTITUDE.WriteUint16(scd4x.device, &scd4x.mu, altitude)
}
