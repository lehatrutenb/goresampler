package goresampler_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"
	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/assert"
)

// another type of tests as in all resamplers - just all in 1 to check that evrything out of base work is fine

var ErrExpectToCallCalcNeedSamplesPerOutAmtBefore = errors.New("error expected to call resamplerAutoTest.CalcNeedSamplesPerOutAmtBefore")

type resamplerAutoTest[optsT goresampler.ResamplerOptionsT] struct {
	inRate    int
	outRate   int
	rsmT      goresampler.ResamplerT
	rsm       goresampler.Resampler
	resampled []int16
	rsmOpts   *optsT
}

func (resamplerAutoTest[optsT]) New(inRate, outRate int, rsmT goresampler.ResamplerT, rsmOpts *optsT) *resamplerAutoTest[optsT] {
	rsm, _, err := goresampler.NewResamplerAuto(inRate, outRate, rsmT, rsmOpts)
	if err != nil {
		panic(err)
	}
	res := new(resamplerAutoTest[optsT])
	*res = resamplerAutoTest[optsT]{inRate, outRate, rsmT, rsm, nil, rsmOpts}
	return res
}

func (rsm resamplerAutoTest[optsT]) Copy() testutils.TestResampler {
	res := resamplerAutoTest[optsT]{}.New(rsm.inRate, rsm.outRate, rsm.rsmT, rsm.rsmOpts)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}
func (rsm resamplerAutoTest[optsT]) String() string {
	return fmt.Sprintf("%d_to_%d_resamplerAuto_%s", rsm.inRate, rsm.outRate, rsm.rsmT)
}
func (rsm *resamplerAutoTest[optsT]) Resample(inp []int16) error { // care moved allocation of output to CalcNeesSamples - logc you can't resample without that
	if rsm.resampled == nil {
		return ErrExpectToCallCalcNeedSamplesPerOutAmtBefore
	}
	return rsm.rsm.Resample(inp, rsm.resampled)
}
func (rsm *resamplerAutoTest[optsT]) calcNeedSamplesPerOutAmt(outAmt int64) int64 {
	var inAmt int64
	inAmt, outAmt = rsm.rsm.CalcInOutSamplesPerOutAmt(outAmt)
	rsm.resampled = make([]int16, outAmt)
	return inAmt
}
func (rsm resamplerAutoTest[optsT]) OutLen() int {
	return len(rsm.resampled)
}
func (rsm resamplerAutoTest[optsT]) OutRate() int {
	return rsm.outRate
}
func (rsm resamplerAutoTest[optsT]) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}
func (rsm resamplerAutoTest[optsT]) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestResampleAuto_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(60)
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerSincT, goresampler.ResamplerFFtT, goresampler.ResamplerBestFitT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := resamplerAutoTest[goresampler.BaseResamplerOptions]{}.New(inRate, outRate, rsmT, nil)
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, int(rsm.calcNeedSamplesPerOutAmt(int64((int(waveDurS)-30)*outRate)))), rsm, 1, t, testutils.TestOpts{}.NewDefault())
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
				err = tObj.Save("rsm_auto")
				if !assert.NoError(t, err, "failed to save test results") {
					t.Error(err)
				}
			}
		}
	}
}

func TestResampleAutoNotDefaultConversions_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(60)
	rsmT := goresampler.ResamplerBestFitNotSafeT
	for _, inRate := range []int{12000, 24000, 22050, 32000, 96000} {
		for _, outRate := range []int{8000, 16000} {
			if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
				continue
			}
			log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
			rsm := resamplerAutoTest[goresampler.BaseResamplerOptions]{}.New(inRate, outRate, rsmT, nil)
			var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, int(rsm.calcNeedSamplesPerOutAmt(int64((int(waveDurS)-30)*outRate)))), rsm, 1, t, testutils.TestOpts{}.NewDefault())
			err := tObj.Run()
			if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
				t.Error(err)
			}
		}
	}
}

func TestResampleAutoNotDefaultConversionsError_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerSincT, goresampler.ResamplerFFtT, goresampler.ResamplerBestFitT} {
		for _, inRate := range []int{12000, 24000, 22050, 32000, 96000} {
			for _, outRate := range []int{8000, 16000} {
				_, _, err := goresampler.NewResamplerAuto[goresampler.BaseResamplerOptions](inRate, outRate, rsmT, nil)
				if !assert.Error(t, err, fmt.Sprintf("expected not to create resampler with config %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
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

	waveDurS := float64(100)
	eg := &errgroup.Group{}
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerSincT, goresampler.ResamplerFFtT, goresampler.ResamplerBestFitT, goresampler.ResamplerBestFitNotSafeT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
					log.Printf("Testing %s from %d to %d with err %v\n", rsmT.String(), inRate, outRate, acc)
					opts := testutils.TestOpts{}.NewDefault().NotCalcDuration().NotFailOnHighDurationErr()

					accOpts := goresampler.BaseResamplerOptions{&acc}
					if (rsmT == goresampler.ResamplerSincT || rsmT == goresampler.ResamplerBestFitT || rsmT == goresampler.ResamplerBestFitNotSafeT) && acc > goresampler.BaseTimeErrRate {
						rsmOpts := goresampler.SincResamplerOptions{BaseResamplerOptions: accOpts}.WithIgnoreRsmQSpoiling(true)
						rsm := resamplerAutoTest[goresampler.SincResamplerOptions]{}.New(inRate, outRate, rsmT, &rsmOpts)
						opts = opts.NotFailOnHighErr() // not fail on error if using sinc resampler with large error rate

						if rsm.calcNeedSamplesPerOutAmt(int64((int(waveDurS)-5)*outRate))-5 >= int64(waveDurS)*int64(inRate) {
							continue
						}
						var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, int(rsm.calcNeedSamplesPerOutAmt(int64(10*outRate)))), rsm, 1, t, opts)
						eg.Go(tObj.Run)
					} else {
						accOptsP := &accOpts
						if rsmT == goresampler.ResamplerConstExprT {
							accOptsP = nil
						}
						rsm := resamplerAutoTest[goresampler.BaseResamplerOptions]{}.New(inRate, outRate, rsmT, accOptsP)
						if rsm.calcNeedSamplesPerOutAmt(int64((int(waveDurS)-5)*outRate))-5 >= int64(waveDurS)*int64(inRate) {
							continue
						}
						var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, int(rsm.calcNeedSamplesPerOutAmt(int64(10*outRate)))), rsm, 1, t, opts)
						eg.Go(tObj.Run)
					}
				}
			}
		}
		assert.NoError(t, eg.Wait())
	}
}

func ExampleResamplerAuto() {
	var err error
	defer func() { _ = err }()

	rsm, _, err := goresampler.NewResamplerAuto[goresampler.SincResamplerOptions](16000, 8000, goresampler.ResamplerBestFitT, nil)
	if err != nil {
		fmt.Printf("failed to initialize resampler with such args; err: %v", err)
		return
	}
	inAmt, outAmt := rsm.CalcInOutSamplesPerOutAmt(10)
	in := make([]int16, inAmt)
	out := make([]int16, outAmt)
	for i := range in {
		in[i] = int16(i)
	}
	rsm.Resample(in, out)
	// Output:
}

func ExampleResamplerAuto_second() {
	var err error
	defer func() { _ = err }()

	rsm, _, err := goresampler.NewResamplerAuto[goresampler.SincResamplerOptions](96000, 8000, goresampler.ResamplerBestFitNotSafeT, nil)
	if err != nil {
		fmt.Printf("failed to initialize resampler with such args; err: %v", err)
		return
	}
	inAmt, outAmt := rsm.CalcInOutSamplesPerOutAmt(10)
	in := make([]int16, inAmt)
	out := make([]int16, outAmt)
	for i := range in {
		in[i] = int16(i)
	}
	rsm.Resample(in, out)
	// Output:
}

// set options for sinc resampler:
//
// windowSize (the more - the better quality + the more time need)
//
// PrecalcedTablePolicy (asked to force precalc table on resampler creation)
func ExampleResamplerAuto_third() {
	var err error
	defer func() { _ = err }()

	rsm, _, err := goresampler.NewResamplerAuto(44100, 16000, goresampler.ResamplerBestFitT, &goresampler.SincResamplerOptions{WindowSize: 16, PrecalcedTablePolicy: goresampler.PolicyForcePrecalced})
	if err != nil {
		fmt.Printf("failed to initialize resampler with such args; err: %v", err)
		return
	}
	inAmt, outAmt := rsm.CalcInOutSamplesPerOutAmt(10)
	in := make([]int16, inAmt)
	out := make([]int16, outAmt)
	for i := range in {
		in[i] = int16(i)
	}
	rsm.Resample(in, out)
	// Output:
}
