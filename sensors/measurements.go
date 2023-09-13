package sensors

type Unit string

const (
	Hectopascal             Unit = "Hectopascal"             // hPa
	Celsius                 Unit = "Celsius"                 // C
	Percentage              Unit = "Percentage"              // %
	PartsPerMillion         Unit = "PartsPerMillion"         // ppm
	Ohm                     Unit = "Ohm"                     // ohm
	AirQualityIndex         Unit = "AirQualityIndex"         // AIQ
	Count                   Unit = "Count"                   // Number of instances
	Micrometre              Unit = "Micrometre"              // um
	MicrogramsPerCubicMetre Unit = "MicrogramsPerCubicMetre" // µg/m³
	VOCIndex                Unit = "VOC Index"               // Range 1 - 500
	NOxIndex                Unit = "NOx Index"               // Range 1 - 500
)

type Metadata string

const (
	SensorName            Metadata = "sensor"
	ParticleSize          Metadata = "particleSize"
	ParticleConcentration Metadata = "particleConcentration"
)

type Measurement struct {
	ID          string
	Description string
	Unit        Unit
	Labels      []string
}

var (
	Pressure = Measurement{
		ID:          "room_pressure",
		Description: "Pressure hPa",
		Unit:        Hectopascal,
		Labels:      []string{string(SensorName)},
	}
	Temperature = Measurement{
		ID:          "room_temperature",
		Description: "Ambient temperature in C",
		Unit:        Celsius,
		Labels:      []string{string(SensorName)},
	}
	Humidity = Measurement{
		ID:          "room_humidity",
		Description: "Ambient relative humidity",
		Unit:        Percentage,
		Labels:      []string{string(SensorName)},
	}
	VOC = Measurement{
		ID:          "room_voc",
		Description: "Volatile organic compounds",
		Unit:        VOCIndex,
		Labels:      []string{string(SensorName)},
	}
	NOx = Measurement{
		ID:          "room_nox",
		Description: "Nitric Oxide",
		Unit:        NOxIndex,
		Labels:      []string{string(SensorName)},
	}
	CarbonDioxide = Measurement{
		ID:          "room_co2",
		Description: "CO2 in ppm",
		Unit:        PartsPerMillion,
		Labels:      []string{string(SensorName)},
	}
	AIQ = Measurement{
		ID:          "room_iaq",
		Description: "Indoor Air Quality",
		Unit:        AirQualityIndex,
		Labels:      []string{string(SensorName)},
	}
	GasResistance = Measurement{
		ID:          "room_gasResistance",
		Description: "Gas resistance in Ohm",
		Unit:        Ohm,
		Labels:      []string{string(SensorName)},
	}
	ParticleMatterEnvironmental = Measurement{
		ID:          "room_air_quality_pm_concentration_env",
		Description: "Air quality. PM concentration in environmental units.",
		Unit:        MicrogramsPerCubicMetre,
		Labels:      []string{string(ParticleConcentration), string(SensorName)},
	}
	ParticleMatterStandard = Measurement{
		ID:          "room_air_quality_pm_concentration_standard",
		Description: "Air quality. PM concentration in standard units.",
		Unit:        MicrogramsPerCubicMetre,
		Labels:      []string{string(ParticleConcentration), string(SensorName)},
	}
	ParticleCount = Measurement{
		ID:          "room_air_quality_particles_count",
		Description: "Air quality. Particulate matter per 0.1L air.",
		Unit:        Count,
		Labels:      []string{string(ParticleSize), string(SensorName)},
	}

	Measurements = []Measurement{Pressure, Temperature, Humidity, CarbonDioxide, AIQ, GasResistance,
		ParticleCount, ParticleMatterEnvironmental, ParticleMatterStandard, NOx, VOC}
)

type MeasurementRecording struct {
	Measure  *Measurement
	Value    float64
	Sensor   string
	Metadata map[Metadata]string
}
