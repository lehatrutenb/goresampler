package resamplerce_test

import (
	"errors"
	"fmt"

	"github.com/lehatrutenb/go_resampler/internal/resample/resamplerauto"
	resampleri "github.com/lehatrutenb/go_resampler/internal/resample/resampleri"
	testutils "github.com/lehatrutenb/go_resampler/internal/test_utils"

	"testing"

	assert "github.com/stretchr/testify/assert"
)

type ResamplerLTest struct {
	inRate    int
	outRate   int
	rsm       resampleri.Resampler
	resampled []int16
}

func (ResamplerLTest) New(inRate, outRate int) *ResamplerLTest {
	rsm, _, err := resamplerauto.New(inRate, outRate, resamplerauto.ResamplerConstExpr, nil)
	if err != nil {
		panic(err)
	}
	res := new(ResamplerLTest)
	*res = ResamplerLTest{inRate, outRate, rsm, nil}
	return res
}

func (rsm ResamplerLTest) Copy() testutils.TestResampler {
	res := ResamplerLTest{}.New(rsm.inRate, rsm.outRate)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}
func (rsm ResamplerLTest) String() string {
	return fmt.Sprintf("%d_to_%d_resamplerL", rsm.inRate, rsm.outRate)
}
func (rsm *ResamplerLTest) Resample(inp []int16) error { // care moved allocation of output to CalcNeesSamples - logc you can't resample without that
	return rsm.rsm.Resample(inp, rsm.resampled)
}
func (rsm *ResamplerLTest) calcNeedSamplesPerOutAmt(outAmt int) int {
	var inAmt int
	inAmt, outAmt = rsm.rsm.CalcInOutSamplesPerOutAmt(outAmt)
	rsm.resampled = make([]int16, outAmt)
	return inAmt
}
func (rsm ResamplerLTest) OutLen() int {
	return len(rsm.resampled)
}
func (rsm ResamplerLTest) OutRate() int {
	return rsm.outRate
}
func (rsm ResamplerLTest) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}
func (rsm ResamplerLTest) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestResample11To8L_SinWave(t *testing.T) {
	inRate := 11000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample16To8L_SinWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample44000To8L_SinWave(t *testing.T) {
	inRate := 44000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To8L_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample8To16L_SinWave(t *testing.T) {
	inRate := 8000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample11To16L_SinWave(t *testing.T) {
	inRate := 11000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample44000To16L_SinWave(t *testing.T) {
	inRate := 44000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To16L_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 16000
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
