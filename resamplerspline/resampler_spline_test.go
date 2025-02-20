package resamplerspline_test

import (
	"errors"
	"sync"
	"testing"

	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"

	"github.com/lehatrutenb/goresampler/resamplerspline"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type resamplerSpline struct {
	inRate      int
	outRate     int
	resampled   []int16
	maxErrRateP *float64
}

func (resamplerSpline) New(inRate int, outRate int, maxErrRateP *float64) resamplerSpline {
	return resamplerSpline{inRate, outRate, []int16{}, maxErrRateP}
}

func (rsm resamplerSpline) Copy() testutils.TestResampler {
	res := new(resamplerSpline)
	*res = rsm.New(rsm.inRate, rsm.outRate, rsm.maxErrRateP)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}

func (rsm resamplerSpline) String() string {
	return fmt.Sprintf("%d_to_%d_spline_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerSpline) Resample(inp []int16) error {
	sw, _ := resamplerspline.New(rsm.inRate, rsm.outRate, nil)
	sw.Resample(inp, rsm.resampled)
	return nil
}
func (rsm *resamplerSpline) calcNeedSamplesPerOutAmt(outAmt int) int {
	var inAmt int
	sr, _ := resamplerspline.New(rsm.inRate, rsm.outRate, nil)
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

func (rsm resamplerSpline) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestFindInRatePerErr(t *testing.T) {
	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
				lInRate := inRate
				lOutRate := outRate
				lAcc := acc
				assert.NotPanics(t, func() { resamplerspline.CalcInAmtPerErrRate(lAcc, lInRate, lOutRate) }, "expected to work without runtime errs")
				inAmt, outAmt, ok := resamplerspline.CalcInAmtPerErrRate(lAcc, lInRate, lOutRate)
				assert.True(t, ok, "expected to find correct value for such input")
				assert.Less(t, inAmt, int(1e6))    // just some not so big number
				assert.GreaterOrEqual(t, inAmt, 5) // just some not so small number >= resamplerspline.minInAmt * 8000/48000
				assert.Less(t, outAmt, int(1e6))
				assert.GreaterOrEqual(t, outAmt, 5)
			}
		}
	}
}

func TestResampleAutoDiffErrsNotFall_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	wg := &sync.WaitGroup{}
	waveDurS := float64(20)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			if inRate == outRate {
				continue
			}
			for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
				rsm := resamplerSpline{}.New(inRate, outRate, &acc)
				opts := testutils.TestOpts{}.NewDefault().NotFailOnHighDurationErr().NotCalcDuration().WithWaitGroup(wg)
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate)), &rsm, 1, t, opts)
				wg.Add(1)
				go tObj.Run()
			}
		}
	}
	wg.Wait()

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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
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
	rsm := resamplerSpline{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_spline")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
