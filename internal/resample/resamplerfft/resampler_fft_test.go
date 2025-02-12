package resamplerfft_test

import (
	"errors"
	"resampler/internal/resample/resamplerfft"
	testutils "resampler/internal/test_utils"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

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
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}

func (rsm resamplerFFT) String() string {
	return fmt.Sprintf("%d_to_%d_fft_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerFFT) Resample(inp []int16) error {
	fr := resamplerfft.New(rsm.inRate, rsm.outRate)
	fr.Resample(inp, rsm.resampled)
	return nil
}
func (rsm *resamplerFFT) calcNeedSamplesPerOutAmt(outAmt int) int {
	var inAmt int
	fr := resamplerfft.New(rsm.inRate, rsm.outRate)
	inAmt, outAmt = fr.CalcInOutSamplesPerOutAmt(outAmt)
	rsm.resampled = make([]int16, outAmt)
	return inAmt
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

func (rsm resamplerFFT) UnresampledInAmt() int {
	return 0
}

func TestResampleFFT11025To8_SinWave(t *testing.T) {
	inRate := 11025
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT16To8_SinWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT44100To8_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT48To8_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT44100To16_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT48To16_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT11025To8(t *testing.T) { // just test that everything counts fine
	inRate := 11025
	outRate := 8000
	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("rsm_fft")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT44100To16(t *testing.T) { // just test that everything counts fine
	inRate := 48000
	outRate := 16000
	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate)

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, nil)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("rsm_fft")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
