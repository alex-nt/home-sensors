package exporters

import "azuremyst.org/go-home-sensors/sensors"

type Exporter interface {
	Export([]sensors.MeasurementRecording)
}
