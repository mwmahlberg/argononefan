# ArgonOne fan speed tools and control daemon

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/mwmahlberg/argononefan/.github%2Fworkflows%2Fgo.yml?logo=GitHub&color=00aa00)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=mwmahlberg_argononefan&metric=bugs)](https://sonarcloud.io/summary/new_code?id=mwmahlberg_argononefan)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=mwmahlberg_argononefan&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=mwmahlberg_argononefan)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=mwmahlberg_argononefan&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=mwmahlberg_argononefan)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=mwmahlberg_argononefan&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=mwmahlberg_argononefan)

![GitHub Tag](https://img.shields.io/github/v/tag/mwmahlberg/argononefan)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mwmahlberg/argononefan)
[![Godoc](https://godoc.org/github.com/mwmahlberg/argononefan?status.svg)](http://godoc.org/github.com/mwmahlberg/argononefan)

![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-raw/mwmahlberg/argononefan?style=flat)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-pr-raw/mwmahlberg/argononefan)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues-pr-closed-raw/mwmahlberg/argononefan)
![GitHub Sponsors](https://img.shields.io/github/sponsors/mwmahlberg)

## Installation

TBD

## Usage

### Daemon Mode

When run in [daemon][wp:daemon] mode, `argononefan` will read the temperature
of the CPU each `$ARGONONEFAN_CHECK_INTERVAL` and if a threshold is crossed
set the fan speed to the according value for the current threshold.

> ***Note***
>
> A Raspberry Pi 4 is safe to run below 85°C, according to [raspberrytips.com][rpitips:cooling]
> and several other sources.
>
> The default thresholds try to find a balance between convenience (fan off) and
> proper cooling (turning the fan on 100% way before a critical temperature is reached).

In the output of the help below (which is also showing the defaults),
this would translate to:

| Temperature | Threshold | Fan Speed       |
| ----------- | --------- | --------------- |
| >=70°C      | 70        | 100%            |
| 61°C        | 60        | 50%             |
| 58°C        | 55        | 10%             |
| <55°C       | -         | 0% (fan is off) |

With the default hysteresis of 1.0 set, if the temperature crossed a higher threshold,
it must drop 1.0°C below the threshold it is coming from before the fan will be
slowed down.

Say the temperature was 56°C and drops to 54.5°C. Without the hysteresis, the
fan would turn off. However, with the hysteresis, the CPU would first have to cool
down to 54°C or lower for the fan to stop.

Note that as a failsafe measure, the hysteresis is never applied when checking whether
the fan should speed up.

```none
$ argononefan daemon -h
Usage: argononefan daemon

Run the fan control daemon

Flags:
  -h, --help                 Show context-sensitive help.
  -d, --debug                Enable debug mode ($ARGONONEFAN_DEBUG)
  -f, --device-file="/sys/class/thermal/thermal_zone0/temp"
                             File path in sysfs containing current CPU
                             temperature ($ARGONONEFAN_DEVICE_FILE)
  -b, --bus=0                I2C bus the fan resides on ($ARGONONEFAN_BUS)

  -t, --thresholds=70=100;60=50;55=10
                             thresholds is map of °C to fan speed in %
                             ($ARGONONEFAN_THRESHOLDS)
      --hysteresis=1.0       hysteresis is the value in °C the temperature must
                             drop below a threshold before the fan is slowed
                             down to the according speed. This is to prevent the
                             fan from constantly switching between two speeds.

                             Note that this only applies to the fan slowing down
                             coming from a higher threshold, not when speeding
                             up.

                               ($ARGONONEFAN_HYSTERESIS)
  -i, --check-interval=5s    Check interval ($ARGONONEFAN_CHECK_INTERVAL)
      --prometheus-bind="localhost:8080"
                             Address to bind the Prometheus metrics server to
                             ($ARGONONEFAN_PROMETHEUS_BIND)
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
[rpitips:cooling]: https://raspberrytips.com/raspberry-pi-temperature/ "Raspberry Pi Temperature: Limits monitoring, cooling and more"

---

[![SonarCloud](https://sonarcloud.io/images/project_badges/sonarcloud-orange.svg)](https://sonarcloud.io/summary/new_code?id=mwmahlberg_argononefan)