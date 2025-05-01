package goresampler_test

import (
	"errors"
	"testing"

	"github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"
	"golang.org/x/sync/errgroup"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type resamplerFFT struct {
	inRate      int
	outRate     int
	resampled   []int16
	maxErrRateP *float64
}

func (resamplerFFT) New(inRate int, outRate int, maxErrRateP *float64) resamplerFFT {
	return resamplerFFT{inRate, outRate, []int16{}, maxErrRateP}
}

func (rsm resamplerFFT) Copy() testutils.TestResampler {
	res := new(resamplerFFT)
	*res = rsm.New(rsm.inRate, rsm.outRate, rsm.maxErrRateP)
	res.resampled = make([]int16, len(rsm.resampled))
	return res
}

func (rsm resamplerFFT) String() string {
	return fmt.Sprintf("%d_to_%d_fft_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerFFT) Resample(inp []int16) error {
	fr, _, err := goresampler.NewResamplerFFT(rsm.inRate, rsm.outRate, nil)
	if err != nil {
		return err
	}
	fr.Resample(inp, rsm.resampled)
	return nil
}
func (rsm *resamplerFFT) calcNeedSamplesPerOutAmt(outAmt int) int {
	fr, _, err := goresampler.NewResamplerFFT(rsm.inRate, rsm.outRate, nil)
	if err != nil {
		panic(err)
	}
	inAmt, resOutAmt := fr.CalcInOutSamplesPerOutAmt(int64(outAmt))
	rsm.resampled = make([]int16, resOutAmt)
	return int(inAmt)
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

func (rsm resamplerFFT) UnresampledUngetInAmt() (int, int) {
	return 0, 0
}

func TestResampleFFTDiffErrsNotFall_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	eg := &errgroup.Group{}
	waveDurS := float64(30)
	for _, inRate := range []int{8000, 11025, 16000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			if inRate < outRate {
				continue
			}
			for _, acc := range []float64{1, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7, 1e-8, 1e-9, 0} {
				rsm := resamplerFFT{}.New(inRate, outRate, &acc)
				if rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate)-5 >= int(waveDurS)*inRate {
					continue
				}
				opts := testutils.TestOpts{}.NewDefault().NotCalcDuration().NotFailOnHighDurationErr()
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-5)*outRate)), &rsm, 1, t, opts)
				go tObj.Run()
			}
		}
	}
	assert.NoError(t, eg.Wait())
}

func testOnSinWaveFFT(t *testing.T, inRate int, outRate int, waveDurS float64) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate, nil)
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-30)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_fft")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT11025To8_SinWave(t *testing.T) {
	inRate := 11025
	outRate := 8000
	waveDurS := float64(60)
	testOnSinWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT16To8_SinWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	waveDurS := float64(60)
	testOnSinWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT44100To8_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	waveDurS := float64(60)
	testOnSinWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT48To8_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	waveDurS := float64(60)
	testOnSinWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT44100To16_SinWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	waveDurS := float64(60)
	testOnSinWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT48To16_SinWave(t *testing.T) {
	inRate := 48000
	outRate := 16000
	waveDurS := float64(60)
	testOnSinWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFTExpErr(t *testing.T) {
	for _, inRate := range []int{8000, 16000} {
		for _, outRate := range []int{8000, 11025, 16000, 41100, 48000} {
			if inRate >= outRate {
				continue
			}
			_, _, err := goresampler.NewResamplerFFT(inRate, outRate, nil)
			assert.ErrorIs(t, err, goresampler.ErrGotInRateLessThanOutRate)
		}
	}
}

func testOnRealWaveFFT(t *testing.T, inRate int, outRate int, waveDurS float64) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	rsm := resamplerFFT{}.New(inRate, outRate, nil)

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, nil), 0, rsm.calcNeedSamplesPerOutAmt((int(waveDurS)-30)*outRate)), &rsm, 1, t, testutils.TestOpts{}.NewDefault().NotFailOnHighErr())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("rsm_fft")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResampleFFT11025To8000_RealWave(t *testing.T) {
	inRate := 11025
	outRate := 8000
	waveDurS := float64(60)
	testOnRealWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT16000To8000_RealWave(t *testing.T) {
	inRate := 16000
	outRate := 8000
	waveDurS := float64(60)
	testOnRealWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT44100To8000_RealWave(t *testing.T) {
	inRate := 44100
	outRate := 8000
	waveDurS := float64(60)
	testOnRealWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT48000To8000_RealWave(t *testing.T) {
	inRate := 48000
	outRate := 8000
	waveDurS := float64(60)
	testOnRealWaveFFT(t, inRate, outRate, waveDurS)
}

func TestResampleFFT44100To16000_RealWave(t *testing.T) {
	inRate := 44100
	outRate := 16000
	waveDurS := float64(60)
	testOnRealWaveFFT(t, inRate, outRate, waveDurS)
}
