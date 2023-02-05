package sensors

import (
	"encoding/binary"
	"fmt"

	"azuremyst.org/go-home-sensors/log"
	"periph.io/x/conn/v3/i2c"
)

type PMSA003I struct {
	device         *i2c.Dev
	PM10Standard   uint16
	PM25Standard   uint16
	PM100Standard  uint16
	PM10Env        uint16
	PM25Env        uint16
	PM100Env       uint16
	Particles03um  uint16
	Particles05um  uint16
	Particles10um  uint16
	Particles25um  uint16
	Particles50um  uint16
	Particles100um uint16
}

func NewPMSA003I(device *i2c.Dev) PMSA003I {
	return PMSA003I{device: device}
}

func (pmsa *PMSA003I) Read() {
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

	pmsa.PM10Standard = binary.BigEndian.Uint16(response[4:6])
	pmsa.PM25Standard = binary.BigEndian.Uint16(response[6:8])
	pmsa.PM100Standard = binary.BigEndian.Uint16(response[8:10])
	pmsa.PM10Env = binary.BigEndian.Uint16(response[10:12])
	pmsa.PM25Env = binary.BigEndian.Uint16(response[12:14])
	pmsa.PM100Env = binary.BigEndian.Uint16(response[14:16])
	pmsa.Particles03um = binary.BigEndian.Uint16(response[16:18])
	pmsa.Particles05um = binary.BigEndian.Uint16(response[18:20])
	pmsa.Particles10um = binary.BigEndian.Uint16(response[20:22])
	pmsa.Particles25um = binary.BigEndian.Uint16(response[22:24])
	pmsa.Particles50um = binary.BigEndian.Uint16(response[24:26])
	pmsa.Particles100um = binary.BigEndian.Uint16(response[26:28])
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