package game

import (
	"fmt"
	"sync/atomic"
)

// PinGenerator generates sequential 4-letter PINs.
type PinGenerator struct {
	counter atomic.Int64
}

func NewPinGenerator() *PinGenerator {
	return &PinGenerator{}
}

// Next returns the next PIN in sequence.
func (p *PinGenerator) Next() string {
	n := p.counter.Add(1) - 1
	return intToPin(n)
}

func intToPin(n int64) string {
	chars := make([]byte, 4)
	for i := 3; i >= 0; i-- {
		chars[i] = byte('A' + n%26)
		n /= 26
	}
	return string(chars)
}

// SetCounter seeds the counter (e.g., from DB count on startup).
func (p *PinGenerator) SetCounter(n int64) {
	p.counter.Store(n)
}

// PinToDisplay formats a PIN for display.
func PinToDisplay(pin string) string {
	return fmt.Sprintf("%s", pin)
}
