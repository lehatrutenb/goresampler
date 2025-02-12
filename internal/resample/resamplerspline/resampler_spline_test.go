package resamplerspline_test

import (
	"errors"
	"resampler/internal/resample/resamplerspline"
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
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}

func (rsm resamplerSpline) String() string {
	return fmt.Sprintf("%d_to_%d_spline_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerSpline) Resample(inp []int16) error {
	sw := resamplerspline.New(rsm.inRate, rsm.outRate)
	sw.Resample(inp, rsm.resampled)
	return nil
}
func (rsm *resamplerSpline) calcNeedSamplesPerOutAmt(outAmt int) int {
	var inAmt int
	sr := resamplerspline.New(rsm.inRate, rsm.outRate)
	inAmt, outAmt = sr.CalcInOutSamplesPerOutAmt(outAmt)
	rsm.resampled = make([]int16, outAmt)
	return inAmt
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

func (rsm resamplerSpline) UnresampledInAmt() int {
	return 0
}

func TestResampleSpline11025To8_SinWave(t *testing.T) {
	inRate := 11025
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline16To8_SinWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline44100To8_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline48To8_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline8To16_SinWave(t *testing.T) {
	inRate := 8000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline11025To16_SinWave(t *testing.T) {
	inRate := 11025
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline4410To16_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline48To16_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
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

func TestResampleSpline44100To16_RealWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleSpline44100To8_RealWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleSpline11025To16_RealWave(t *testing.T) {
	inRate := 11025
	outRate := 16000
	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
func TestResampleSpline11025To8_RealWave(t *testing.T) {
	inRate := 11025
	outRate := 16000
	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSpline{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
