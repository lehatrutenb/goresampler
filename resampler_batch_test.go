package goresampler_test

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"
	"testing"

	"github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"

	"github.com/stretchr/testify/assert"
)

var ErrUnexpTestCfg = errors.New("got unexpected test cfg")

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

type ResampleBatchTest struct {
	inRate    int
	outRate   int
	rsmT      goresampler.ResamplerT
	rsm       goresampler.ResampleBatch
	resampled []int16
	opts      batchWorkType
}

func (ResampleBatchTest) New(inRate, outRate int, rsmT goresampler.ResamplerT, opts batchWorkType) *ResampleBatchTest {
	rsm, _, err := goresampler.NewResamplerAuto(inRate, outRate, rsmT, nil)
	if err != nil {
		panic(err)
	}
	res := new(ResampleBatchTest)
	*res = ResampleBatchTest{inRate, outRate, rsmT, goresampler.NewResampleBatch(rsm), nil, opts}
	return res
}

func (rsm ResampleBatchTest) Copy() testutils.TestResampler {
	res := ResampleBatchTest{}.New(rsm.inRate, rsm.outRate, rsm.rsmT, rsm.opts)
	return res
}
func (rsm ResampleBatchTest) String() string {
	return fmt.Sprintf("%d_to_%d_ResampleBatch_%s", rsm.inRate, rsm.outRate, rsm.rsmT)
}
func (rsm *ResampleBatchTest) Resample(inp []int16) error { //  TODO RM COMM? care moved allocation of output to CalcNeedSamples - log you can't resample without that
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
	if !errors.Is(err, goresampler.ErrNotEnoughSamples) {
		return err
	}

	return nil
}
func (rsm ResampleBatchTest) OutLen() int {
	return len(rsm.resampled)
}
func (rsm ResampleBatchTest) OutRate() int {
	return rsm.outRate
}
func (rsm ResampleBatchTest) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}
func (rsm ResampleBatchTest) UnresampledUngetInAmt() (int, int) {
	return rsm.rsm.UnresampledUngetInAmt(-1)
}

func TestResampleBatch_SinWave(t *testing.T) {
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := ResampleBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 200))
				var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, outRate), 0, inAmt), rsm, 1, t, testutils.TestOpts{}.NewDefault())
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
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := ResampleBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 480))
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
	if testing.Short() { // TODO timely solution cause of large RAM use
		t.Skip("skipping test in short mode.")
	}

	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	wg := &sync.WaitGroup{}
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, outRate := range []int{8000, 16000} {
				for _, addLB := range []bool{false, true} {
					for _, getLB := range []bool{false, true} {
						if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
							continue
						}
						rsm := ResampleBatchTest{}.New(inRate, outRate, rsmT, setParams(addLB, getLB, 1000, 480))
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
	if testing.Short() { // TODO timely solution cause of large RAM use
		t.Skip("skipping test in short mode.")
	}

	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	wg := &sync.WaitGroup{}
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
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
					rsm := ResampleBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, addAmt, 480))
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
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	inAmt := int(5e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waves := testutils.LoadAllRealWaves(0, nil, nil, nil, &inAmt)

	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := ResampleBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 480))

				waveRsmT := rsmT
				if testutils.CheckRsmCompAb(goresampler.ResamplerConstExprT, inRate, outRate) == nil {
					waveRsmT = goresampler.ResamplerConstExprT
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
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	inAmt := int(5e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waves := testutils.LoadAllRealWaves(0, nil, nil, nil, &inAmt)

	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate := range []int{8000, 16000} {
			for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, outRate)
				rsm := ResampleBatchTest{}.New(inRate, outRate, rsmT, setParams(false, false, 1000, 480))

				waveRsmT := rsmT
				if testutils.CheckRsmCompAb(goresampler.ResamplerConstExprT, inRate, outRate) == nil {
					waveRsmT = goresampler.ResamplerConstExprT
				}
				var tObj testutils.TestObj = testutils.TestObj{}.New(waves[testutils.GetWaveName(waveRsmT, inRate, outRate)], rsm, 1, t, testutils.TestOpts{}.NewDefault())
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

func ExampleNewResampleBatch() {
	rsmT := goresampler.ResamplerBestFitT
	rsm, ok, err := goresampler.NewResamplerAuto(16000, 8000, rsmT, nil)
	if !ok {
		fmt.Printf("failed fo fit base %s error sec difference in resampling from %d to %d", rsmT.String(), 16000, 8000)
		return
	}
	if err != nil {
		fmt.Printf("unable to resample with %s from %d to %d", rsmT.String(), 16000, 8000)
		return
	}

	_ = goresampler.NewResampleBatch(rsm)
	fmt.Println("Resampler correctly initialized")
	// Output: Resampler correctly initialized
}

func ExampleResampleBatch_AddBatch() {
	rsmT := goresampler.ResamplerBestFitT
	rsm, _, _ := goresampler.NewResamplerAuto(16000, 8000, rsmT, nil)

	rsmBatch := goresampler.NewResampleBatch(rsm)

	resampledWave := make([]int16, 20)
	for i := int16(0); rsmBatch.GetBatch(resampledWave) == goresampler.ErrNotEnoughSamples; i += 5 {
		rsmBatch.AddBatch([]int16{i, i + 1, i + 2, i + 3, i + 4})
	}

	// fmt.Println(resampledWave) - probably [0 0 2 4 6 8 10 12 14 16 18 20 22 24 26 28 30 32 34 36]

	fmt.Println(len(resampledWave))
	// Output: 20
}

func ExampleResampleBatch_GetBatch() {
	rsmT := goresampler.ResamplerBestFitT
	rsm, _, _ := goresampler.NewResamplerAuto(16000, 8000, rsmT, nil)

	rsmBatch := goresampler.NewResampleBatch(rsm)

	resampledWave := make([]int16, 20)
	for i := int16(0); rsmBatch.GetBatch(resampledWave) == goresampler.ErrNotEnoughSamples; i += 5 {
		rsmBatch.AddBatch([]int16{i, i + 1, i + 2, i + 3, i + 4})
	}

	// fmt.Println(resampledWave) - probably [0 0 2 4 6 8 10 12 14 16 18 20 22 24 26 28 30 32 34 36]

	fmt.Println(len(resampledWave))
	// Output: 20
}

func ExampleResampleBatch_GetLargeBatch() {
	rsmT := goresampler.ResamplerBestFitT
	rsm, _, _ := goresampler.NewResamplerAuto(16000, 8000, rsmT, nil)

	rsmBatch := goresampler.NewResampleBatch(rsm)

	resampledWave := new([]int16)
	*resampledWave = make([]int16, 20)
	for i := int16(0); rsmBatch.GetLargeBatch(resampledWave) == goresampler.ErrNotEnoughSamples; i += 5 {
		rsmBatch.AddBatch([]int16{i, i + 1, i + 2, i + 3, i + 4})
	}

	// fmt.Println(*resampledWave) - probably [0 0 2 4 6 8 10 12 14 16 18 20 22 24 26 28 30 32 34 36]

	fmt.Println(len(*resampledWave))
	// Output: 20
}

func ExampleResampleBatch_UnresampledUngetInAmt() {
	errRate := 1e-6 // fix err rate not to fail after change of it inside resampler
	rsm, _, _ := goresampler.NewResamplerAuto(8000, 16000, goresampler.ResamplerBestFitT, &errRate)

	rsmBatch := goresampler.NewResampleBatch(rsm)
	rsmBatch.AddBatch(make([]int16, 1000))

	resampledWave := make([]int16, 481)
	rsmBatch.GetBatch(resampledWave)

	fmt.Println(rsmBatch.UnresampledUngetInAmt()) // input doesn't matter - used just to have same api as 2Waves batch resampler
	// Output: 759 1
}
