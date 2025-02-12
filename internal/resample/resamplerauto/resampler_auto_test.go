package resamplerauto_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"resampler/internal/resample/resamplerauto"
	"resampler/internal/resample/resampleri"
	testutils "resampler/internal/test_utils"

	"github.com/stretchr/testify/assert"
)

// another type of tests as in all resamplers - just all in 1 to check that evrything out of base work is fine

var ErrExpectToCallCalcNeedSamplesPerOutAmtBefore = errors.New("error expected to call resamplerAutoTest.CalcNeedSamplesPerOutAmtBefore")

type resamplerAutoTest struct {
	inRate    int
	outRate   int
	rsmT      resamplerauto.ResamplerT
	rsm       resampleri.Resampler
	resampled []int16
}

func (resamplerAutoTest) New(inRate, outRate int, rsmT resamplerauto.ResamplerT) *resamplerAutoTest {
	rsm, err := resamplerauto.New(inRate, outRate, rsmT)
	if err != nil {
		panic(err)
	}
	res := new(resamplerAutoTest)
	*res = resamplerAutoTest{inRate, outRate, rsmT, rsm, nil}
	return res
}

func (rsm resamplerAutoTest) Copy() testutils.TestResampler {
	res := resamplerAutoTest{}.New(rsm.inRate, rsm.outRate, rsm.rsmT)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}
func (rsm resamplerAutoTest) String() string {
	return fmt.Sprintf("%d_to_%d_resamplerAuto", rsm.inRate, rsm.outRate)
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
func (rsm resamplerAutoTest) UnresampledInAmt() int {
	return 0
}

func TestResampleAuto_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	for _, rsmT := range []resamplerauto.ResamplerT{resamplerauto.ResamplerConstExpr, resamplerauto.ResamplerSpline, resamplerauto.ResamplerFFT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := resamplerAutoTest{}.New(inRate, outRate, rsmT)
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-10)*outRate)), rsm, 1, t, testutils.TestOpts{}.NewDefault())
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
			}
		}
	}
}
