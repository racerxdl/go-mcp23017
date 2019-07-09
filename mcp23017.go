/*
    MCP23017 I2C Library based on Adafruit Arduino Driver
    https://github.com/adafruit/Adafruit-MCP23017-Arduino-Library/blob/master/Adafruit_MCP23017.cpp
 */
package mcp23017

import (
    "fmt"
    "github.com/d2r2/go-i2c"
)

const (
    _Address  = 0x20
    _IODIRA   = 0x00
    _IPOLA    = 0x02
    _GPINTENA = 0x04
    _DEFVALA  = 0x06
    _INTCONA  = 0x08
    _IOCONA   = 0x0A
    _GPPUA    = 0x0C
    _INTFA    = 0x0E
    _INTCAPA  = 0x10
    _GPIOA    = 0x12
    _OLATA    = 0x14
    _IODIRB   = 0x01
    _IPOLB    = 0x03
    _GPINTENB = 0x05
    _DEFVALB  = 0x07
    _INTCONB  = 0x09
    _IOCONB   = 0x0B
    _GPPUB    = 0x0D
    _INTFB    = 0x0F
    _INTCAPB  = 0x11
    _GPIOB    = 0x13
    _OLATB    = 0x15
    INTERR    = 0xFF
)

type PinMode uint8
type PinLevel bool
type DevicePort uint8

const (
    INPUT  PinMode = 0
    OUTPUT         = 1
)

const (
    PORTA DevicePort = 0
    PORTB            = 1
)

const (
    LOW  PinLevel = false
    HIGH          = true
)

type Device struct {
    dev *i2c.I2C
}

// Open opens a MCP23017 device and returns a handler.
// Returns error in case devNum > 8 or if device is not present
func Open(bus, devNum uint8) (*Device, error) {
    if devNum > 8 {
        return nil, fmt.Errorf("only 8 devices are supported on a single I2C bus")
    }

    dev, err := i2c.NewI2C(_Address+devNum, int(bus))

    if err != nil {
        return nil, err
    }

    err = dev.WriteRegU8(_IODIRA, 0xFF)

    if err != nil {
        dev.Close()
        return nil, fmt.Errorf("error setting defaults to device: %s", err)
    }

    err = dev.WriteRegU8(_IODIRB, 0xFF)

    if err != nil {
        dev.Close()
        return nil, fmt.Errorf("error setting defaults to device: %s", err)
    }

    return &Device{
        dev: dev,
    }, nil
}

// PinMode sets the specified pin (0-15 range) mode
func (d *Device) PinMode(pin uint8, mode PinMode) error {
    v := uint8(0)
    if mode == INPUT {
        v = 1
    }
    return d.updateRegisterBit(pin, v, _IODIRA, _IODIRB)
}

// DigitalWrite sets pin (0-15 range) to specified level
func (d *Device) DigitalWrite(pin uint8, level PinLevel) error {
    v := uint8(0)
    if level == HIGH {
        v = 1
    }

    bit := bitForPin(pin)
    addr := regForPin(pin, _OLATA, _OLATB)

    // Read Current State
    gpio, err := d.dev.ReadRegU8(addr)

    if err != nil {
        return err
    }

    // Set bit
    gpio = bitWrite(gpio, bit, v)

    // Write GPIO
    addr = regForPin(pin, _GPIOA, _GPIOB)
    return d.dev.WriteRegU8(addr, gpio)
}

// DigitalRead returns the level of the specified pin (0-15 range)
func (d *Device) DigitalRead(pin uint8) (PinLevel, error) {
    bit := bitForPin(pin)
    addr := regForPin(pin, _GPIOA, _GPIOB)

    v, err := d.dev.ReadRegU8(addr)

    if err != nil {
        return LOW, err
    }

    if (v>>bit)&0x1 > 0 {
        return HIGH, nil
    }

    return LOW, nil
}

// SetPullUp enables/disables pull up in specified pin (0-15 range)
func (d *Device) SetPullUp(pin uint8, enabled bool) error {
    v := uint8(0)
    if enabled {
        v = 1
    }

    return d.updateRegisterBit(pin, v, _GPPUA, _GPPUB)
}

// WriteGPIOAB sets GPIO AB to specified value
func (d *Device) WriteGPIOAB(value uint16) error {
    return d.dev.WriteRegU16BE(_GPIOA, value)
}

// ReadGPIOAB reads both PORT A and B and returns a 16 bit value containing AB
func (d *Device) ReadGPIOAB() (uint16, error) {
    return d.dev.ReadRegU16BE(_GPIOA)
}

// ReadGPIO reads the specified port and returns it's value
func (d *Device) ReadGPIO(n DevicePort) (uint8, error) {
    p := uint8(_GPIOA)
    if n == PORTB {
        p = _GPIOB
    }

    return d.dev.ReadRegU8(p)
}

// SetupInterrupts configures the device interrupts for both ports A and B.
// mirroring will OR both INTA / INTB pins
// openDrain will set INT pin to specified value or Open Drain
// polarity will set LOW or HIGH on interrupt
// Default values after reset are: false, false, LOW)
func (d *Device) SetupInterrupts(mirroring, openDrain bool, polarity PinLevel) error {

    p := uint8(0)
    if polarity == HIGH {
        p = 1
    }

    m := uint8(0)
    if mirroring {
        m = 1
    }

    od := uint8(0)
    if openDrain {
        od = 1
    }

    // Configure Port A
    r, err := d.dev.ReadRegU8(_IOCONA)
    if err != nil {
        return err
    }

    r = bitWrite(r, 6, m)
    r = bitWrite(r, 2, od)
    r = bitWrite(r, 1, p)

    err = d.dev.WriteRegU8(_IOCONA, r)
    if err != nil {
        return err
    }

    // Configure Port B
    r, err = d.dev.ReadRegU8(_IOCONB)
    if err != nil {
        return err
    }

    r = bitWrite(r, 6, m)
    r = bitWrite(r, 2, od)
    r = bitWrite(r, 1, p)

    err = d.dev.WriteRegU8(_IOCONB, r)
    if err != nil {
        return err
    }

    return nil
}

// GetLastInterruptPin returns the last pin that triggered a interrupt.
// In case of any error (or no interrupt triggered) returns INTERR
func (d *Device) GetLastInterruptPin() uint8 {
    // Check PortA
    f, err := d.dev.ReadRegU8(_INTFA)
    if err != nil {
        return INTERR
    }

    for i := uint8(0); i < 8; i++ {
        if bitRead(f, i) > 0 {
            return i
        }
    }

    // Check PortB
    f, err = d.dev.ReadRegU8(_INTFB)
    if err != nil {
        return INTERR
    }

    for i := uint8(0); i < 8; i++ {
        if bitRead(f, i) > 0 {
            return i + 8
        }
    }

    return INTERR
}

// GetLastInterruptPinValue returns the level of the pin that triggered a interrupt by last
func (d *Device) GetLastInterruptPinValue() (PinLevel, error) {
    i := d.GetLastInterruptPin()
    if i != INTERR {
        addr := regForPin(i, _INTCAPA, _INTCAPB)
        bit := bitForPin(i)

        v, err := d.dev.ReadRegU8(addr)
        if err != nil {
            return LOW, err
        }

        if (v>>bit)&1 > 0 {
            return HIGH, nil
        }

        return LOW, nil
    }

    return LOW, fmt.Errorf("no interrupt triggered")
}

// Close Closes the device
func (d *Device) Close() error {
    return d.dev.Close()
}

// region Private
func (d *Device) updateRegisterBit(pin, value, portA, portB uint8) error {
    addr := regForPin(pin, portA, portB)
    bit := bitForPin(pin)

    regVal, err := d.dev.ReadRegU8(addr)
    if err != nil {
        return err
    }

    regVal = bitWrite(regVal, bit, value)

    return d.dev.WriteRegU8(addr, regVal)
}
// endregion