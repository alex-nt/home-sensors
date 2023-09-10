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
	mu sync.Mutex
)

type Command struct {
	code        uint16
	description string
	delay       time.Duration
	size        uint8
}

var (
	SCD4XX_REINIT                          = Command{code: 0x3646, description: "Reinit", delay: time.Duration(0.02 * float64(time.Second)), size: 3}
	SCD4X_FACTORYRESET                     = Command{code: 0x3632, description: "Factory reset", delay: time.Duration(1.2 * float64(time.Second)), size: 0}
	SCD4X_FORCEDRECAL                      = Command{code: 0x362F, description: "Force recal", delay: time.Duration(0.5 * float64(time.Second)), size: 3}
	SCD4X_SELFTEST                         = Command{code: 0x3639, description: "Self test", delay: time.Duration(10 * float64(time.Second)), size: 3}
	SCD4X_DATAREADY                        = Command{code: 0xE4B8, description: "Data Ready", delay: time.Duration(0.001 * float64(time.Second)), size: 3}
	SCD4X_STOPPERIODICMEASUREMENT          = Command{code: 0x3F86, description: "Stop periodic measurement", delay: time.Duration(0.5 * float64(time.Second)), size: 0}
	SCD4X_STARTPERIODICMEASUREMENT         = Command{code: 0x21B1, description: "Start periodic measurement", delay: time.Duration(0 * float64(time.Second)), size: 0}
	SCD4X_STARTLOWPOWERPERIODICMEASUREMENT = Command{code: 0x21AC, description: "Start low power periodic measurement", delay: time.Duration(0 * float64(time.Second)), size: 0}
	SCD4X_READMEASUREMENT                  = Command{code: 0xEC05, description: "Read measurement", delay: time.Duration(0.001 * float64(time.Second)), size: 9}
	SCD4X_SERIALNUMBER                     = Command{code: 0x3682, description: "Serial number", delay: time.Duration(0.001 * float64(time.Second)), size: 9}
	SCD4X_GETTEMPOFFSET                    = Command{code: 0x2318, description: "Get temp offset", delay: time.Duration(0.001 * float64(time.Second)), size: 3}
	SCD4X_SETTEMPOFFSET                    = Command{code: 0x241D, description: "Set temp offset", delay: time.Duration(0 * float64(time.Second)), size: 0}
	SCD4X_GETALTITUDE                      = Command{code: 0x2322, description: "Get altitude", delay: time.Duration(0.001 * float64(time.Second)), size: 3}
	SCD4X_SETALTITUDE                      = Command{code: 0x2427, description: "Set altitude", delay: time.Duration(0 * float64(time.Second)), size: 0}
	SCD4X_SETPRESSURE                      = Command{code: 0xE000, description: "Set pressure", delay: time.Duration(0 * float64(time.Second)), size: 0}
	SCD4X_PERSISTSETTINGS                  = Command{code: 0x3615, description: "Persist settings", delay: time.Duration(0.8 * float64(time.Second)), size: 0}
	SCD4X_GETASCE                          = Command{code: 0x2313, description: "Get asce", delay: time.Duration(0.001 * float64(time.Second)), size: 3}
	SCD4X_SETASCE                          = Command{code: 0x2416, description: "Set asce", delay: time.Duration(0 * float64(time.Second)), size: 0}
)

type SCD4X struct {
	device           *i2c.Dev
	Temperature      float64
	RelativeHumidity float64
	CO2              uint16
}

func init() {
	sensors.RegisterSensor(&SCD4X{})
}

func (scd4x *SCD4X) Initialize(bus i2c.Bus, addr uint16) {
	scd4x.device = &i2c.Dev{Addr: addr, Bus: bus}
	scd4x.StartPeriodicMeasurement()
}

func (scd4x *SCD4X) Name() string {
	return "scd4x"
}

func (scd4x *SCD4X) Family(name string) bool {
	return len(name) == 5 && strings.HasPrefix(strings.ToLower(name), "scd4")
}

func (scd4x *SCD4X) Collect() []sensors.MeasurementRecording {
	measurements := make([]sensors.MeasurementRecording, 3)
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.Temperature,
		Value:   float64(scd4x.GetTemperature()),
		Sensor:  scd4x.Name(),
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.Humidity,
		Value:   float64(scd4x.GetRelativeHumidity()),
		Sensor:  scd4x.Name(),
	})
	measurements = append(measurements, sensors.MeasurementRecording{
		Measure: &sensors.CarbonDioxide,
		Value:   float64(scd4x.GetCO2()),
		Sensor:  scd4x.Name(),
	})
	return measurements
}

func (scd4x *SCD4X) ReadCommand(command *Command) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	c := make([]byte, 2)
	r := make([]byte, command.size)
	binary.BigEndian.PutUint16(c, command.code)
	if err := scd4x.device.Tx(c, r); err != nil {
		return nil, fmt.Errorf("error while %s: %v", command.description, err)
	}
	if err := scd4x.checkBufferCRC(r); err != nil {
		return nil, err
	}

	if command.delay > 0 {
		time.Sleep(command.delay)
	}
	return r, nil
}

func (scd4x *SCD4X) WriteCommand(command *Command) error {
	mu.Lock()
	defer mu.Unlock()

	encodedCommand := make([]byte, 2)
	binary.BigEndian.PutUint16(encodedCommand, command.code)
	if _, err := scd4x.device.Write(encodedCommand); err != nil {
		return fmt.Errorf("error while running %s: %v", command.description, err)
	}

	if command.delay > 0 {
		time.Sleep(command.delay)
	}
	return nil
}

func (scd4x *SCD4X) WriteCommandValue(command *Command, value uint16) error {
	mu.Lock()
	defer mu.Unlock()

	encodedCommand := make([]byte, 5)
	binary.BigEndian.PutUint16(encodedCommand, command.code)
	encodedCommand[2] = byte((value >> 8) & 0xFF)
	encodedCommand[3] = byte(value & 0xFF)
	encodedCommand[4] = scd4x.crc8(encodedCommand[2:4])
	if _, err := scd4x.device.Write(encodedCommand); err != nil {
		return fmt.Errorf("error while running %s: %v", command.description, err)
	}

	if command.delay > 0 {
		time.Sleep(command.delay)
	}
	return nil
}

func (scd4x *SCD4X) ReadCommandValue(command *Command, value uint16) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	c := make([]byte, 5)
	r := make([]byte, command.size)
	binary.BigEndian.PutUint16(c, command.code)
	c[2] = byte((value >> 8) & 0xFF)
	c[3] = byte(value & 0xFF)
	c[4] = scd4x.crc8(c[2:4])
	if err := scd4x.device.Tx(c, r); err != nil {
		return nil, fmt.Errorf("error while %s: %v", command.description, err)
	}
	if err := scd4x.checkBufferCRC(r); err != nil {
		return nil, err
	}

	if command.delay > 0 {
		time.Sleep(command.delay)
	}
	return r, nil
}

func (scd4x *SCD4X) dataReady() bool {
	response, err := scd4x.ReadCommand(&SCD4X_DATAREADY)
	if err != nil {
		log.ErrorLog.Printf("Data not ready %v", err)
		return false
	}
	return !((response[0]&0x07 == 0) && (response[1] == 0))
}

func (scd4x *SCD4X) readData() error {
	response, err := scd4x.ReadCommand(&SCD4X_READMEASUREMENT)
	scd4x.CO2 = binary.BigEndian.Uint16(response[0:2])
	scd4x.Temperature = (-45 + 175*(float64(binary.BigEndian.Uint16(response[3:5]))/math.Pow(2, 16)))
	scd4x.RelativeHumidity = 100 * (float64(binary.BigEndian.Uint16(response[6:8])) / math.Pow(2, 16))
	return err
}

func (scd4x *SCD4X) GetCO2() uint16 {
	if scd4x.dataReady() {
		scd4x.readData()
	}
	return scd4x.CO2
}

func (scd4x *SCD4X) GetTemperature() float64 {
	if scd4x.dataReady() {
		scd4x.readData()
	}
	return scd4x.Temperature
}

func (scd4x *SCD4X) GetRelativeHumidity() float64 {
	if scd4x.dataReady() {
		scd4x.readData()
	}
	return scd4x.RelativeHumidity
}

func (scd4x *SCD4X) StopPeriodicMeasurement() {
	scd4x.WriteCommand(&SCD4X_STOPPERIODICMEASUREMENT)
}

func (scd4x *SCD4X) StartPeriodicMeasurement() {
	scd4x.WriteCommand(&SCD4X_STARTPERIODICMEASUREMENT)
}

func (scd4x *SCD4X) reinit() {
	scd4x.StopPeriodicMeasurement()
	scd4x.WriteCommand(&SCD4XX_REINIT)
}

func (scd4x *SCD4X) FactoryReset() {
	scd4x.StopPeriodicMeasurement()
	scd4x.WriteCommand(&SCD4X_FACTORYRESET)
}

func (scd4x *SCD4X) crc8(buffer []byte) byte {
	crc := byte(0xFF)
	for _, v := range buffer {
		crc ^= v
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc = crc << 1
			}
		}
	}
	return crc
}

func (scd4x *SCD4X) checkBufferCRC(buffer []byte) error {
	for i := 0; i < len(buffer); i += 3 {
		crcBuffer := make([]byte, 2)
		crcBuffer[0] = buffer[i]
		crcBuffer[1] = buffer[i+1]
		if scd4x.crc8(crcBuffer) != buffer[i+2] {
			return fmt.Errorf("CRC check failed")
		}
	}
	return nil
}

func (scd4x *SCD4X) IsCalibrationEnabled() bool {
	response, err := scd4x.ReadCommand(&SCD4X_GETASCE)
	if err != nil {
		log.ErrorLog.Printf("Calibration status could not be read %v", err)
		return false
	}
	return response[1] == 1
}

func (scd4x *SCD4X) ToggleCalibration(enable bool) {
	var value uint16
	if enable {
		value = 1
	}
	scd4x.WriteCommandValue(&SCD4X_SETASCE, value)
}

func (scd4x *SCD4X) ForceCalibration(targetCO2 uint16) error {
	scd4x.StopPeriodicMeasurement()
	response, err := scd4x.ReadCommandValue(&SCD4X_FORCEDRECAL, targetCO2)
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
	scd4x.StopPeriodicMeasurement()
	response, err := scd4x.ReadCommand(&SCD4X_SELFTEST)
	if err != nil {
		return err
	}
	if response[0] != 0 || response[1] != 0 {
		return fmt.Errorf("self test failed")
	}
	return nil
}

func (scd4x *SCD4X) SerialNumber() ([]uint8, error) {
	serialNumber := make([]uint8, 6)
	response, err := scd4x.ReadCommand(&SCD4X_SERIALNUMBER)
	if err != nil {
		return nil, err
	}
	serialNumber[0] = response[0]
	serialNumber[1] = response[1]
	serialNumber[2] = response[3]
	serialNumber[3] = response[4]
	serialNumber[4] = response[6]
	serialNumber[5] = response[7]
	return serialNumber, nil
}

func (scd4x *SCD4X) StartLowPeriodicMeasurement() {
	scd4x.WriteCommand(&SCD4X_STARTLOWPOWERPERIODICMEASUREMENT)
}

func (scd4x *SCD4X) PersistSettings() {
	scd4x.WriteCommand(&SCD4X_PERSISTSETTINGS)
}

func (scd4x *SCD4X) SetAmbientPressure(ambientPressure uint16) error {
	return scd4x.WriteCommandValue(&SCD4X_SETPRESSURE, ambientPressure)
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
	response, err := scd4x.ReadCommand(&SCD4X_GETTEMPOFFSET)
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
	return scd4x.WriteCommandValue(&SCD4X_SETTEMPOFFSET, temp)
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
	response, err := scd4x.ReadCommand(&SCD4X_GETALTITUDE)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(response[0:2]), nil
}

func (scd4x *SCD4X) SetAltitude(altitude uint16) error {
	return scd4x.WriteCommandValue(&SCD4X_SETALTITUDE, altitude)
}
