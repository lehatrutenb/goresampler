package goresampler_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type resamplerSpline2Waves struct {
	inRate      int
	outRate1    int
	outRate2    int
	resampled1  []int16
	resampled2  []int16
	maxErrRateP *float64
	returnFirst bool
}

func (resamplerSpline2Waves) New(inRate int, outRate1, outRate2 int, maxErrRateP *float64, returnFirst bool) resamplerSpline2Waves {
	return resamplerSpline2Waves{inRate, outRate1, outRate2, []int16{}, []int16{}, maxErrRateP, returnFirst}
}

func (rsm resamplerSpline2Waves) Copy() testutils.TestResampler {
	res := new(resamplerSpline2Waves)
	*res = rsm.New(rsm.inRate, rsm.outRate1, rsm.outRate2, rsm.maxErrRateP, rsm.returnFirst)
	res.resampled1 = make([]int16, len(rsm.resampled1))
	res.resampled2 = make([]int16, len(rsm.resampled2))
	return res
}

func (rsm resamplerSpline2Waves) String() string {
	if rsm.returnFirst {
		return fmt.Sprintf("%d_to_%d_spline2Waves_resampler", rsm.inRate, rsm.outRate1)
	}
	return fmt.Sprintf("%d_to_%d_spline2Waves_resampler", rsm.inRate, rsm.outRate2)
}

func (rsm *resamplerSpline2Waves) Resample(inp []int16) error {
	sw, _ := goresampler.NewResamplerSpline2Waves(rsm.inRate, rsm.outRate1, rsm.outRate2, nil)
	sw.Resample(inp, rsm.resampled1, rsm.resampled2)
	return nil
}
func (rsm *resamplerSpline2Waves) calcNeedSamplesPerOutAmt(outAmt1, outAmt2 int) int {
	var inAmt int
	sr, _ := goresampler.NewResamplerSpline2Waves(rsm.inRate, rsm.outRate1, rsm.outRate2, nil)
	inAmt, outAmt1, outAmt2 = sr.CalcInOutSamplesPerOutAmt(outAmt1, outAmt2)
	rsm.resampled1 = make([]int16, outAmt1)
	rsm.resampled2 = make([]int16, outAmt2)
	return inAmt
}

func (rsm resamplerSpline2Waves) OutLen() int {
	if rsm.returnFirst {
		return len(rsm.resampled1)
	}
	return len(rsm.resampled2)
}

func (rsm resamplerSpline2Waves) OutRate() int {
	if rsm.returnFirst {
		return rsm.outRate1
	}
	return rsm.outRate2
}

func (rsm resamplerSpline2Waves) Get(ind int) (int16, error) {
	if rsm.returnFirst {
		if ind >= len(rsm.resampled1) {
			return 0, errors.New("out of bounds")
		}
		return rsm.resampled1[ind], nil
	}
	if ind >= len(rsm.resampled2) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled2[ind], nil
}

func (rsm resamplerSpline2Waves) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestResampleSpline2WavesDiffErrsNotFall_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	wg := &sync.WaitGroup{}
	waveDurS := float64(20)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate1 := range []int{8000, 16000} {
			for _, outRate2 := range []int{8000, 16000} {
				for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
					for _, useFirstWave := range []bool{false, true} {
						curOutRate := outRate1
						if !useFirstWave {
							curOutRate = outRate2
						}
						rsm := resamplerSpline2Waves{}.New(inRate, outRate1, outRate2, &acc, useFirstWave)
						opts := testutils.TestOpts{}.NewDefault().NotFailOnHighDurationErr().NotCalcDuration().WithWaitGroup(wg)
						var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, curOutRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate1, (int(waveDurS)-5)*outRate2)), &rsm, 1, t, opts)
						wg.Add(1)
						go tObj.Run()
					}
				}
			}
		}
	}
	wg.Wait()

}

func runTestResampling(inRate, outRate1, outRate2 int, useFirstWave bool, t *testing.T) {
	waveDurS := float64(30)

	curOutRate := outRate1
	if !useFirstWave {
		curOutRate = outRate2
	}
	rsm := resamplerSpline2Waves{}.New(inRate, outRate1, outRate2, nil, useFirstWave)
	opts := testutils.TestOpts{}.NewDefault()
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, curOutRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate1, (int(waveDurS)-5)*outRate2)), &rsm, 1, t, opts)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
}

func TestResampleSpline2Waves8000_SinWave(t *testing.T) {
	inRate := 11025
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	for _, outRate1 := range []int{8000, 16000} {
		for _, outRate2 := range []int{8000, 16000} {
			for _, useFirstWave := range []bool{false, true} {
				runTestResampling(inRate, outRate1, outRate2, useFirstWave, t)
			}
		}
	}
}

func TestResampleSpline2Waves11025_SinWave(t *testing.T) {
	inRate := 11025
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	for _, outRate1 := range []int{8000, 16000} {
		for _, outRate2 := range []int{8000, 16000} {
			for _, useFirstWave := range []bool{false, true} {
				runTestResampling(inRate, outRate1, outRate2, useFirstWave, t)
			}
		}
	}
}

func TestResampleSpline2Waves16000_SinWave(t *testing.T) {
	inRate := 16000
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	for _, outRate1 := range []int{8000, 16000} {
		for _, outRate2 := range []int{8000, 16000} {
			for _, useFirstWave := range []bool{false, true} {
				runTestResampling(inRate, outRate1, outRate2, useFirstWave, t)
			}
		}

	}
}

func TestResampleSpline2Waves44100_SinWave(t *testing.T) {
	inRate := 44100
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	for _, outRate1 := range []int{8000, 16000} {
		for _, outRate2 := range []int{8000, 16000} {
			for _, useFirstWave := range []bool{false, true} {
				runTestResampling(inRate, outRate1, outRate2, useFirstWave, t)
			}
		}
	}
}

func TestResampleSpline2Waves48000_SinWave(t *testing.T) {
	inRate := 48000
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	for _, outRate1 := range []int{8000, 16000} {
		for _, outRate2 := range []int{8000, 16000} {
			for _, useFirstWave := range []bool{false, true} {
				runTestResampling(inRate, outRate1, outRate2, useFirstWave, t)
			}
		}
	}
}
