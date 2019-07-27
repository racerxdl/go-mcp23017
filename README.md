# go-mcp23017

GoLang MCP23017 driver that uses Linux I2C calls

Based on Adafruit Arduino Implementation: https://github.com/adafruit/Adafruit-MCP23017-Arduino-Library/


# Example Usage

This code snippet turns on each I/O sequentially and then turns off.

```go
package main

import (
    "github.com/quan-to/slog"
    "github.com/racerxdl/go-mcp23017"
    "time"
)

var log = slog.Scope("TEST")

func main() {
    d, err := mcp23017.Open(1, 0)
    if err != nil {
        log.Error(err)
    }

    defer d.Close()

    for i := 0; i < 16; i++ {
        err := d.PinMode(uint8(i), mcp23017.OUTPUT)
        if err != nil {
            log.Error(err)
        }
    }

    for {
        log.Info("Turning On")
        for i := 0; i < 16; i++ {
            err = d.DigitalWrite(uint8(i), mcp23017.HIGH)
            if err != nil {
                log.Error(err)
            }
            time.Sleep(time.Millisecond * 100)
        }
        time.Sleep(time.Second)
        
        log.Info("Turning Off")
        for i := 0; i < 16; i++ {
            err = d.DigitalWrite(uint8(i), mcp23017.LOW)
            if err != nil {
                log.Error(err)
            }
            time.Sleep(time.Millisecond * 100)
        }
    }
}

```