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
)

type Measurement struct {
	ID          string
	Description string
	Unit        Unit
}

var (
	Pressure = Measurement{
		ID:          "room_pressure",
		Description: "Pressure hPa",
		Unit:        Hectopascal,
	}
	Temperature = Measurement{
		ID:          "room_temperature",
		Description: "Ambient temperature in C",
		Unit:        Celsius,
	}
	Humidity = Measurement{
		ID:          "room_humidity",
		Description: "Ambient relative humidity",
		Unit:        Percentage,
	}
	CarbonDioxide = Measurement{
		ID:          "room_co2",
		Description: "CO2 in ppm",
		Unit:        PartsPerMillion,
	}
	AIQ = Measurement{
		ID:          "room_iaq",
		Description: "Indoor Air Quality",
		Unit:        AirQualityIndex,
	}
	GasResistance = Measurement{
		ID:          "room_gasResistance",
		Description: "Gas resistance in Ohm",
		Unit:        Ohm,
	}
	ParticleMatterEnvironmental = Measurement{
		ID:          "room_air_quality_pm_concentration_env",
		Description: "Air quality. PM concentration in environmental units.",
		Unit:        MicrogramsPerCubicMetre,
	}
	ParticleMatterStandard = Measurement{
		ID:          "room_air_quality_pm_concentration_standard",
		Description: "Air quality. PM concentration in standard units.",
		Unit:        MicrogramsPerCubicMetre,
	}
	ParticleCount = Measurement{
		ID:          "room_air_quality_particles_count",
		Description: "Air quality. Particulate matter per 0.1L air.",
		Unit:        Count,
	}

	Measurements = []Measurement{Pressure, Temperature, Humidity, CarbonDioxide, AIQ, GasResistance,
		ParticleCount, ParticleMatterEnvironmental, ParticleMatterStandard}
)

type MeasurementRecording struct {
	Measure  *Measurement
	Value    float64
	Sensor   string
	Metadata map[string]string
}
