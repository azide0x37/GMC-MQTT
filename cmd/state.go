package main

import "math"

func buildState(cpm uint16, voltage float64, fw, model, serial string, uptime int) map[string]interface{} {
	// Calculate uSv/h and mR/h with rounding to 4 decimal places.
	usv := math.Round(float64(cpm)*0.0057*10000) / 10000
	mr := math.Round(float64(cpm)*0.00057*10000) / 10000

	return map[string]interface{}{
		"cpm":     cpm,
		"battery": voltage,
		"version": fw,
		"model":   model,
		"serial":  serial,
		"uptime":  uptime,
		"usv":     usv,
		"mr":      mr,
	}
}
