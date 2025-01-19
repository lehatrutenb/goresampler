package resample

import (
	"errors"
	testutils "resampler/internal/test_utils"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type resamplerSpline struct {
	inRate    int
	outRate   int
	resampled []int16
}

func (resamplerSpline) New(inRate int, outRate int) resamplerSpline {
	return resamplerSpline{inRate, outRate, []int16{}}
}

func (rsm resamplerSpline) Copy() testutils.TestResampler {
	res := new(resamplerSpline)
	*res = rsm.New(rsm.inRate, rsm.outRate)
	return res
}

func (rsm resamplerSpline) String() string {
	return fmt.Sprintf("%d_to_%d_spline_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerSpline) Resample(inp []int16) error {
	sw := SplineWave{}.New()
	sw.SetInOutWave(inp, rsm.inRate, rsm.outRate)
	sw.ResampleSpline()
	rsm.resampled = sw.GetOutWave()
	return nil
}

func (rsm resamplerSpline) OutLen() int {
	return len(rsm.resampled)
}

func (rsm resamplerSpline) OutRate() int {
	return rsm.outRate
}

func (rsm resamplerSpline) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func TestResampleSpline48To32(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 48000, 32000), testutils.TestResampler(&resamplerSpline{inRate: 48000, outRate: 32000}), 10, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleSpline11To8(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 11000, 8000), testutils.TestResampler(&resamplerSpline{inRate: 11000, outRate: 8000}), 10, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleSplineRealWave0(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outR := 8000
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.RealWave{}.New(0, 11025, &outR, nil), testutils.TestResampler(&resamplerSpline{inRate: 11025, outRate: 8000}), 10, t, testutils.TestOpts{}.New().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

type resamplerFFT struct {
	inRate    int
	outRate   int
	resampled []int16
}

func (resamplerFFT) New(inRate int, outRate int) resamplerFFT {
	return resamplerFFT{inRate, outRate, []int16{}}
}

func (rsm resamplerFFT) Copy() testutils.TestResampler {
	res := new(resamplerFFT)
	*res = rsm.New(rsm.inRate, rsm.outRate)
	return res
}

func (rsm resamplerFFT) String() string {
	return fmt.Sprintf("%d_to_%d_fft_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerFFT) Resample(inp []int16) error {
	fr := FFTResampler{}.New(inp, rsm.inRate, rsm.outRate)
	fr.Resample()
	rsm.resampled = fr.GetOutWave()
	return nil
}

func (rsm resamplerFFT) OutLen() int {
	return len(rsm.resampled)
}

func (rsm resamplerFFT) OutRate() int {
	return rsm.outRate
}

func (rsm resamplerFFT) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func TestResampleFFT8To8(t *testing.T) { // just test that everything counts fine
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 1, 8000, 8000), testutils.TestResampler(&resamplerFFT{inRate: 8000, outRate: 8000}), 1, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT16To8(t *testing.T) { // just test that everything counts fine
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 1, 16000, 8000), testutils.TestResampler(&resamplerFFT{inRate: 16000, outRate: 8000}), 1, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT11025To8(t *testing.T) { // just test that everything counts fine
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 1, 11025, 8000), testutils.TestResampler(&resamplerFFT{inRate: 11025, outRate: 8000}), 1, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
