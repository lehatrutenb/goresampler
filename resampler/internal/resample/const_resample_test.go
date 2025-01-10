package resample

import (
	"errors"
	testutils "resampler/internal/test_utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

type resampler48To32 struct {
	resampled []int16
}

func (resampler48To32) Copy() testutils.TestResampler {
	return new(resampler48To32)
}

func (resampler48To32) String() string {
	return "48000_to_32000_resampler"
}

func (rsm *resampler48To32) Resample(inp []int16) error {
	rsm.resampled = Resample48To32(inp)
	return nil
}

func (rsm resampler48To32) OutLen() int {
	return len(rsm.resampled)
}

func (resampler48To32) OutRate() int {
	return 32000
}

func (rsm resampler48To32) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

type resampler48To32L struct {
	resampled []int16
}

func (resampler48To32L) Copy() testutils.TestResampler {
	return new(resampler48To32L)
}

func (resampler48To32L) String() string {
	return "48000_to_32000_resampler_legacy"
}

func (rsm *resampler48To32L) Resample(inp []int16) error {
	rsm.resampled = Resample48To32L(inp)
	return nil
}

func (rsm resampler48To32L) OutLen() int {
	return len(rsm.resampled)
}

func (resampler48To32L) OutRate() int {
	return 32000
}

func (rsm resampler48To32L) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func TestResample48To32(t *testing.T) {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 50, 48000, 32000), testutils.TestResampler(&resampler48To32{}), 10, t)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To32RealWave(t *testing.T) {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.RealWave{}.New(0, 48000), testutils.TestResampler(&resampler48To32{}), 10, t)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To32L(t *testing.T) {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 50, 48000, 32000), testutils.TestResampler(&resampler48To32L{}), 10, t)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
