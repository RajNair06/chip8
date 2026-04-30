package cpu

import (
	"fmt"
	"time"
)

// MetricsReporter reads InstrCount from the CPU and logs IPS + jitter every second.
// Run as a goroutine. Stops when done channel is closed.
func MetricsReporter(c *Chip8, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastCount int64
	var lastFrameTime time.Time
	var jitterSamples []float64

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			current := c.InstrCount.Load()
			ips := current - lastCount
			lastCount = current

			// Frame jitter: measure drift from 60Hz = 16.67ms
			if !lastFrameTime.IsZero() {
				elapsed := t.Sub(lastFrameTime).Milliseconds()
				drift := float64(elapsed) - 1000.0
				jitterSamples = append(jitterSamples, drift)
				if len(jitterSamples) > 10 {
					jitterSamples = jitterSamples[1:]
				}
			}
			lastFrameTime = t

			// Compute avg jitter
			avgJitter := 0.0
			for _, j := range jitterSamples {
				if j < 0 {
					j = -j
				}
				avgJitter += j
			}
			if len(jitterSamples) > 0 {
				avgJitter /= float64(len(jitterSamples))
			}

			status := "✓"
			if ips < 450 || ips > 550 {
				status = "⚠"
			}
			fmt.Printf("[METRICS] IPS: %4d %s  | Jitter: %.1fms\n", ips, status, avgJitter)

			// Store latest IPS for potential UI display
			c.lastIPS = ips
		}
	}
}

// GetIPS returns the last measured instructions per second.
func (c *Chip8) GetIPS() int64 {
	return c.lastIPS
}