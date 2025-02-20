package resample_test

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"
	"testing"

	"github.com/lehatrutenb/go_resampler/internal/resample"
	"github.com/lehatrutenb/go_resampler/internal/resample/resamplerauto"
	testutils "github.com/lehatrutenb/go_resampler/internal/test_utils"

	"github.com/stretchr/testify/assert"
)

var ErrUnexpTestCfg = errors.New("got unexpected test cfg")
var PATH_TO_BASE_WAVES = "../../base_waves/"
var testPath = "../../test/"

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
	rsm, _, err := resamplerauto.New(inRate, outRate, rsmT, nil)
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
	return fmt.Sprintf("%d_to_%d_resamplerBatch_%s", rsm.inRate, rsm.outRate, rsm.rsmT)
}
func (rsm *resamplerBatchTest) Resample(inp []int16) error { //  TODO RM COMM? care moved allocation of output to CalcNeedSamples - log you can't resample without that
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
			rsm.resampled = rsm.resampled[:ind+rsm.opts.getBSz]
			out := rsm.resampled[ind : ind+rsm.opts.getBSz]
			err = rsm.rsm.GetLargeBatch(&out)
			copy(rsm.resampled[ind:ind+rsm.opts.getBSz], out)
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
func (rsm resamplerBatchTest) UnresampledUngetInAmt() (int, int) {
	return rsm.rsm.UnresampledUngetInAmt()
}

func TestResampleBatch_SinWave(t *testing.T) {
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
				rsm := resamplerBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 200))
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, inAmt), rsm, 1, t, testutils.TestOpts{}.New(false, &testPath))
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
				err = tObj.Save("rsm_batch")
				if !assert.NoError(t, err, "failed to save test results") {
					t.Error(err)
				}
			}
		}
	}
}

func TestResampleBatch_SinWave2Ch(t *testing.T) {
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
				sw := testutils.SinWave{}.New(0, waveDurS, inRate, outRate)
				sw = (sw.(testutils.SinWave)).WithChannelAmt(2)
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(sw, 0, inAmt), rsm, 1, t, testutils.TestOpts{}.NewDefault())
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
			}
		}
	}
}

func TestResampleBatchDiffAddGetTypes_SinWave(t *testing.T) {
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	wg := &sync.WaitGroup{}
	for _, rsmT := range []resamplerauto.ResamplerT{resamplerauto.ResamplerConstExpr, resamplerauto.ResamplerSpline, resamplerauto.ResamplerFFT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				for _, addLB := range []bool{false, true} {
					for _, getLB := range []bool{false, true} {
						if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
							continue
						}
						rsm := resamplerBatchTest{}.New(inRate, outRate, rsmT, setParams(addLB, getLB, 1000, 480))
						var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, inAmt), rsm, 1, t, testutils.TestOpts{}.NewDefault().WithWaitGroup(wg))
						wg.Add(1)
						go tObj.Run()
					}
				}
			}
		}
	}

	wg.Wait()
}

func TestResampleBatchDiffAddAmt_SinWave(t *testing.T) {
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	wg := &sync.WaitGroup{}
	for _, rsmT := range []resamplerauto.ResamplerT{resamplerauto.ResamplerConstExpr, resamplerauto.ResamplerSpline, resamplerauto.ResamplerFFT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				for addAmt := 1; addAmt <= 500; addAmt++ {
					curInAmt := inAmt
					if curInAmt%addAmt != 0 {
						curInAmt += addAmt - (curInAmt % addAmt)
					}
					inWave := testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, curInAmt)
					rsm := resamplerBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, addAmt, 480))
					var tObj testutils.TestObj = testutils.TestObj{}.New(inWave, rsm, 1, t, testutils.TestOpts{}.NewDefault().WithWaitGroup(wg))
					wg.Add(1)
					go tObj.Run()
				}
			}
		}
	}

	wg.Wait()
}

func TestResampleBatch_RealWave(t *testing.T) {
	inAmt := int(5e5)
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
				var tObj testutils.TestObj = testutils.TestObj{}.New(waves[testutils.GetWaveName(waveRsmT, inRate, outRate)], rsm, 1, t, testutils.TestOpts{}.NewDefault())
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
			}
		}
	}
}

func TestResampleBatchSaveReports_RealWave(t *testing.T) {
	inAmt := int(5e5)
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
				var tObj testutils.TestObj = testutils.TestObj{}.New(waves[testutils.GetWaveName(waveRsmT, inRate, outRate)], rsm, 1, t, testutils.TestOpts{}.New(true, &testPath))
				err := tObj.Run()
				if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, outRate)) {
					t.Error(err)
				}
				err = tObj.Save("rsm_batch")
				if !assert.NoError(t, err, "failed to save test results") {
					t.Error(err)
				}
			}
		}
	}
}
