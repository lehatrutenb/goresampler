package goresampler_test

import (
	"errors"
	"fmt"

	"github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"

	"testing"

	assert "github.com/stretchr/testify/assert"
)

type ResamplerLTest struct {
	inRate    int
	outRate   int
	rsm       goresampler.Resampler
	resampled []int16
}

func (ResamplerLTest) New(inRate, outRate int) *ResamplerLTest {
	rsm, _, err := goresampler.NewResamplerAuto[goresampler.BaseResamplerOptions](inRate, outRate, goresampler.ResamplerConstExprT, nil)
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
	inAmt, resOutAmt := rsm.rsm.CalcInOutSamplesPerOutAmt(int64(outAmt))
	rsm.resampled = make([]int16, resOutAmt)
	return int(inAmt)
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

func testOnSinWaveConstExpr(t *testing.T, inRate, outRate int) {
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

func TestResampleConstExpr11000To8000_SinWave(t *testing.T) {
	inRate := 11000
	outRate := 8000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr16000To8000_SinWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr44000To8000_SinWave(t *testing.T) {
	inRate := 44000
	outRate := 8000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr48000To8000_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr8000To16000_SinWave(t *testing.T) {
	inRate := 8000
	outRate := 16000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr11000To16000_SinWave(t *testing.T) {
	inRate := 11000
	outRate := 16000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr44000To16000_SinWave(t *testing.T) {
	inRate := 44000
	outRate := 16000
	testOnSinWaveConstExpr(t, inRate, outRate)
}

func testOnRealWaveConstExpr(t *testing.T, inRate, outRate int) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := ResamplerLTest{}.New(inRate, outRate)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault().WithCrSF(true).NotFailOnHighErr())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleConstExpr11000To8000_RealWave(t *testing.T) {
	inRate := 11000
	outRate := 8000
	testOnRealWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr16000To8000_RealWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	testOnRealWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr44000To8000_RealWave(t *testing.T) {
	inRate := 44000
	outRate := 8000
	testOnRealWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr48000To8000_RealWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	testOnRealWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr8000To16000_RealWave(t *testing.T) {
	inRate := 8000
	outRate := 16000
	testOnRealWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr11000To16000_RealWave(t *testing.T) {
	inRate := 11000
	outRate := 16000
	testOnRealWaveConstExpr(t, inRate, outRate)
}

func TestResampleConstExpr44000To16000_RealWave(t *testing.T) {
	inRate := 44000
	outRate := 16000
	testOnRealWaveConstExpr(t, inRate, outRate)
}
