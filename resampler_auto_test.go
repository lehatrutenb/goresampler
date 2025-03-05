package goresampler_test

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"

	"github.com/stretchr/testify/assert"
)

// another type of tests as in all resamplers - just all in 1 to check that evrything out of base work is fine

var ErrExpectToCallCalcNeedSamplesPerOutAmtBefore = errors.New("error expected to call resamplerAutoTest.CalcNeedSamplesPerOutAmtBefore")

type resamplerAutoTest struct {
	inRate      int
	outRate     int
	rsmT        goresampler.ResamplerT
	rsm         goresampler.Resampler
	resampled   []int16
	maxErrRateP *float64
}

func (resamplerAutoTest) New(inRate, outRate int, rsmT goresampler.ResamplerT, maxErrRateP *float64) *resamplerAutoTest {
	rsm, _, err := goresampler.NewResamplerAuto(inRate, outRate, rsmT, maxErrRateP)
	if err != nil {
		panic(err)
	}
	res := new(resamplerAutoTest)
	*res = resamplerAutoTest{inRate, outRate, rsmT, rsm, nil, maxErrRateP}
	return res
}

func (rsm resamplerAutoTest) Copy() testutils.TestResampler {
	res := resamplerAutoTest{}.New(rsm.inRate, rsm.outRate, rsm.rsmT, rsm.maxErrRateP)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}
func (rsm resamplerAutoTest) String() string {
	return fmt.Sprintf("%d_to_%d_resamplerAuto_%s", rsm.inRate, rsm.outRate, rsm.rsmT)
}
func (rsm *resamplerAutoTest) Resample(inp []int16) error { // care moved allocation of output to CalcNeesSamples - logc you can't resample without that
	if rsm.resampled == nil {
		return ErrExpectToCallCalcNeedSamplesPerOutAmtBefore
	}
	return rsm.rsm.Resample(inp, rsm.resampled)
}
func (rsm *resamplerAutoTest) calcNeedSamplesPerOutAmt(outAmt int) int {
	var inAmt int
	inAmt, outAmt = rsm.rsm.CalcInOutSamplesPerOutAmt(outAmt)
	rsm.resampled = make([]int16, outAmt)
	return inAmt
}
func (rsm resamplerAutoTest) OutLen() int {
	return len(rsm.resampled)
}
func (rsm resamplerAutoTest) OutRate() int {
	return rsm.outRate
}
func (rsm resamplerAutoTest) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}
func (rsm resamplerAutoTest) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestResampleAuto_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(60)
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT, goresampler.ResamplerBestFitT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := resamplerAutoTest{}.New(inRate, outRate, rsmT, nil)
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-30)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
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
			rsm := resamplerAutoTest{}.New(inRate, outRate, rsmT, nil)
			var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-30)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
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

	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT, goresampler.ResamplerBestFitT} {
		for _, inRate := range []int{12000, 24000, 22050, 32000, 96000} {
			for _, outRate := range []int{8000, 16000} {
				_, _, err := goresampler.NewResamplerAuto(inRate, outRate, rsmT, nil)
				if !assert.Error(t, err, fmt.Sprintf("expected not to create resampler with config %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
			}
		}
	}
}

func TestResampleAutoDiffErrsNotFall_SinWave(t *testing.T) {
	if testing.Short() { // TODO timely solution cause of large RAM use
		t.Skip("skipping test in short mode.")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(30)
	wg := &sync.WaitGroup{}
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT, goresampler.ResamplerBestFitT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
					rsm := resamplerAutoTest{}.New(inRate, outRate, rsmT, &acc)
					opts := testutils.TestOpts{}.NewDefault().NotFailOnHighDurationErr().NotCalcDuration().WithWaitGroup(wg)
					if rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate)-5 >= int(waveDurS)*inRate {
						continue
					}
					var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate)), rsm, 1, t, opts)
					wg.Add(1)
					go tObj.Run()
				}
			}
		}
	}
	wg.Wait()
}

func ExampleResamplerAuto_Resample() {
	rsm, _, _ := goresampler.NewResamplerAuto(16000, 8000, goresampler.ResamplerBestFitT, nil)
	inAmt, outAmt := rsm.CalcInOutSamplesPerOutAmt(10)
	in := make([]int16, inAmt)
	out := make([]int16, outAmt)
	for i := 0; i < len(in); i++ {
		in[i] = int16(i)
	}
	rsm.Resample(in, out)
	// Output:
}

func ExampleResamplerAuto_Resample_second() {
	rsm, _, _ := goresampler.NewResamplerAuto(96000, 8000, goresampler.ResamplerBestFitNotSafeT, nil)
	inAmt, outAmt := rsm.CalcInOutSamplesPerOutAmt(10)
	in := make([]int16, inAmt)
	out := make([]int16, outAmt)
	for i := 0; i < len(in); i++ {
		in[i] = int16(i)
	}
	rsm.Resample(in, out)
	// Output:
}
