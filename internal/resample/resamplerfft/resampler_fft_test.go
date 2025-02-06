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
	return res
}

func (rsm resamplerFFT) String() string {
	return fmt.Sprintf("%d_to_%d_fft_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerFFT) Resample(inp []int16) error {
	fr := resamplerfft.FFTResampler{}.New(inp, rsm.inRate, rsm.outRate)
	fr.ResampleFFT()
	rsm.resampled = fr.GetOutWave()
	return nil
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

func TestResampleFFT11025To8(t *testing.T) { // just test that everything counts fine
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, 100, 11025, 8000), 0, 90317*10), testutils.TestResampler(&resamplerFFT{inRate: 11025, outRate: 8000}), 2, t, nil) // 22580*2
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("rsm_fft")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT11025To8RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outR := 8000
	//var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 11025, &outR, &testutils.PATH_TO_BASE_WAVES), 0, 22580*20), testutils.TestResampler(&resamplerFFT{inRate: 11025, outRate: 8000}), 1, t, &testutils.TestOpts{true, "../../../plots"})
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 11025, &outR, &testutils.PATH_TO_BASE_WAVES), 0, 90317*10), testutils.TestResampler(&resamplerFFT{inRate: 11025, outRate: 8000}), 2, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_fft")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
