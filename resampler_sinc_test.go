package goresampler_test

import (
	"errors"
	"testing"

	"github.com/lehatrutenb/goresampler"
	"github.com/lehatrutenb/goresampler/internal/resampleutils"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"
	"golang.org/x/sync/errgroup"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type resamplerSinc struct {
	inRate    int
	outRate   int
	resampled []int16
	opts      *goresampler.SincResamplerOptions
}

func (resamplerSinc) New(inRate int, outRate int, opts *goresampler.SincResamplerOptions) resamplerSinc {
	return resamplerSinc{inRate, outRate, []int16{}, opts}
}

func (rsm resamplerSinc) Copy() testutils.TestResampler {
	res := new(resamplerSinc)
	*res = rsm.New(rsm.inRate, rsm.outRate, rsm.opts)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}

func (rsm resamplerSinc) String() string {
	return fmt.Sprintf("%d_to_%d_sinc_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerSinc) Resample(inp []int16) error {
	sw, _, err := goresampler.NewResamplerSinc(rsm.inRate, rsm.outRate, rsm.opts)
	if err != nil {
		return err
	}
	err = sw.Resample(inp, rsm.resampled)
	return err
}
func (rsm *resamplerSinc) calcNeedSamplesPerOutAmt(outAmt int) int {
	sr, _, err := goresampler.NewResamplerSinc(rsm.inRate, rsm.outRate, rsm.opts)
	if err != nil {
		panic(err)
	}
	inAmt, resOutAmt := sr.CalcInOutSamplesPerOutAmt(int64(outAmt))
	rsm.resampled = make([]int16, resOutAmt)
	return int(inAmt)
}

func (rsm resamplerSinc) OutLen() int {
	return len(rsm.resampled)
}

func (rsm resamplerSinc) OutRate() int {
	return rsm.outRate
}

func (rsm resamplerSinc) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func (rsm resamplerSinc) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestPassIncorrectParams(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	err := resamplerSinc{}.New(8000, 1600, &goresampler.SincResamplerOptions{WindowSize: 11}).Copy().Resample([]int16{1, 2, 3})
	assert.ErrorIs(t, err, goresampler.ErrGotOddWindowSize)

	err = resamplerSinc{}.New(8000, 1600, &goresampler.SincResamplerOptions{MinInAmt: 5}).Copy().Resample([]int16{1, 2, 3})
	assert.ErrorIs(t, err, goresampler.ErrGotWindowSizeLargerSincMinInAmt)

	acc := float64(1)
	opts := goresampler.SincResamplerOptions{BaseResamplerOptions: goresampler.BaseResamplerOptions{MaxErrRateP: &acc}}.WithIgnoreRsmQSpoiling(false)
	err = resamplerSinc{}.New(8000, 1600, &opts).Copy().Resample([]int16{1, 2, 3})
	assert.ErrorIs(t, err, goresampler.ErrGotTooLargeErrRateP)

	opts = goresampler.SincResamplerOptions{BaseResamplerOptions: goresampler.BaseResamplerOptions{MaxErrRateP: &acc}}
	err = resamplerSinc{}.New(8000, 1600, &opts).Copy().Resample([]int16{1, 2, 3})
	assert.ErrorIs(t, err, goresampler.ErrGotTooLargeErrRateP)
}

func TestSincFindInRatePerErr(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
				lInRate := inRate
				lOutRate := outRate
				lAcc := acc
				assert.NotPanics(t, func() { resampleutils.CalcInAmtPerErrRate(lAcc, lInRate, lOutRate, goresampler.SplineMinInAmt) }, "expected to work without runtime errs")
				inAmt, outAmt, ok := resampleutils.CalcInAmtPerErrRate(lAcc, lInRate, lOutRate, goresampler.SplineMinInAmt)
				assert.True(t, ok, "expected to find correct value for such input")
				assert.Less(t, inAmt, int(1e6))    // just some not so big number
				assert.GreaterOrEqual(t, inAmt, 5) // just some not so small number >= resamplerSinc.minInAmt * 8000/48000
				assert.Less(t, outAmt, int(1e6))
				assert.GreaterOrEqual(t, outAmt, 5)
			}
		}
	}
}

func TestResampleSincDiffErrsNotFall_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	eg := &errgroup.Group{}
	waveDurS := float64(20)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
				opts := testutils.TestOpts{}.NewDefault().NotCalcDuration().NotFailOnHighDurationErr()
				var rsm resamplerSinc
				if acc > goresampler.BaseTimeErrRate {
					rsmOpts := goresampler.SincResamplerOptions{BaseResamplerOptions: goresampler.BaseResamplerOptions{MaxErrRateP: &acc}}.WithIgnoreRsmQSpoiling(true)
					rsm = resamplerSinc{}.New(inRate, outRate, &rsmOpts)
					opts = opts.NotFailOnHighErr() // not fail on error if using sinc resampler with large error rate
				} else {
					rsm = resamplerSinc{}.New(inRate, outRate, &goresampler.SincResamplerOptions{BaseResamplerOptions: goresampler.BaseResamplerOptions{MaxErrRateP: &acc}})
				}
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate)), &rsm, 1, t, opts)
				eg.Go(tObj.Run)
			}
		}
	}
	assert.NoError(t, eg.Wait())
}

func TestResampleSincDiffWindowSzNotFall_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	eg := &errgroup.Group{}
	waveDurS := float64(100)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, ws := range []int{2, 4, 6, 8, 10, 16, 32, 64, 128, 256, 1000, 2000} {
				var rsm resamplerSinc
				if ws > goresampler.SincMinInAmt {
					rsm = resamplerSinc{}.New(inRate, outRate, &goresampler.SincResamplerOptions{WindowSize: ws, MinInAmt: ws})
				} else {
					rsm = resamplerSinc{}.New(inRate, outRate, &goresampler.SincResamplerOptions{WindowSize: ws})
				}
				opts := testutils.TestOpts{}.NewDefault().NotCalcDuration().NotFailOnHighErr()
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt(3*outRate)), &rsm, 1, t, opts) // fixed out amount not to fall on O(n^2) duration
				eg.Go(tObj.Run)
			}
		}
	}
	assert.NoError(t, eg.Wait())
}

func TestResampleSincDiffMinAmtNotFall_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	eg := &errgroup.Group{}
	waveDurS := float64(5)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			ws := goresampler.SincBaseWindowSize
			for _, minAmtAdd := range []int{24 - ws, 25 - ws, 26 - ws, 32 - ws, 64 - ws, 128 - ws, 256 - ws, 1000 - ws} {
				rsm := resamplerSinc{}.New(inRate, outRate, &goresampler.SincResamplerOptions{MinInAmt: ws + minAmtAdd})
				opts := testutils.TestOpts{}.NewDefault().NotCalcDuration().NotFailOnHighErr()
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt(5000)), &rsm, 1, t, opts) // fixed out amount not to fall on O(n^2) duration
				eg.Go(tObj.Run)
			}
		}
	}
	assert.NoError(t, eg.Wait())
}

func TestResampleDiffPrecalcingPolicy_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	eg := &errgroup.Group{}
	waveDurS := float64(20)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, policy := range []goresampler.PrecalcedTablePolicyT{goresampler.PolicyAuto, goresampler.PolicyForcePrecalced, goresampler.PolicyForceRuntime} {
				rsm := resamplerSinc{}.New(inRate, outRate, &goresampler.SincResamplerOptions{PrecalcedTablePolicy: policy})
				opts := testutils.TestOpts{}.NewDefault().NotCalcDuration().NotFailOnHighErr()
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt(outRate*10)), &rsm, 1, t, opts) // fixed out amount not to fall on O(n^2) duration
				eg.Go(tObj.Run)
			}
		}
	}
	assert.NoError(t, eg.Wait())
}

func testOnSinWave(t *testing.T, inRate int, outRate int) {
	waveDurS := float64(30)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	rsm := resamplerSinc{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	errTest := tObj.Run()
	errSave := tObj.Save("rsm_sinc")
	if !assert.NoError(t, errTest, "failed to run resampler") {
		t.Error(errTest)
	}
	if !assert.NoError(t, errSave, "failed to save test results") {
		t.Error(errSave)
	}
}

func TestResampleSinc11025To8000_SinWave(t *testing.T) {
	inRate := 11025
	outRate := 8000
	testOnSinWave(t, inRate, outRate)
}

func TestResampleSinc16000To8000_SinWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	testOnSinWave(t, inRate, outRate)
}

func TestResampleSinc44100To8000_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	testOnSinWave(t, inRate, outRate)
}

func TestResampleSinc48000To8000_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	testOnSinWave(t, inRate, outRate)
}

func TestResampleSinc8000To16000_SinWave(t *testing.T) {
	inRate := 8000
	outRate := 16000
	testOnSinWave(t, inRate, outRate)
}

func TestResampleSinc11025To16000_SinWave(t *testing.T) {
	inRate := 11025
	outRate := 16000
	testOnSinWave(t, inRate, outRate)
}

func TestResampleSinc44100To16000_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	testOnSinWave(t, inRate, outRate)
}

func testOnRealWave(t *testing.T, inRate int, outRate int) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	waveDurS := float64(60)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerSinc{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault().WithCrSF(true).NotFailOnHighErr())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_sinc")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleSinc11025To8000_RealWave(t *testing.T) {
	inRate := 11025
	outRate := 8000
	testOnRealWave(t, inRate, outRate)
}

func TestResampleSinc16000To8000_RealWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	testOnRealWave(t, inRate, outRate)
}

func TestResampleSinc44100To8000_RealWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	testOnRealWave(t, inRate, outRate)
}

func TestResampleSinc48000To8000_RealWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	testOnRealWave(t, inRate, outRate)
}

func TestResampleSinc8000To16000_RealWave(t *testing.T) {
	inRate := 8000
	outRate := 16000
	testOnRealWave(t, inRate, outRate)
}

func TestResampleSinc11025To16000_RealWave(t *testing.T) {
	inRate := 11025
	outRate := 16000
	testOnRealWave(t, inRate, outRate)
}

func TestResampleSinc44100To16000_RealWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	testOnRealWave(t, inRate, outRate)
}
