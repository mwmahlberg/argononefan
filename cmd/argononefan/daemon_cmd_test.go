package main

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

type fakeReader struct {
	temp float32
	err  error
}

func (f *fakeReader) Celsius() (float32, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.temp, nil
}

type fakeFan struct {
	err   error
	calls []int
}

func (f *fakeFan) SetSpeed(s int) error {
	f.calls = append(f.calls, s)
	return f.err
}

func newTestController(r *fakeReader, f *fakeFan) *fanController {
	th := &thresholds{thresholds: map[float32]int{60: 100, 55: 50, 50: 10}}
	th.GenerateIndex()
	return &fanController{logger: hclog.NewNullLogger(), config: th, hysteresis: 1, fan: f, reader: r, currentSpeed: -1}
}

func TestStepReadFailureKeepsSpeed(t *testing.T) {
	r, f := &fakeReader{temp: 61}, &fakeFan{}
	c := newTestController(r, f)
	assert.NoError(t, c.step())
	assert.Equal(t, 100, c.currentSpeed)

	r.err = errors.New("boom")
	assert.Error(t, c.step())
	assert.Equal(t, 100, c.currentSpeed)
	assert.Equal(t, []int{100}, f.calls)

	r.err, r.temp = nil, 59.5
	assert.NoError(t, c.step())
	assert.Equal(t, 100, c.currentSpeed, "hysteresis still applies after recovery")
	r.temp = 54
	assert.NoError(t, c.step())
	assert.Equal(t, 10, c.currentSpeed)
}

func TestStepWriteFailureRetries(t *testing.T) {
	r, f := &fakeReader{temp: 56}, &fakeFan{err: errors.New("i2c")}
	c := newTestController(r, f)
	assert.Error(t, c.step())
	assert.Equal(t, -1, c.currentSpeed)

	f.err = nil
	assert.NoError(t, c.step())
	assert.Equal(t, 50, c.currentSpeed)
	assert.Equal(t, []int{50, 50}, f.calls)
}
