# Sensor data collector/exporter

Expose sensor data as prometheus metrics or just collect it in a sqlite db.

The supported metrics can be seen in [/sensors/measurements.go](sensors/measurements.go).

## Usage

### Linux

1. Copy and edit the [following config](config.toml) 
2. Run `go-home-sensors ----config.file config.toml`

### NixOS

Just include the dependency in your flake confing and enable the service.

```nix
alex-nt.services.home-sensors = {
    enable = true;
    settings = {
        sensors = {
            bme68x.enable = true;
            scd4x.enable = true;
            sen5x.enable = true;
            pmsa003i.enable = true;
        };
        exporters = {
            prometheus.enable = true;
            sqlite.enable = true;
        };
    };
};
```

### Supported sensors

> The sources below were used **heavily** to create the implementations in this repo.

* Bosch
  * BME68x
    * [Bosch](https://github.com/boschsensortec/BME68x-Sensor-API)
* Sensirion
  * SCD41
    * [Adafruit](https://github.com/adafruit/Adafruit_CircuitPython_SCD4X/blob/main/adafruit_scd4x.py)
    * [aldenero](https://github.com/aldernero/scd4x/blob/main/scd4x.go)
  * SEN5x
    * [Sensirion](https://github.com/Sensirion/raspberry-pi-i2c-sen5x/blob/master/sen5x_i2c.c)
* Plantower
  * PMSA003I
    * [Adafruit](https://github.com/adafruit/Adafruit_CircuitPython_PM25)
