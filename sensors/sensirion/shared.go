package sensirion

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"periph.io/x/conn/v3/i2c"
)

type Command struct {
	code        uint16
	description string
	delay       time.Duration
	size        uint8
}

func crc8(buffer []byte) byte {
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

func checkBufferCRC(buffer []byte) error {
	for i := 0; i < len(buffer); i += 3 {
		crcBuffer := make([]byte, 2)
		crcBuffer[0] = buffer[i]
		crcBuffer[1] = buffer[i+1]
		if crc8(crcBuffer) != buffer[i+2] {
			return fmt.Errorf("CRC check failed")
		}
	}
	return nil
}

func (cmd *Command) Write(device *i2c.Dev, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()

	encodedCommand := make([]byte, 2)
	binary.BigEndian.PutUint16(encodedCommand, cmd.code)
	if _, err := device.Write(encodedCommand); err != nil {
		return fmt.Errorf("error while running %s: %q", cmd.description, err)
	}

	if cmd.delay > 0 {
		time.Sleep(cmd.delay)
	}
	return nil
}

func (cmd *Command) Read(device *i2c.Dev, mu *sync.Mutex) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	actualSize := (cmd.size / 2) * 3

	c := make([]byte, 2)
	r := make([]byte, actualSize)
	binary.BigEndian.PutUint16(c, cmd.code)
	if err := device.Tx(c, r); err != nil {
		return nil, fmt.Errorf("error while %s: %q", cmd.description, err)
	}

	if err := checkBufferCRC(r); err != nil {
		return nil, err
	}

	response := make([]byte, cmd.size)
	idx := 0
	for i, b := range r {
		if (i+1)%3 != 0 {
			response[idx] = b
			idx++
		}
	}

	if cmd.delay > 0 {
		time.Sleep(cmd.delay)
	}
	return response, nil
}

func (cmd *Command) WriteUint16(device *i2c.Dev, mu *sync.Mutex, value uint16) error {
	mu.Lock()
	defer mu.Unlock()

	encodedCommand := make([]byte, 5)
	binary.BigEndian.PutUint16(encodedCommand, cmd.code)
	encodedCommand[2] = byte((value >> 8) & 0xFF)
	encodedCommand[3] = byte(value & 0xFF)
	encodedCommand[4] = crc8(encodedCommand[2:4])
	if _, err := device.Write(encodedCommand); err != nil {
		return fmt.Errorf("error while running %s: %q", cmd.description, err)
	}

	if cmd.delay > 0 {
		time.Sleep(cmd.delay)
	}
	return nil
}

func (cmd *Command) WriteUint32(device *i2c.Dev, mu *sync.Mutex, value uint32) error {
	mu.Lock()
	defer mu.Unlock()

	encodedCommand := make([]byte, 8)
	binary.BigEndian.PutUint16(encodedCommand, cmd.code)
	encodedCommand[2] = byte((value & 0xFF000000) >> 24)
	encodedCommand[3] = byte((value & 0x00FF0000) >> 16)
	encodedCommand[4] = crc8(encodedCommand[2:4])
	encodedCommand[5] = byte((value & 0x0000FF00) >> 8)
	encodedCommand[6] = byte((value & 0x000000FF) >> 0)
	encodedCommand[7] = crc8(encodedCommand[5:7])
	if _, err := device.Write(encodedCommand); err != nil {
		return fmt.Errorf("error while running %s: %q", cmd.description, err)
	}

	if cmd.delay > 0 {
		time.Sleep(cmd.delay)
	}
	return nil
}

func (cmd *Command) ReadUint16(device *i2c.Dev, mu *sync.Mutex, value uint16) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	c := make([]byte, 5)
	r := make([]byte, cmd.size)
	binary.BigEndian.PutUint16(c, cmd.code)
	c[2] = byte((value >> 8) & 0xFF)
	c[3] = byte(value & 0xFF)
	c[4] = crc8(c[2:4])
	if err := device.Tx(c, r); err != nil {
		return nil, fmt.Errorf("error while %s: %q", cmd.description, err)
	}
	if err := checkBufferCRC(r); err != nil {
		return nil, err
	}

	if cmd.delay > 0 {
		time.Sleep(cmd.delay)
	}
	return r, nil
}
