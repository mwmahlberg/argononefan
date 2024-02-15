package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHysteresis(t *testing.T) {
	thresholds := &thresholds{
		thresholds: map[float32]int{
			60: 100,
			55: 50,
			50: 10,
		},
	}
	thresholds.GenerateIndex()
	assert.Equal(t, 100, thresholds.GetSpeed(60))
	assert.Equal(t, 50, thresholds.GetSpeed(55))
	assert.Equal(t, 50, thresholds.GetSpeedWithHysteresis(54.5, 1))
	assert.Equal(t, 10, thresholds.GetSpeedWithHysteresis(54, 1))
	assert.Equal(t, 0, thresholds.GetSpeedWithHysteresis(49, 1))
}
