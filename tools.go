package mcp23017

// bitForPin returns the normalized pin value (pin % 8)
func bitForPin(pin uint8) uint8 {
	return pin % 8
}

// regForPin returns the port address for specified pin in 0-15 range
func regForPin(pin, portA, portB uint8) uint8 {
	if pin < 8 {
		return portA
	}

	return portB
}

// bitRead returns bit "bit" in "value"
func bitRead(value, bit uint8) uint8 {
	return value >> bit & 0x01
}

// bitSet sets bit "bit" in "value"
func bitSet(value, bit uint8) uint8 {
	return value | 1<<bit
}

// bitClear resets bit "bit" in "value"
func bitClear(value, bit uint8) uint8 {
	return value & ^(1 << bit)
}

// bitWrite sets bit "bit" in "value" to "b"
func bitWrite(value, bit, b uint8) uint8 {
	if b > 0 {
		return bitSet(value, bit)
	}

	return bitClear(value, bit)
}
