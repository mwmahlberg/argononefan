# ArgonOne fan speed tools and control daemon

## Installation

TBD

## Usage

### Daemon Mode

When run in [daemon][wp:daemon] mode, `argononefan` will read the temperature
of the CPU each `$ARGONONEFAN_CHECK_INTERVAL` and if a threshold is crossed
set the fan speed to the according value for the current threshold.

In the output of the help below (which is also showing the defaults),
this would translate to:

| Temperature | Threshold | Fan Speed       |
| ----------- | --------- | --------------- |
| >60°C       | 60        | 100%            |
| 59°C        | 55        | 50%             |
| 50°C        | 50        | 10%             |
| <50°C       | -         | 0% (fan is off) |

```none
Usage: argononefan daemon

Run the fan control daemon

Flags:
  -h, --help                             Show context-sensitive help.
  -d, --debug                            Enable debug mode ($ARGONONEFAN_DEBUG)
  -f, --device-file="/sys/class/thermal/thermal_zone0/temp"
                                         File path in sysfs containing current
                                         CPU temperature 
                                         ($ARGONONEFAN_DEVICE_FILE)
  -b, --bus=0                            I2C bus the fan resides on ($ARGONONEFAN_BUS)

  -t, --thresholds=60=100;55=50;50=10    Threshold map of °C to fan speed in %
                                         ($ARGONONEFAN_THRESHOLDS)
  -i, --check-interval=5s                Check interval ($ARGONONEFAN_CHECK_INTERVAL)
```

### Read the temperature of the CPU

```none
Usage: argononefan temperature

Read the current CPU temperature

Flags:
  -h, --help        Show context-sensitive help.
  -d, --debug       Enable debug mode ($ARGONONEFAN_DEBUG)
  -f, --device-file="/sys/class/thermal/thermal_zone0/temp"
                    File path in sysfs containing current CPU temperature
                    ($ARGONONEFAN_DEVICE_FILE)
  -b, --bus=0       I2C bus the fan resides on ($ARGONONEFAN_BUS)

  -i, --imperial    Display temperature in imperial system
```

### Set the fan speed statically

You can set the fan speed with `argononefan set-speed <speed>`.
Not that this will only temporarily overwrite the adjustments made when
`argononefan` is also running in daemon mode on the same machine.

```none
Usage: argononefan set-speed <speed>

Set the fan speed manually

Arguments:
  <speed>    Fan speed

Flags:
  -h, --help     Show context-sensitive help.
  -d, --debug    Enable debug mode ($ARGONONEFAN_DEBUG)
  -f, --device-file="/sys/class/thermal/thermal_zone0/temp"
                 File path in sysfs containing current CPU temperature ($ARGONONEFAN_DEVICE_FILE)
  -b, --bus=0    I2C bus the fan resides on ($ARGONONEFAN_BUS)
```

## Thanks

This tool started as a fork of [samonzeweb/argononefan](https://github.com/samonzeweb/argononefan).

I realized pretty soon that I'd have to overhaul it substantially to fit my needs.
It turned out to be true: I have rewritten pretty much every line of code.

Nevertheless, I would like to thank @samonzeweb for the inspiration and the
foundational work. As usual as an open source developer one stands on the
shoulders of giants.

[wp:daemon]: https://en.wikipedia.org/wiki/Daemon_(computing) "Wikipedia page on 'daemon (computing)'"
