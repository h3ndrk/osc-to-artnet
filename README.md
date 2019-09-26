# osc-to-artnet

This project requests fader values of channels via Open Sound Control (OSC) from e.g. a [Behringer X32](https://www.behringer.com/Categories/Behringer/Mixers/Digital/X32/p/P0ASF) and transmits them via ArtNet DMX packets into a light network. Each channel controls a channel in the DMX universe, which enables to control DMX via e.g. the X32. For the X32 this project uses the OSC message `/ch/%02d/mix/fader` to receive the fader value.

## Build

```bash
go build
```

## Usage

```
Usage: osc-to-artnet OSC_ADDR ARTNET_ADDR ARTNET_UNIVERSE ARTNET_CHANNEL_OFFSET
Example: osc-to-artnet 192.168.1.65:10023 2.0.0.1:6454 0 0
```
