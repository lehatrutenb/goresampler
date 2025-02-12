package resample_test

import (
	"errors"
	"fmt"
	"log"
	"resampler/internal/resample"
	"resampler/internal/resample/resamplerauto"
	testutils "resampler/internal/test_utils"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ErrUnexpTestCfg = errors.New("got unexpected test cfg")
var PATH_TO_BASE_WAVES = "../../base_waves/"

type batchWorkType struct {
	addLargeBatches bool
	getLargeBatches bool
	addBSz          int
	getBSz          int
	//addAtOnce       bool
	//getAtOnce       bool
}

func setParams(addLBs bool, getLBs bool, addBSz, getBSz int) batchWorkType {
	return batchWorkType{addLBs, getLBs, addBSz, getBSz}
}

type resamplerBatchTest struct {
	inRate    int
	outRate   int
	rsmT      resamplerauto.ResamplerT
	rsm       resample.ResamplerBatch
	resampled []int16
	opts      batchWorkType
}

func (resamplerBatchTest) New(inRate, outRate int, rsmT resamplerauto.ResamplerT, opts batchWorkType) *resamplerBatchTest {
	rsm, err := resamplerauto.New(inRate, outRate, rsmT)
	if err != nil {
		panic(err)
	}
	res := new(resamplerBatchTest)
	*res = resamplerBatchTest{inRate, outRate, rsmT, resample.New(rsm), nil, opts}
	return res
}

func (rsm resamplerBatchTest) Copy() testutils.TestResampler {
	res := resamplerBatchTest{}.New(rsm.inRate, rsm.outRate, rsm.rsmT, rsm.opts)
	return res
}
func (rsm resamplerBatchTest) String() string {
	return fmt.Sprintf("%d_to_%d_resamplerBatch", rsm.inRate, rsm.outRate)
}
func (rsm *resamplerBatchTest) Resample(inp []int16) error { // care moved allocation of output to CalcNeesSamples - logc you can't resample without that
	rsm.resampled = make([]int16, rsm.opts.getBSz)
	var err error

	if len(inp)%rsm.opts.addBSz != 0 {
		return ErrUnexpTestCfg
	}
	for len(inp) != 0 {
		rsm.rsm.AddBatch(inp[:rsm.opts.addBSz])
		inp = inp[rsm.opts.addBSz:]
	}

	ind := 0
	for {
		if rsm.opts.getLargeBatches {
			out := rsm.resampled[ind:]
			err = rsm.rsm.GetLargeBatch(&out)
		} else {
			rsm.resampled = rsm.resampled[:ind+rsm.opts.getBSz]
			err = rsm.rsm.GetBatch(rsm.resampled[ind : ind+rsm.opts.getBSz])
		}
		if err != nil {
			rsm.resampled = rsm.resampled[:ind]
			break
		}
		ind += rsm.opts.getBSz
		rsm.resampled = slices.Grow(rsm.resampled, rsm.opts.getBSz)
	}
	if !errors.Is(err, resample.ErrNotEnoughSamples) {
		return err
	}

	return nil
}
func (rsm resamplerBatchTest) OutLen() int {
	return len(rsm.resampled)
}
func (rsm resamplerBatchTest) OutRate() int {
	return rsm.outRate
}
func (rsm resamplerBatchTest) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}
func (rsm resamplerBatchTest) UnresampledInAmt() int {
	return rsm.rsm.UnresampledInAmt()
}

func TestResampleAuto_SinWave(t *testing.T) {
	inAmt := int(1e5)
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
				rsm := resamplerBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 480))
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, inAmt), rsm, 1, t, nil)
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
			}
		}
	}
}

func TestResampleAuto_RealWave(t *testing.T) {
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waves := testutils.LoadAllRealWaves(0, &PATH_TO_BASE_WAVES, nil, nil, &inAmt)

	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, rsmT := range []resamplerauto.ResamplerT{resamplerauto.ResamplerConstExpr, resamplerauto.ResamplerSpline, resamplerauto.ResamplerFFT} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := resamplerBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 480))

				waveRsmT := rsmT
				if testutils.CheckRsmCompAb(resamplerauto.ResamplerConstExpr, inRate, outRate) == nil {
					waveRsmT = resamplerauto.ResamplerConstExpr
				}
				var tObj testutils.TestObj = testutils.TestObj{}.New(waves[testutils.GetWaveName(waveRsmT, inRate, outRate)], rsm, 1, t, nil)
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
			}
		}
	}
}
