package sensors

import (
	"fmt"
	"time"

	"azuremyst.org/go-home-sensors/log"
	"periph.io/x/conn/v3/i2c"
)

const (
	BME68X_PERIOD_POLL      = uint32(10000)
	BME68X_CHIP_ID          = uint8(0x61)
	BME68X_PERIOD_RESET     = uint32(10000)
	BME68X_I2C_ADDRESS_LOW  = uint8(0x76)
	BME68X_I2C_ADDRESS_HIGH = uint8(0x77)

	BME68X_SOFT_RESET_CMD = uint8(0xb6)
	BME68X_OK             = uint8(0)
	BME68X_VARIANT_HIGH   = uint8(0x01)
)

// Errors
const (
	BME68X_E_NULL_PTR       = int8(-1) // Null pointer passed
	BME68X_E_COM_FAIL       = int8(-2) // Communication failure
	BME68X_E_DEV_NOT_FOUND  = int8(-3) // Sensor not found
	BME68X_E_INVALID_LENGTH = int8(-4) // Incorrect length parameter
	BME68X_E_SELF_TEST      = int8(-5) // Self test fail error
)

// Warnings
const (
	BME68X_W_DEFINE_OP_MODE       = int8(1)  // Define a valid operation mode
	BME68X_W_NO_NEW_DATA          = int8(2)  // No new data was found
	BME68X_W_DEFINE_SHD_HEATR_DUR = int8(3)  // Define the shared heating duration
	BME68X_I_PARAM_CORR           = uint8(1) // Information - only available via bme68x_dev.info_msg
)

// Register address in i2c
const (
	BME68X_REG_COEFF3        = uint8(0x00) // Register for 3rd group of coefficients
	BME68X_REG_FIELD0        = uint8(0x1d) // 0th Field address
	BME68X_REG_IDAC_HEAT0    = uint8(0x50) // 0th Current DAC address
	BME68X_REG_RES_HEAT0     = uint8(0x5a) // 0th Res heat address
	BME68X_REG_GAS_WAIT0     = uint8(0x64) // 0th Gas wait address
	BME68X_REG_SHD_HEATR_DUR = uint8(0x6E) // Shared heating duration address
	BME68X_REG_CTRL_GAS_0    = uint8(0x70) // CTRL_GAS_0 address
	BME68X_REG_CTRL_GAS_1    = uint8(0x71) // CTRL_GAS_1 address
	BME68X_REG_CTRL_HUM      = uint8(0x72) // CTRL_HUM address
	BME68X_REG_CTRL_MEAS     = uint8(0x74) // CTRL_MEAS address
	BME68X_REG_CONFIG        = uint8(0x75) // CONFIG address
	BME68X_REG_MEM_PAGE      = uint8(0xf3) // MEM_PAGE address
	BME68X_REG_UNIQUE_ID     = uint8(0x83) // Unique ID address
	BME68X_REG_COEFF1        = uint8(0x8a) // Register for 1st group of coefficients
	BME68X_REG_CHIP_ID       = uint8(0xd0) // Chip ID address
	BME68X_REG_SOFT_RESET    = uint8(0xe0) // Soft reset address
	BME68X_REG_COEFF2        = uint8(0xe1) // Register for 2nd group of coefficients
	BME68X_REG_VARIANT_ID    = uint8(0xF0) // Variant ID Register
)

// Oversampling setting macros
const (
	BME68X_OS_NONE = uint8(0) // Switch off measurement
	BME68X_OS_1X   = uint8(1) // Perform 1 measurement
	BME68X_OS_2X   = uint8(2) // Perform 2 measurement
	BME68X_OS_4X   = uint8(3) // Perform 4 measurement
	BME68X_OS_8X   = uint8(4) // Perform 8 measurement
	BME68X_OS_16X  = uint8(5) // Perform 16 measurement
)

// IIR Filter settings
const (
	BME68X_FILTER_OFF      = uint8(0) // Switch off the filter
	BME68X_FILTER_SIZE_1   = uint8(1) // Filter coefficient of 2
	BME68X_FILTER_SIZE_3   = uint8(2) // Filter coefficient of 4
	BME68X_FILTER_SIZE_7   = uint8(3) // Filter coefficient of 8
	BME68X_FILTER_SIZE_15  = uint8(4) // Filter coefficient of 16
	BME68X_FILTER_SIZE_31  = uint8(5) // Filter coefficient of 32
	BME68X_FILTER_SIZE_63  = uint8(6) // Filter coefficient of 64
	BME68X_FILTER_SIZE_127 = uint8(7) // Filter coefficient of 128
)

// ODR/Standby time macros
const (
	BME68X_ODR_0_59_MS = uint8(0) // Standby time of 0.59ms
	BME68X_ODR_62_5_MS = uint8(1) // Standby time of 62.5ms
	BME68X_ODR_125_MS  = uint8(2) // Standby time of 125ms
	BME68X_ODR_250_MS  = uint8(3) // Standby time of 250ms
	BME68X_ODR_500_MS  = uint8(4) // Standby time of 500ms
	BME68X_ODR_1000_MS = uint8(5) // Standby time of 1s
	BME68X_ODR_10_MS   = uint8(6) // Standby time of 10ms
	BME68X_ODR_20_MS   = uint8(7) // Standby time of 20ms
	BME68X_ODR_NONE    = uint8(8) // No standby time
)

// Operating mode macros
const (
	BME68X_SLEEP_MODE      = uint8(0) // Sleep operation mode
	BME68X_FORCED_MODE     = uint8(1) // Forced operation mode
	BME68X_PARALLEL_MODE   = uint8(2) // Parallel operation mode
	BME68X_SEQUENTIAL_MODE = uint8(3) // Sequential operation mode
)

// Coefficient index macros
const (
	BME68X_LEN_COEFF_ALL       = 42 // Length for all coefficients
	BME68X_LEN_COEFF1          = 23 // Length for 1st group of coefficients
	BME68X_LEN_COEFF2          = 14 // Length for 2nd group of coefficients
	BME68X_LEN_COEFF3          = 5  // Length for 3rd group of coefficients
	BME68X_LEN_FIELD           = 17 // Length of the field
	BME68X_LEN_FIELD_OFFSET    = 17 // Length between two fields
	BME68X_LEN_CONFIG          = 5  // Length of the configuration register
	BME68X_LEN_INTERLEAVE_BUFF = 20 // Length of the interleaved buffer
)

// Coefficient index macros
const (
	BME68X_IDX_T2_LSB         = 0  // Coefficient T2 LSB position
	BME68X_IDX_T2_MSB         = 1  // Coefficient T2 MSB position
	BME68X_IDX_T3             = 2  // Coefficient T3 position
	BME68X_IDX_P1_LSB         = 4  // Coefficient P1 LSB position
	BME68X_IDX_P1_MSB         = 5  // Coefficient P1 MSB position
	BME68X_IDX_P2_LSB         = 6  // Coefficient P2 LSB position
	BME68X_IDX_P2_MSB         = 7  // Coefficient P2 MSB position
	BME68X_IDX_P3             = 8  // Coefficient P3 position
	BME68X_IDX_P4_LSB         = 10 // Coefficient P4 LSB position
	BME68X_IDX_P4_MSB         = 11 // Coefficient P4 MSB position
	BME68X_IDX_P5_LSB         = 12 // Coefficient P5 LSB position
	BME68X_IDX_P5_MSB         = 13 // Coefficient P5 MSB position
	BME68X_IDX_P7             = 14 // Coefficient P7 position
	BME68X_IDX_P6             = 15 // Coefficient P6 position
	BME68X_IDX_P8_LSB         = 18 // Coefficient P8 LSB position
	BME68X_IDX_P8_MSB         = 19 // Coefficient P8 MSB position
	BME68X_IDX_P9_LSB         = 20 // Coefficient P9 LSB position
	BME68X_IDX_P9_MSB         = 21 // Coefficient P9 MSB position
	BME68X_IDX_P10            = 22 // Coefficient P10 position
	BME68X_IDX_H2_MSB         = 23 // Coefficient H2 MSB position
	BME68X_IDX_H2_LSB         = 24 // Coefficient H2 LSB position
	BME68X_IDX_H1_LSB         = 24 // Coefficient H1 LSB position
	BME68X_IDX_H1_MSB         = 25 // Coefficient H1 MSB position
	BME68X_IDX_H3             = 26 // Coefficient H3 position
	BME68X_IDX_H4             = 27 // Coefficient H4 position
	BME68X_IDX_H5             = 28 // Coefficient H5 position
	BME68X_IDX_H6             = 29 // Coefficient H6 position
	BME68X_IDX_H7             = 30 // Coefficient H7 position
	BME68X_IDX_T1_LSB         = 31 // Coefficient T1 LSB position
	BME68X_IDX_T1_MSB         = 32 // Coefficient T1 MSB position
	BME68X_IDX_GH2_LSB        = 33 // Coefficient GH2 LSB position
	BME68X_IDX_GH2_MSB        = 34 // Coefficient GH2 MSB position
	BME68X_IDX_GH1            = 35 // Coefficient GH1 position
	BME68X_IDX_GH3            = 36 // Coefficient GH3 position
	BME68X_IDX_RES_HEAT_VAL   = 37 // Coefficient res heat value position
	BME68X_IDX_RES_HEAT_RANGE = 39 // Coefficient res heat range position
	BME68X_IDX_RANGE_SW_ERR   = 41 // Coefficient res heat range position
)

const (
	BME68X_DISABLE_GAS_MEAS  = uint8(0x00) // Disable gas measurement
	BME68X_ENABLE_GAS_MEAS_L = uint8(0x01) // Enable gas measurement low
	BME68X_ENABLE_GAS_MEAS_H = uint8(0x02) // Enable gas measurement high
)

// Heater control macros
const (
	BME68X_ENABLE_HEATER   = uint8(0x00)    // Enable heater
	BME68X_DISABLE_HEATER  = uint8(0x01)    // Disable heater
	BME68X_MIN_TEMPERATURE = int16(0)       // 0 degree Celsius
	BME68X_MAX_TEMPERATURE = int16(6000)    // 60 degree Celsius
	BME68X_MIN_PRESSURE    = uint32(90000)  // 900 hecto Pascals
	BME68X_MAX_PRESSURE    = uint32(110000) // 1100 hecto Pascals
	BME68X_MIN_HUMIDITY    = uint32(20000)  // 20% relative humidity
	BME68X_MAX_HUMIDITY    = uint32(80000)  // 80% relative humidity

	BME68X_HEATR_DUR1       = uint16(1000)
	BME68X_HEATR_DUR2       = uint16(2000)
	BME68X_HEATR_DUR1_DELAY = uint32(1000000)
	BME68X_HEATR_DUR2_DELAY = uint32(2000000)
	BME68X_N_MEAS           = uint8(6)
	BME68X_LOW_TEMP         = uint8(150)
	BME68X_HIGH_TEMP        = uint16(350)
)

// Mask macros
const (
	BME68X_NBCONV_MSK      = uint8(0x0f) // Mask for number of conversions
	BME68X_FILTER_MSK      = uint8(0x1c) // Mask for IIR filter
	BME68X_ODR3_MSK        = uint8(0x80) // Mask for ODR[3]
	BME68X_ODR20_MSK       = uint8(0xe0) // Mask for ODR[2:0]
	BME68X_OST_MSK         = uint8(0xe0) // Mask for temperature oversampling
	BME68X_OSP_MSK         = uint8(0x1c) // Mask for pressure oversampling
	BME68X_OSH_MSK         = uint8(0x07) // Mask for humidity oversampling
	BME68X_HCTRL_MSK       = uint8(0x08) // Mask for heater control
	BME68X_RUN_GAS_MSK     = uint8(0x30) // Mask for run gas
	BME68X_MODE_MSK        = uint8(0x03) // Mask for operation mode
	BME68X_RHRANGE_MSK     = uint8(0x30) // Mask for res heat range
	BME68X_RSERROR_MSK     = uint8(0xf0) // Mask for range switching error
	BME68X_NEW_DATA_MSK    = uint8(0x80) // Mask for new data
	BME68X_GAS_INDEX_MSK   = uint8(0x0f) // Mask for gas index
	BME68X_GAS_RANGE_MSK   = uint8(0x0f) // Mask for gas range
	BME68X_GASM_VALID_MSK  = uint8(0x20) // Mask for gas measurement valid
	BME68X_HEAT_STAB_MSK   = uint8(0x10) // Mask for heater stability
	BME68X_BIT_H1_DATA_MSK = uint8(0x0f) // Mask for the H1 calibration coefficient
)

// Position macros
const (
	BME68X_FILTER_POS  = uint8(2) // Filter bit position
	BME68X_OST_POS     = uint8(5) // Temperature oversampling bit position
	BME68X_OSP_POS     = uint8(2) // Pressure oversampling bit position
	BME68X_ODR3_POS    = uint8(7) // ODR[3] bit position
	BME68X_ODR20_POS   = uint8(5) // ODR[2:0] bit position
	BME68X_RUN_GAS_POS = uint8(4) // Run gas bit position
	BME68X_HCTRL_POS   = uint8(3) // Heater control bit position
)

type BME68X struct {
	device *i2c.Dev

	status     uint8
	heatStable bool
	gasIdx     uint8
	measIdx    uint8

	temperature   int16  // Celsius degrees x 100
	pressure      uint32 // Pressure in Pascal
	humidity      uint32 // Humidity in % relative humidity x 1000
	gasResistance uint32 // GasResistance in 0hms

	calibData     BME68XCalibrationData
	tphSettings   BME68XTPHSettings
	gasSettings   BME68XHeaterConf
	variantId     uint8 // GAS LOW 0 / HIGH 1
	burnInGasData []float32

	Data BME68XAdjustedReadings
}

type BME68XAdjustedReadings struct {
	Temperature   float32
	Humidity      float32
	Pressure      float32
	GasResistance float32
	IAQ           uint16
}

type BME68XCalibrationData struct {
	par_h1 uint16 // Calibration coefficient for the humidity sensor
	par_h2 uint16 // Calibration coefficient for the humidity sensor
	par_h3 int8   // Calibration coefficient for the humidity sensor
	par_h4 int8   // Calibration coefficient for the humidity sensor
	par_h5 int8   // Calibration coefficient for the humidity sensor
	par_h6 uint8  // Calibration coefficient for the humidity sensor
	par_h7 int8   // Calibration coefficient for the humidity sensor

	par_gh1 int8  // Calibration coefficient for the gas sensor
	par_gh2 int16 // Calibration coefficient for the gas sensor
	par_gh3 int8  // Calibration coefficient for the gas sensor

	par_t1 uint16 // Calibration coefficient for the temperature sensor
	par_t2 int16  // Calibration coefficient for the temperature sensor
	par_t3 int8   // Calibration coefficient for the temperature sensor

	par_p1  uint16 // Calibration coefficient for the pressure sensor
	par_p2  int16  // Calibration coefficient for the pressure sensor
	par_p3  int8   // Calibration coefficient for the pressure sensor
	par_p4  int16  // Calibration coefficient for the pressure sensor
	par_p5  int16  // Calibration coefficient for the pressure sensor
	par_p6  int8   // Calibration coefficient for the pressure sensor
	par_p7  int8   // Calibration coefficient for the pressure sensor
	par_p8  int16  // Calibration coefficient for the pressure sensor
	par_p9  int16  // Calibration coefficient for the pressure sensor
	par_p10 uint8  // Calibration coefficient for the pressure sensor

	t_fine int32 // Variable to store the intermediate temperature coefficient

	res_heat_range uint8 // Heater resistance range coefficient
	res_heat_val   int8  // Heater resistance range coefficient

	range_sw_err int8 // Gas resistance range switching error coefficient
}

type BME68XTPHSettings struct {
	osHum  uint8 // Humidity oversampling
	osTemp uint8 // Temperature oversampling
	osPres uint8 // Pressure oversampling
	filter uint8 // Filter coefficient
	odr    uint8 // Standby time between sequential mode measurement profiles.
}

type BME68XHeaterConf struct {
	enable           uint8  // Enable gas measurement
	heatr_temp       uint16 // Store the heater temperature for forced mode degree Celsius
	heatr_dur        uint16 // Store the heating duration for forced mode in milliseconds
	heatr_temp_prof  uint16 // Store the heater temperature profile in degree Celsius
	heatr_dur_prof   uint16 // Store the heating duration profile in milliseconds
	profile_len      uint8  // Variable to store the length of the heating profile
	shared_heatr_dur uint16 // Variable to store heating duration for parallel mode in milliseconds
}

func NewBME68X(device *i2c.Dev) BME68X {
	return BME68X{device: device}
}

func (bme68x *BME68X) Init() {
	if err := bme68x.softReset(); err != nil {
		log.ErrorLog.Printf("could not reset device; %v/n", err)
		return
	}
	if err := bme68x.chipID(); err != nil {
		log.ErrorLog.Printf("could not read chipId; %v/n", err)
		return
	}

	if err := bme68x.setPowerMode(BME68X_SLEEP_MODE); err != nil {
		log.ErrorLog.Printf("could not set to sleep; %v/n", err)
		return
	}

	if err := bme68x.getCalibrationData(); err != nil {
		log.ErrorLog.Printf("could not retrieve calibrationData; %v/n", err)
		return
	}

	if err := bme68x.setHumidityOversample(BME68X_OS_2X); err != nil {
		log.ErrorLog.Printf("could not set humidity oversample; %v/n", err)
		return
	}

	if err := bme68x.setPressureOversample(BME68X_OS_4X); err != nil {
		log.ErrorLog.Printf("could not set pressure oversample; %v/n", err)
		return
	}

	if err := bme68x.setTemperatureOversample(BME68X_OS_8X); err != nil {
		log.ErrorLog.Printf("could not set temperature oversample; %v/n", err)
		return
	}
	if err := bme68x.setFilter(BME68X_FILTER_SIZE_3); err != nil {
		log.ErrorLog.Printf("could not set gas filter; %v/n", err)
		return
	}

	if err := bme68x.setGasStatus(); err != nil {
		log.ErrorLog.Printf("could not enable gas heater; %v/n", err)
		return
	}

	if err := bme68x.SetGasHeaterTemperature(320); err != nil {
		log.ErrorLog.Printf("could not set gas heater temperature; %v/n", err)
		return
	}
	if err := bme68x.SetGasHeaterDuration(150); err != nil {
		log.ErrorLog.Printf("could not set gas heater duration; %v/n", err)
		return
	}
	if err := bme68x.SetGasHeaterProfile(0); err != nil {
		log.ErrorLog.Printf("could not set gas heater profile; %v/n", err)
		return
	}
}

func (bme68x *BME68X) chipID() error {
	chipID := make([]byte, 1)
	if err := bme68x.device.Tx([]byte{BME68X_REG_CHIP_ID}, chipID); err != nil {
		return err
	}

	if chipID[0] != BME68X_CHIP_ID {
		log.ErrorLog.Fatalf("Failed to find BME68X! Chip ID %v", chipID[0])
	}

	variantId := make([]byte, 1)
	if err := bme68x.device.Tx([]byte{BME68X_REG_VARIANT_ID}, variantId); err != nil {
		return err
	}
	bme68x.variantId = variantId[0]

	return nil
}

func (bme68x *BME68X) getCalibrationData() error {
	coefficients := make([]byte, BME68X_LEN_COEFF_ALL)

	if err := bme68x.device.Tx([]byte{BME68X_REG_COEFF1}, coefficients[0:BME68X_LEN_COEFF1+1]); err != nil {
		return err
	}
	if err := bme68x.device.Tx([]byte{BME68X_REG_COEFF2}, coefficients[BME68X_LEN_COEFF1+1:BME68X_LEN_COEFF1+BME68X_LEN_COEFF2+1]); err != nil {
		return err
	}
	if err := bme68x.device.Tx([]byte{BME68X_REG_COEFF3}, coefficients[BME68X_LEN_COEFF1+BME68X_LEN_COEFF2+1:BME68X_LEN_COEFF_ALL]); err != nil {
		return err
	}

	/* Temperature related coefficients */
	bme68x.calibData.par_t1 = uint16(coefficients[BME68X_IDX_T1_MSB])<<8 | uint16(coefficients[BME68X_IDX_T1_LSB])
	bme68x.calibData.par_t2 = int16(uint16(coefficients[BME68X_IDX_T2_MSB])<<8 | uint16(coefficients[BME68X_IDX_T2_LSB]))
	bme68x.calibData.par_t3 = int8(coefficients[BME68X_IDX_T3])

	/* Pressure related coefficients */
	bme68x.calibData.par_p1 = uint16(coefficients[BME68X_IDX_P1_MSB])<<8 | uint16(coefficients[BME68X_IDX_P1_LSB])
	bme68x.calibData.par_p2 = int16(uint16(coefficients[BME68X_IDX_P2_MSB])<<8 | uint16(coefficients[BME68X_IDX_P2_LSB]))
	bme68x.calibData.par_p3 = int8(coefficients[BME68X_IDX_P3])
	bme68x.calibData.par_p4 = int16(uint16(coefficients[BME68X_IDX_P4_MSB])<<8 | uint16(coefficients[BME68X_IDX_P4_LSB]))
	bme68x.calibData.par_p5 = int16(uint16(coefficients[BME68X_IDX_P5_MSB])<<8 | uint16(coefficients[BME68X_IDX_P5_LSB]))
	bme68x.calibData.par_p6 = int8(coefficients[BME68X_IDX_P6])
	bme68x.calibData.par_p7 = int8(coefficients[BME68X_IDX_P7])
	bme68x.calibData.par_p8 = int16(uint16(coefficients[BME68X_IDX_P8_MSB])<<8 | uint16(coefficients[BME68X_IDX_P8_LSB]))
	bme68x.calibData.par_p9 = int16(uint16(coefficients[BME68X_IDX_P9_MSB])<<8 | uint16(coefficients[BME68X_IDX_P9_LSB]))
	bme68x.calibData.par_p10 = coefficients[BME68X_IDX_P10]

	/* Humidity related coefficients */
	bme68x.calibData.par_h1 = uint16(coefficients[BME68X_IDX_H1_MSB])<<4 | uint16(coefficients[BME68X_IDX_H1_LSB]&BME68X_BIT_H1_DATA_MSK)
	bme68x.calibData.par_h2 = uint16(coefficients[BME68X_IDX_H2_MSB])<<4 | uint16(coefficients[BME68X_IDX_H2_LSB]>>4)
	bme68x.calibData.par_h3 = int8(coefficients[BME68X_IDX_H3])
	bme68x.calibData.par_h4 = int8(coefficients[BME68X_IDX_H4])
	bme68x.calibData.par_h5 = int8(coefficients[BME68X_IDX_H5])
	bme68x.calibData.par_h6 = coefficients[BME68X_IDX_H6]
	bme68x.calibData.par_h7 = int8(coefficients[BME68X_IDX_H7])

	/* Gas heater related coefficients */
	bme68x.calibData.par_gh1 = int8(coefficients[BME68X_IDX_GH1])
	bme68x.calibData.par_gh2 = int16(uint16(coefficients[BME68X_IDX_GH2_MSB])<<8 | uint16(coefficients[BME68X_IDX_GH2_LSB]))
	bme68x.calibData.par_gh3 = int8(coefficients[BME68X_IDX_GH3])

	/* Other coefficients */
	bme68x.calibData.res_heat_range = (coefficients[BME68X_IDX_RES_HEAT_RANGE] & BME68X_RHRANGE_MSK) / 16
	bme68x.calibData.res_heat_val = int8(coefficients[BME68X_IDX_RES_HEAT_VAL])
	bme68x.calibData.range_sw_err = int8(coefficients[BME68X_IDX_RANGE_SW_ERR]&BME68X_RSERROR_MSK) / 16

	return nil
}

func bytesToWord(msb, lsb uint8) uint16 {
	return uint16(msb)<<8 | uint16(lsb)
}

func (bme68x *BME68X) getOperatingMode() error {
	response := make([]byte, 1)
	if err := bme68x.device.Tx([]byte{BME68X_REG_CTRL_MEAS}, response); err != nil {
		return err
	}

	bme68x.status = response[0] & BME68X_MODE_MSK
	return nil
}

func (bme68x *BME68X) softReset() error {
	message := make([]byte, 2)
	message[0] = BME68X_REG_SOFT_RESET
	message[1] = BME68X_SOFT_RESET_CMD
	return bme68x.device.Tx(message, []byte{})
}

func (bme68x *BME68X) setBits(register uint8, mask uint8, pos uint8, value uint8) error {
	temp, err := bme68x.readRegs(register, 1)
	if err != nil {
		return fmt.Errorf("failed to read bits %v %v", register, err)
	}
	temp[0] &= ^mask
	temp[0] |= value << pos
	if err = bme68x.setRegs([]byte{register, temp[0]}); err != nil {
		return fmt.Errorf("failed to set bits %v %v", register, err)
	}
	return nil
}

func (bme68x *BME68X) addGasData() {
	bme68x.burnInGasData = append(bme68x.burnInGasData, float32(bme68x.gasResistance))
	if len(bme68x.burnInGasData) > 50 {
		bme68x.burnInGasData = bme68x.burnInGasData[1:]
	}
}

func (bme68x *BME68X) readRegs(reg uint8, length uint8) ([]byte, error) {
	response := make([]byte, length)
	if err := bme68x.device.Tx([]byte{reg}, response); err != nil {
		return response, fmt.Errorf("failed to read registry %v %v", reg, err)
	}
	return response, nil
}

func (bme68x *BME68X) setRegs(b []byte) error {
	if err := bme68x.device.Tx(b, []byte{}); err != nil {
		return fmt.Errorf("failed to write command %v %v", b[0], err)
	}
	return nil
}

func (bme68x *BME68X) setHumidityOversample(oversample uint8) error {
	bme68x.tphSettings.osHum = oversample
	return bme68x.setBits(BME68X_REG_CTRL_HUM, BME68X_OSH_MSK, 0, oversample)
}

func (bme68x *BME68X) getHumidityOversample() (uint8, error) {
	data, err := bme68x.readRegs(BME68X_REG_CTRL_HUM, 1)
	if err != nil {
		return data[0], err
	}
	return (data[0] & BME68X_OSH_MSK) >> 0, nil
}

func (bme68x *BME68X) setPressureOversample(oversample uint8) error {
	bme68x.tphSettings.osPres = oversample
	return bme68x.setBits(BME68X_REG_CTRL_MEAS, BME68X_OSP_MSK, BME68X_OSP_POS, oversample)
}

func (bme68x *BME68X) getPressureOversample() (uint8, error) {
	data, err := bme68x.readRegs(BME68X_REG_CTRL_MEAS, 1)
	if err != nil {
		return data[0], err
	}
	return (data[0] & BME68X_OSP_MSK) >> BME68X_OSP_POS, nil
}

func (bme68x *BME68X) setTemperatureOversample(oversample uint8) error {
	bme68x.tphSettings.osTemp = oversample
	return bme68x.setBits(BME68X_REG_CTRL_MEAS, BME68X_OST_MSK, BME68X_OST_POS, oversample)
}

func (bme68x *BME68X) getTemperatureOversample() (uint8, error) {
	data, err := bme68x.readRegs(BME68X_REG_CTRL_MEAS, 1)
	if err != nil {
		return data[0], err
	}
	return (data[0] & BME68X_OST_MSK) >> BME68X_OST_POS, nil
}

func (bme68x *BME68X) setFilter(size uint8) error {
	bme68x.tphSettings.filter = size
	return bme68x.setBits(BME68X_REG_CONFIG, BME68X_FILTER_MSK, BME68X_FILTER_POS, size)
}

func (bme68x *BME68X) getFilter() (uint8, error) {
	data, err := bme68x.readRegs(BME68X_REG_CONFIG, 1)
	if err != nil {
		return data[0], err
	}
	return (data[0] & BME68X_FILTER_MSK) >> BME68X_FILTER_POS, nil
}

func (bme68x *BME68X) setGasStatus() error {
	if bme68x.variantId == BME68X_VARIANT_HIGH {
		bme68x.gasSettings.enable = BME68X_ENABLE_GAS_MEAS_H
	} else {
		bme68x.gasSettings.enable = BME68X_ENABLE_GAS_MEAS_L
	}
	return bme68x.setBits(BME68X_REG_CTRL_GAS_1, BME68X_RUN_GAS_MSK, BME68X_RUN_GAS_POS, bme68x.gasSettings.enable)
}

func (bme68x *BME68X) getGasStatus() (uint8, error) {
	data, err := bme68x.readRegs(BME68X_REG_CTRL_GAS_1, 1)
	if err != nil {
		return data[0], err
	}
	return (data[0] & BME68X_RUN_GAS_MSK) >> BME68X_RUN_GAS_POS, nil
}

func (bme680 *BME68X) setPowerMode(mode uint8) error {
	if err := bme680.getPowerMode(); err != nil {
		return err
	}
	if bme680.status != mode {
		if err := bme680.setBits(BME68X_REG_CTRL_MEAS, BME68X_MODE_MSK, 0, mode); err != nil {
			return err
		}

		for {
			log.InfoLog.Printf("Desired %b actual %b", mode, bme680.status)
			if err := bme680.getPowerMode(); err != nil {
				return err
			}
			if bme680.status != mode {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			break
		}
	}
	return nil
}

func (bme680 *BME68X) getPowerMode() error {
	data, err := bme680.readRegs(BME68X_REG_CTRL_MEAS, 1)
	if err != nil {
		return err
	}
	bme680.status = (data[0] & BME68X_MODE_MSK)
	return nil
}

func (bme68x *BME68X) GetSensorData() {
	if err := bme68x.setPowerMode(BME68X_FORCED_MODE); err != nil {
		log.ErrorLog.Printf("Could not toggle to forced mode; %v\n", err)
		return
	}

	for i := 0; i < 10; i++ {
		status, err := bme68x.readRegs(BME68X_REG_FIELD0, 1)
		log.InfoLog.Printf("Status %b", status[0])
		if err != nil {
			log.ErrorLog.Printf("Could not read status; %v\n", err)
			return
		}

		if (status[0] & BME68X_NEW_DATA_MSK) == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		regs, err := bme68x.readRegs(BME68X_REG_FIELD0, BME68X_LEN_FIELD)
		if err != nil {
			log.ErrorLog.Printf("Could not read data; %+v\n", err)
			return
		}

		bme68x.status = regs[0] & BME68X_NEW_DATA_MSK
		bme68x.gasIdx = regs[0] & BME68X_GAS_INDEX_MSK
		bme68x.measIdx = regs[1]

		log.InfoLog.Println("Calibration data %v", bme68x.calibData)

		adc_pres := (uint32(regs[2]) * 4096) | (uint32(regs[3]) * 16) | (uint32(regs[4]) / 16)
		adc_temp := (uint32(regs[5]) * 4096) | (uint32(regs[6]) * 16) | (uint32(regs[7]) / 16)
		adc_hum := uint16((uint32(regs[8]) * 256) | uint32(regs[9]))
		adc_gas_res_low := uint16((uint32(regs[13]) * 4) | (uint32(regs[14]) / 64))
		adc_gas_res_high := uint16(uint32((regs[15])*4) | (uint32(regs[16]) / 64))
		gas_range_l := regs[14] & BME68X_GAS_RANGE_MSK
		gas_range_h := regs[16] & BME68X_GAS_RANGE_MSK

		if bme68x.variantId == BME68X_VARIANT_HIGH {
			bme68x.status |= regs[16] & BME68X_GASM_VALID_MSK
			bme68x.status |= regs[16] & BME68X_HEAT_STAB_MSK
		} else {
			bme68x.status |= regs[14] & BME68X_GASM_VALID_MSK
			bme68x.status |= regs[14] & BME68X_HEAT_STAB_MSK
		}

		bme68x.heatStable = (bme68x.status & BME68X_HEAT_STAB_MSK) > 0
		bme68x.temperature = bme68x.computeTemperature(adc_temp)
		bme68x.pressure = bme68x.computePressure(adc_pres)
		bme68x.humidity = bme68x.computeHumidity(adc_hum)
		if bme68x.variantId == BME68X_VARIANT_HIGH {
			bme68x.gasResistance = bme68x.calcGasResistanceHigh(adc_gas_res_high, gas_range_h)
		} else {
			bme68x.gasResistance = bme68x.calcGasResistanceLow(adc_gas_res_low, gas_range_l)
		}
		bme68x.addGasData()

		bme68x.Data.Temperature = float32(bme68x.temperature) / 100.0
		bme68x.Data.Humidity = float32(bme68x.humidity) / 1000.0
		bme68x.Data.Pressure = float32(bme68x.pressure) / 100.0
		bme68x.Data.GasResistance = float32(bme68x.gasResistance)

		log.InfoLog.Printf("%+v - metrics", bme68x.Data)
		log.InfoLog.Printf("%+v %+v %+v %+v %+v- metrics", bme68x.temperature, bme68x.humidity, bme68x.pressure, bme68x.gasResistance, bme68x.burnInGasData)
	}
}

func sumFLOAT32(array []float32) float32 {
	var result float32
	for _, v := range array {
		result += v
	}
	return result
}

func (bme68x *BME68X) computeAQI() {
	gas_baseline := sumFLOAT32(bme68x.burnInGasData) / 50.0
	hum_baseline := 40.0
	hum_weighting := 0.25

	if bme68x.heatStable {
		gas_offset := gas_baseline - float32(bme68x.gasResistance)
		hum_offset := float64(bme68x.Data.Humidity) - hum_baseline

		var hum_score float64
		if hum_offset > 0 {
			hum_score = (100 - hum_baseline - hum_offset)
			hum_score /= (100 - hum_baseline)
			hum_score *= (hum_weighting * 100)

		} else {
			hum_score = (hum_baseline + hum_offset)
			hum_score /= hum_baseline
			hum_score *= (hum_weighting * 100)
		}

		var gas_score float64
		if gas_offset > 0 {
			gas_score = (float64(bme68x.gasResistance) / float64(gas_baseline))
			gas_score *= (100 - (hum_weighting * 100))
		} else {
			gas_score = 100 - (hum_weighting * 100)
		}
		bme68x.Data.IAQ = uint16(hum_score + gas_score)
	}
}

// Convert the raw temperature to degrees C using calibration_data
func (bme680 *BME68X) computeTemperature(adc_temp uint32) int16 {
	var var1, var2, var3 int64
	var1 = int64((int32(adc_temp) >> 3) - (int32(bme680.calibData.par_t1) << 1))
	var2 = int64((var1 * int64(bme680.calibData.par_t2)) >> 11)
	var3 = ((var1 >> 1) * (var1 >> 1)) >> 12
	var3 = ((var3) * int64(bme680.calibData.par_t3<<4)) >> 14

	// Save teperature data for pressure calculations
	bme680.calibData.t_fine = int32(var2 + var3)
	return int16(((bme680.calibData.t_fine * 5) + 128) >> 8)
}

/* This value is used to check precedence to multiplication or division
 * in the pressure compensation equation to achieve least loss of precision and
 * avoiding overflows.
 * i.e Comparing value, pres_ovf_check = (1 << 31) >> 1
 */
const pres_ovf_check = int32(0x40000000)

// This internal API is used to calculate the pressure value
func (bme680 *BME68X) computePressure(adc_pressure uint32) uint32 {
	var var1, var2, var3, pressureComp int32

	var1 = ((bme680.calibData.t_fine) >> 1) - 64000
	var2 = ((((var1 >> 2) * (var1 >> 2)) >> 11) * int32(bme680.calibData.par_p6)) >> 2
	var2 = var2 + ((var1 * int32(bme680.calibData.par_p5)) << 1)
	var2 = (var2 >> 2) + (int32(bme680.calibData.par_p4) << 16)
	var1 = (((((var1 >> 2) * (var1 >> 2)) >> 13) * (int32(bme680.calibData.par_p3) << 5)) >> 3) +
		((int32(bme680.calibData.par_p2) * var1) >> 1)
	var1 = var1 >> 18
	var1 = ((32768 + var1) * int32(bme680.calibData.par_p1)) >> 15
	pressureComp = 1048576 - int32(adc_pressure)
	pressureComp = int32((pressureComp - (var2 >> 12)) * int32(3125))
	if pressureComp >= pres_ovf_check {
		pressureComp = ((pressureComp / var1) << 1)
	} else {
		pressureComp = ((pressureComp << 1) / var1)
	}

	var1 = (int32(bme680.calibData.par_p9) * int32(((pressureComp>>3)*(pressureComp>>3))>>13)) >> 12
	var2 = (int32(pressureComp>>2) * int32(bme680.calibData.par_p8)) >> 13
	var3 = (int32(pressureComp>>8) * int32(pressureComp>>8) * int32(pressureComp>>8) * int32(bme680.calibData.par_p10)) >> 17
	pressureComp = int32(pressureComp) + ((var1 + var2 + var3 + (int32(bme680.calibData.par_p7) << 7)) >> 4)
	return uint32(pressureComp)
}

func (bme680 *BME68X) computeHumidity(adc_humidity uint16) uint32 {
	var var1, var2, var3, var4, var5, var6, tempScaled, calcHum int32

	tempScaled = ((int32(bme680.calibData.t_fine) * 5) + 128) >> 8
	var1 = int32(int32(adc_humidity)-(int32(int32(bme680.calibData.par_h1)*16))) -
		(((tempScaled * int32(bme680.calibData.par_h3)) / int32(100)) >> 1)
	var2 = (int32(bme680.calibData.par_h2) *
		(((tempScaled * int32(bme680.calibData.par_h4)) / int32(100)) +
			(((tempScaled * ((tempScaled * int32(bme680.calibData.par_h5)) / int32(100))) >> 6) / int32(100)) +
			int32(1<<14))) >> 10
	var3 = var1 * var2
	var4 = int32(bme680.calibData.par_h6) << 7
	var4 = ((var4) + ((tempScaled * int32(bme680.calibData.par_h7)) / int32(100))) >> 4
	var5 = ((var3 >> 14) * (var3 >> 14)) >> 10
	var6 = (var4 * var5) >> 1
	calcHum = (((var3 + var6) >> 10) * int32(1000)) >> 12
	// Cap at 100%rH
	if calcHum > 100000 {
		calcHum = 100000
	} else if calcHum < 0 {
		calcHum = 0
	}

	return uint32(calcHum)
}

var (
	lookupTable1 = []uint32{2147483647, 2147483647, 2147483647, 2147483647,
		2147483647, 2126008810, 2147483647, 2130303777, 2147483647,
		2147483647, 2143188679, 2136746228, 2147483647, 2126008810,
		2147483647, 2147483647}

	lookupTable2 = []uint32{4096000000, 2048000000, 1024000000, 512000000,
		255744255, 127110228, 64000000, 32258064,
		16016016, 8000000, 4000000, 2000000,
		1000000, 500000, 250000, 125000}
)

func (bme68x *BME68X) calcGasResistanceLow(gas_res_adc uint16, gas_range uint8) uint32 {
	var1 := ((1340 + (5 * int64(bme68x.calibData.range_sw_err))) * int64(lookupTable1[gas_range])) >> 16
	var2 := uint64((int64(int64(gas_res_adc)<<15) - int64(16777216)) + var1)
	var3 := ((int64(lookupTable2[gas_range]) * int64(var1)) >> 9)
	return uint32((var3 + (int64(var2) >> 1)) / int64(var2))
}

func (bme68x *BME68X) calcGasResistanceHigh(gas_res_adc uint16, gas_range uint8) uint32 {
	var1 := uint32(262144) >> gas_range
	var2 := int32(gas_res_adc) - int32(512)

	var2 *= int32(3)
	var2 = int32(4096) + var2

	// multiplying 10000 then dividing then multiplying by 100 instead of multiplying by 1000000 to prevent overflow
	calc_gas_res := (uint32(10000) * var1) / uint32(var2)
	calc_gas_res = calc_gas_res * 100

	return calc_gas_res
}

func (bme68x *BME68X) SetGasHeaterTemperature(temperature uint16) error {
	bme68x.gasSettings.heatr_temp = temperature
	temp := bme68x.calcHeaterTemperature(bme68x.gasSettings.heatr_temp)
	return bme68x.setRegs([]byte{BME68X_REG_RES_HEAT0, temp})
}

// Convert raw heater resistance using calibration data
func (bme68x *BME68X) calcHeaterTemperature(temp uint16) uint8 {
	var1 := ((int32(bme68x.temperature) * int32(bme68x.calibData.par_gh3)) / 1000) * 256
	var2 := (int32(bme68x.calibData.par_gh1) + 784) * (((((int32(bme68x.calibData.par_gh2) + 154009) * int32(temp) * 5) / 100) + 3276800) / 10)
	var3 := var1 + (var2 / 2)
	var4 := (var3 / (int32(bme68x.calibData.res_heat_range) + 4))
	var5 := (131 * int32(bme68x.calibData.res_heat_val)) + 65536
	heatr_res_x100 := int32(((var4 / var5) - 250) * 34)
	return uint8((heatr_res_x100 + 50) / 100)
}

func (bme68x *BME68X) SetGasHeaterDuration(duration uint16) error {
	bme68x.gasSettings.heatr_dur = duration
	temp := bme68x.calcHeaterDuration(bme68x.gasSettings.heatr_dur)
	return bme68x.setRegs([]byte{BME68X_REG_GAS_WAIT0, temp})
}

func (bme68x *BME68X) calcHeaterDuration(duration uint16) uint8 {
	var factor, durval uint8

	if duration >= 0xfc0 {
		durval = 0xff // Max duration
	} else {
		for duration > 0x3F {
			duration = duration / 4
			factor += 1
		}
		durval = uint8(uint8(duration) + (factor * 64))
	}

	return durval
}

func (bme68x *BME68X) getGasHeaterProfile() (uint8, error) {
	data, err := bme68x.readRegs(BME68X_REG_CTRL_GAS_1, 1)
	if err != nil {
		return data[0], err
	}

	return data[0] & BME68X_NBCONV_MSK, nil
}

func (bme68x *BME68X) SetGasHeaterProfile(profile uint8) error {
	return bme68x.setBits(BME68X_REG_CTRL_GAS_1, BME68X_NBCONV_MSK, 0, profile)
}
