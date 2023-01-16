package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"periph.io/x/conn/v3/i2c"
)

type Command struct {
	code        uint16
	description string
	delay       time.Duration
}

type Response struct {
	data []byte
	crc  byte
}

var (
	SCD4XX_REINIT                          = Command{code: 0x3646, description: "Reinit", delay: time.Duration(0.02 * float64(time.Second))}
	SCD4X_FACTORYRESET                     = Command{code: 0x3632, description: "Factory reset", delay: time.Duration(1.2 * float64(time.Second))}
	SCD4X_FORCEDRECAL                      = Command{code: 0x362F, description: "Force recal", delay: time.Duration(0.5 * float64(time.Second))}
	SCD4X_SELFTEST                         = Command{code: 0x3639, description: "Self test", delay: time.Duration(10 * float64(time.Second))}
	SCD4X_DATAREADY                        = Command{code: 0xE4B8, description: "Data Ready", delay: time.Duration(0.001 * float64(time.Second))}
	SCD4X_STOPPERIODICMEASUREMENT          = Command{code: 0x3F86, description: "Stop periodic measurement", delay: time.Duration(0.5 * float64(time.Second))}
	SCD4X_STARTPERIODICMEASUREMENT         = Command{code: 0x21B1, description: "Start periodic measurement", delay: time.Duration(0 * float64(time.Second))}
	SCD4X_STARTLOWPOWERPERIODICMEASUREMENT = Command{code: 0x21AC, description: "Start low power periodic measurement", delay: time.Duration(0 * float64(time.Second))}
	SCD4X_READMEASUREMENT                  = Command{code: 0xEC05, description: "Read measurement", delay: time.Duration(0.001 * float64(time.Second))}
	SCD4X_SERIALNUMBER                     = Command{code: 0x3682, description: "Serial number", delay: time.Duration(0.001 * float64(time.Second))}
	SCD4X_GETTEMPOFFSET                    = Command{code: 0x2318, description: "Get temp offset", delay: time.Duration(0.001 * float64(time.Second))}
	SCD4X_SETTEMPOFFSET                    = Command{code: 0x241D, description: "Set temp offset", delay: time.Duration(0 * float64(time.Second))}
	SCD4X_GETALTITUDE                      = Command{code: 0x2322, description: "Get altitude", delay: time.Duration(0.001 * float64(time.Second))}
	SCD4X_SETALTITUDE                      = Command{code: 0x2427, description: "Set altitude", delay: time.Duration(0 * float64(time.Second))}
	SCD4X_SETPRESSURE                      = Command{code: 0xE000, description: "Set pressure", delay: time.Duration(0 * float64(time.Second))}
	SCD4X_PERSISTSETTINGS                  = Command{code: 0x3615, description: "Persist settings", delay: time.Duration(0.8 * float64(time.Second))}
	SCD4X_GETASCE                          = Command{code: 0x2313, description: "Get asce", delay: time.Duration(0.001 * float64(time.Second))}
	SCD4X_SETASCE                          = Command{code: 0x2416, description: "Set asce", delay: time.Duration(0 * float64(time.Second))}
)

type SCD4X struct {
	device    *i2c.Dev
	buffer    [18]byte
	cmd       [2]byte
	crcBuffer [2]byte

	Temperature      float32
	RelativeHumidity float32
	CO2              int64
}

func NewSCD4X(device *i2c.Dev) SCD4X {
	return SCD4X{device: device}
}

func ReadCommand(command *Command) (Response, error) {
	return Response{}, nil
}

func (scd4x *SCD4X) WriteCommand(command *Command) error {
	encodedCommand := make([]byte, 2)
	binary.BigEndian.PutUint16(encodedCommand, command.code)
	if _, err := scd4x.device.Write(encodedCommand); err != nil {
		return fmt.Errorf("Error while running %s: %v", command.description, err)
	}

	if command.delay > 0 {
		time.Sleep(command.delay)
	}
	return nil
}
