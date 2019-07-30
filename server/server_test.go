package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLatencyDistribution(t *testing.T) {
	d := []WeightedLatency{
		WeightedLatency{
			Weight:  1,
			Latency: time.Millisecond * 5,
		},
	}

	l, err := getLatencyFromDistribution(d, 0)
	assert.Equal(t, time.Millisecond*5, l)
	assert.Nil(t, err)

	// Give bogus rand num.
	_, err = getLatencyFromDistribution(d, 100)
	assert.NotNil(t, err)

	// Test more complex distribution
	d = []WeightedLatency{
		WeightedLatency{
			Weight:  1,
			Latency: time.Millisecond * 5,
		},
		WeightedLatency{
			Weight:  1,
			Latency: time.Millisecond * 6,
		},
		WeightedLatency{
			Weight:  20,
			Latency: time.Millisecond * 7,
		},
		WeightedLatency{
			Weight:  100,
			Latency: time.Millisecond * 8,
		},
	}
	l, err = getLatencyFromDistribution(d, 0)
	assert.Equal(t, time.Millisecond*5, l)

	l, err = getLatencyFromDistribution(d, 1)
	assert.Equal(t, time.Millisecond*6, l)

	l, err = getLatencyFromDistribution(d, 2)
	assert.Equal(t, time.Millisecond*7, l)
	l, err = getLatencyFromDistribution(d, 3)
	assert.Equal(t, time.Millisecond*7, l)

	l, err = getLatencyFromDistribution(d, 50)
	assert.Equal(t, time.Millisecond*8, l)
	l, err = getLatencyFromDistribution(d, 55)
	assert.Equal(t, time.Millisecond*8, l)

	l, err = getLatencyFromDistribution(d, 150)
	assert.NotNil(t, err)
}
