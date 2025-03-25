package goresampler_test

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"
	"testing"

	goresampler "github.com/lehatrutenb/goresampler"

	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"

	"github.com/stretchr/testify/assert"
)

type ResampleBatch2WavesTest struct {
	inRate      int
	outRate1    int
	outRate2    int
	rsmT        goresampler.Resampler2WavesT
	rsm         goresampler.ResampleBatch2Waves
	resampled   []int16
	opts        batchWorkType
	returnFirst bool
}

func (ResampleBatch2WavesTest) New(inRate, outRate1, outRate2 int, rsmT goresampler.Resampler2WavesT, opts batchWorkType, returnFirst, resampleTail bool) *ResampleBatch2WavesTest {
	rsm, _, err := goresampler.NewResamplerAuto2Waves(inRate, outRate1, outRate2, rsmT, nil)
	if err != nil {
		panic(err)
	}
	res := new(ResampleBatch2WavesTest)
	*res = ResampleBatch2WavesTest{inRate, outRate1, outRate2, rsmT, goresampler.NewResampleBatch2Waves(rsm, inRate, outRate1, outRate2), nil, opts, returnFirst}
	return res
}

func (rsm ResampleBatch2WavesTest) Copy() testutils.TestResampler {
	res := ResampleBatch2WavesTest{}.New(rsm.inRate, rsm.outRate1, rsm.outRate2, rsm.rsmT, rsm.opts, rsm.returnFirst, rsm.returnFirst)
	return res
}
func (rsm ResampleBatch2WavesTest) String() string {
	if rsm.returnFirst {
		return fmt.Sprintf("%d_to_%d_ResampleBatch2Waves_%s", rsm.inRate, rsm.outRate1, rsm.rsmT)
	}
	return fmt.Sprintf("%d_to_%d_ResampleBatch2Waves_%s", rsm.inRate, rsm.outRate2, rsm.rsmT)
}
func (rsm *ResampleBatch2WavesTest) Resample(inp []int16) error { //  TODO RM COMM? care moved allocation of output to CalcNeedSamples - log you can't resample without that
	rsm.resampled = make([]int16, rsm.opts.getBSz)
	var err error

	if len(inp)%rsm.opts.addBSz != 0 && !rsm.opts.resampleTail {
		return ErrUnexpTestCfg
	}
	for len(inp) >= rsm.opts.addBSz {
		rsm.rsm.AddBatch(inp[:rsm.opts.addBSz])
		inp = inp[rsm.opts.addBSz:]
	}

	ind := 0
	for {
		if rsm.opts.getLargeBatches {
			rsm.resampled = rsm.resampled[:ind+rsm.opts.getBSz]
			out := rsm.resampled[ind : ind+rsm.opts.getBSz]
			if rsm.returnFirst {
				err = rsm.rsm.GetLargeBatchFirstWave(&out)
			} else {
				err = rsm.rsm.GetLargeBatchSecondWave(&out)
			}
			copy(rsm.resampled[ind:ind+rsm.opts.getBSz], out)
		} else {
			rsm.resampled = rsm.resampled[:ind+rsm.opts.getBSz]
			if rsm.returnFirst {
				err = rsm.rsm.GetBatchFirstWave(rsm.resampled[ind : ind+rsm.opts.getBSz])
			} else {
				err = rsm.rsm.GetBatchSecondWave(rsm.resampled[ind : ind+rsm.opts.getBSz])
			}
		}
		if err != nil {
			rsm.resampled = rsm.resampled[:ind]
			break
		}
		ind += rsm.opts.getBSz
		rsm.resampled = slices.Grow(rsm.resampled, rsm.opts.getBSz)
	}
	if errors.Is(err, goresampler.ErrNotEnoughSamples) && rsm.opts.resampleTail {
		rsm.rsm.ResampleAllInBuf()
		_, outSz1, outSz2 := rsm.rsm.UnresampledUngetInAmt()
		if rsm.returnFirst {
			rsm.resampled = slices.Grow(rsm.resampled, outSz1)
			rsm.resampled = rsm.resampled[:ind+outSz1]
			err = rsm.rsm.GetBatchFirstWave(rsm.resampled[ind : ind+outSz1])
		} else {
			rsm.resampled = slices.Grow(rsm.resampled, outSz2)
			rsm.resampled = rsm.resampled[:ind+outSz2]
			err = rsm.rsm.GetBatchFirstWave(rsm.resampled[ind : ind+outSz2])
		}
	}
	if !errors.Is(err, goresampler.ErrNotEnoughSamples) {
		return err
	}

	return nil
}
func (rsm ResampleBatch2WavesTest) OutLen() int {
	return len(rsm.resampled)
}
func (rsm ResampleBatch2WavesTest) OutRate() int {
	if rsm.returnFirst {
		return rsm.outRate1
	}
	return rsm.outRate2
}
func (rsm ResampleBatch2WavesTest) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}
func (rsm ResampleBatch2WavesTest) UnresampledUngetInAmt() (int, int) {
	f, s, th := rsm.rsm.UnresampledUngetInAmt()
	if rsm.returnFirst {
		return f, s
	}
	return f, th
}

func TestResampleBatch2Waves_SinWave(t *testing.T) {
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	rsmT := goresampler.Resampler2WavesSplineT
	waveDurS := float64(20)
	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate1 := range []int{8000, 16000} {
			for _, outRate2 := range []int{8000, 16000} {
				for _, useFirstWave := range []bool{false, true} {
					if testutils.CheckRsmCompAb(rsmT, inRate, outRate1) != nil && testutils.CheckRsmCompAb(rsmT, inRate, outRate2) != nil {
						continue
					}
					curOutRate := outRate1
					if !useFirstWave {
						curOutRate = outRate2
					}
					log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, curOutRate)
					rsm := ResampleBatch2WavesTest{}.New(inRate, outRate1, outRate2, rsmT, setParams(false, false, 1000, 200, false), useFirstWave, false)
					var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, curOutRate), 0, inAmt), rsm, 1, t, testutils.TestOpts{}.NewDefault())
					err := tObj.Run()
					if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, curOutRate)) {
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
}

func TestResampleBatch2Waves_Reset(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(60)
	for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			outRate1 := 8000
			outRate2 := 16000

			if testutils.CheckRsmCompAb(rsmT, inRate, outRate1) != nil {
				continue
			}
			if testutils.CheckRsmCompAb(rsmT, inRate, outRate2) != nil {
				continue
			}

			waveLenGet1 := (int(waveDurS) - 40) * outRate1
			waveLenGet2 := (int(waveDurS) - 40) * outRate2

			var err error
			waves := make([][]int16, 2)
			res := make([][]int16, 4)
			waves[0], err = testutils.GetFullInWave(testutils.SinWave{}.New(0, waveDurS, inRate, outRate1))
			waves[1], _ = testutils.GetFullInWave(testutils.SinWave{}.New(0, waveDurS, inRate, outRate1))
			inRsm, _, _ := goresampler.NewResamplerAuto2Waves(8000, 16000, 8000, goresampler.Resampler2WavesSplineT, nil)
			rsm := goresampler.NewResampleBatch2Waves(inRsm, inRate, outRate1, outRate2)
			for attemptInd := 0; attemptInd < 2; attemptInd++ {
				rsm.Reset()
				if !assert.NoError(t, err) {
					t.Error(err)
				}
				err = rsm.AddBatch(waves[attemptInd])
				if !assert.NoError(t, err) {
					t.Error(err)
				}

				res[attemptInd*2] = make([]int16, waveLenGet1)
				res[attemptInd*2+1] = make([]int16, waveLenGet2)
				err = rsm.GetBatchFirstWave(res[attemptInd*2])
				if !assert.NoError(t, err) {
					t.Error(err)
				}
				err = rsm.GetBatchSecondWave(res[attemptInd*2+1])
				if !assert.NoError(t, err) {
					t.Error(err)
				}
				rsm.Reset()
			}

			assert.Equal(t, res[0], res[2], "After reset previous resampling mustn't affect on output of second saame wave")
			assert.Equal(t, res[1], res[3], "After reset previous resampling mustn't affect on output of second saame wave")
		}
	}
}

func TestResampleBatch2Waves_SinWave2Ch(t *testing.T) {
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	rsmT := goresampler.Resampler2WavesSplineT
	waveDurS := float64(20)
	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate1 := range []int{8000, 16000} {
			for _, outRate2 := range []int{8000, 16000} {
				for _, useFirstWave := range []bool{false, true} {
					if testutils.CheckRsmCompAb(rsmT, inRate, outRate1) != nil && testutils.CheckRsmCompAb(rsmT, inRate, outRate2) != nil {
						continue
					}
					curOutRate := outRate1
					if !useFirstWave {
						curOutRate = outRate2
					}
					log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, curOutRate)
					rsm := ResampleBatch2WavesTest{}.New(inRate, outRate1, outRate2, rsmT, setParams(false, false, 1000, 480, false), useFirstWave, false)
					sw := testutils.SinWave{}.New(0, waveDurS, inRate, curOutRate)
					sw = (sw.(testutils.SinWave)).WithChannelAmt(2)
					var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(sw, 0, inAmt), rsm, 1, t, testutils.TestOpts{}.NewDefault())
					err := tObj.Run()
					if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, curOutRate)) {
						t.Error(err)
					}
				}
			}
		}
	}
}

func TestResampleBatch2WavesDiffAddGetTypes_SinWave(t *testing.T) {
	if testing.Short() { // TODO timely solution cause of large RAM use
		t.Skip("skipping test in short mode.")
	}

	rsmT := goresampler.Resampler2WavesSplineT
	inAmt := int(1e5)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	wg := &sync.WaitGroup{}
	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate1 := range []int{8000, 16000} {
			for _, outRate2 := range []int{8000, 16000} {
				for _, useFirstWave := range []bool{false, true} {
					for _, resampleTail := range []bool{false, true} {
						if testutils.CheckRsmCompAb(rsmT, inRate, outRate1) != nil && testutils.CheckRsmCompAb(rsmT, inRate, outRate2) != nil {
							continue
						}
						curOutRate := outRate1
						if !useFirstWave {
							curOutRate = outRate2
						}

						opts := testutils.TestOpts{}.NewDefault().WithWaitGroup(wg)
						if resampleTail {
							opts.NotFailOnHighDurationErr()
						}
						for _, addLB := range []bool{false, true} {
							for _, getLB := range []bool{false, true} {
								rsm := ResampleBatch2WavesTest{}.New(inRate, outRate1, outRate2, rsmT, setParams(addLB, getLB, 1000, 480, resampleTail), useFirstWave, resampleTail)
								var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, curOutRate), 0, inAmt), rsm, 1, t, opts)
								wg.Add(1)
								go tObj.Run()
							}
						}
					}
				}
			}
		}
	}

	wg.Wait()
}

func TestResampleBatch2WavesDiffAddAmt_SinWave(t *testing.T) {
	if testing.Short() { // TODO timely solution cause of large RAM use
		t.Skip("skipping test in short mode.")
	}

	inAmt := int(1e5)
	rsmT := goresampler.Resampler2WavesSplineT
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waveDurS := float64(20)
	wg := &sync.WaitGroup{}
	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate1 := range []int{8000, 16000} {
			for _, outRate2 := range []int{8000, 16000} {
				for _, useFirstWave := range []bool{false, true} {
					if testutils.CheckRsmCompAb(rsmT, inRate, outRate1) != nil && testutils.CheckRsmCompAb(rsmT, inRate, outRate2) != nil {
						continue
					}
					curOutRate := outRate1
					if !useFirstWave {
						curOutRate = outRate2
					}
					log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, curOutRate)
					for addAmt := 1; addAmt <= 100; addAmt++ {
						curInAmt := inAmt
						if curInAmt%addAmt != 0 {
							curInAmt += addAmt - (curInAmt % addAmt)
						}
						inWave := testutils.CutWave{}.New(testutils.SinWave{}.New(0, waveDurS, inRate, curOutRate), 0, curInAmt)
						rsm := ResampleBatch2WavesTest{}.New(inRate, outRate1, outRate2, rsmT, setParams(false, false, addAmt, 480, false), useFirstWave, false)
						var tObj testutils.TestObj = testutils.TestObj{}.New(inWave, rsm, 1, t, testutils.TestOpts{}.NewDefault().WithWaitGroup(wg))
						wg.Add(1)
						go tObj.Run()
					}
				}
			}
		}
	}

	wg.Wait()
}

func TestResampleBatch2Waves_RealWave(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	inAmt := int(5e5)
	rsmT := goresampler.Resampler2WavesSplineT
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	waves := testutils.LoadAllRealWaves(0, nil, nil, nil, &inAmt)

	for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
		for _, outRate1 := range []int{8000, 16000} {
			for _, outRate2 := range []int{8000, 16000} {
				for _, useFirstWave := range []bool{false, true} {
					curOutRate := outRate1
					if !useFirstWave {
						curOutRate = outRate2
					}
					if testutils.CheckRsmCompAb(goresampler.Resampler2WavesSplineT, inRate, outRate1) != nil && testutils.CheckRsmCompAb(goresampler.Resampler2WavesSplineT, inRate, outRate2) != nil {
						continue
					}
					log.Printf("Testing %s from %d to %d\n", rsmT.String(), inRate, curOutRate)
					rsm := ResampleBatch2WavesTest{}.New(inRate, outRate1, outRate2, rsmT, setParams(false, false, 1000, 480, false), useFirstWave, false)

					waveRsmT := goresampler.ResamplerSplineT
					if testutils.CheckRsmCompAb(goresampler.ResamplerConstExprT, inRate, curOutRate) == nil {
						waveRsmT = goresampler.ResamplerConstExprT
					}
					var tObj testutils.TestObj = testutils.TestObj{}.New(waves[testutils.GetWaveName(waveRsmT, inRate, curOutRate)], rsm, 1, t, testutils.TestOpts{}.NewDefault().NotFailOnHighErr())
					err := tObj.Run()
					if !assert.NoError(t, err, fmt.Sprintf("failed to convert via %s from %d to %d", rsmT, inRate, curOutRate)) {
						t.Error(err)
					}
				}
			}
		}
	}
}

func ExampleNewResampleBatch2Waves() {
	rsmT := goresampler.Resampler2WavesSplineT
	rsm, ok, err := goresampler.NewResamplerAuto2Waves(16000, 8000, 16000, rsmT, nil)
	if !ok {
		fmt.Printf("failed fo fit base %s error sec difference in resampling from %d to %d", rsmT.String(), 16000, 8000)
		return
	}
	if err != nil {
		fmt.Printf("unable to resample with %s from %d to %d", rsmT.String(), 16000, 8000)
		return
	}

	_ = goresampler.NewResampleBatch2Waves(rsm, 16000, 8000, 16000)
	fmt.Println("Resampler correctly initialized")
	// Output: Resampler correctly initialized
}

func ExampleResampleBatch2Waves_AddBatch() {
	var err error
	defer func() { _ = err }()

	rsmT := goresampler.Resampler2WavesSplineT
	rsm, _, err := goresampler.NewResamplerAuto2Waves(16000, 8000, 16000, rsmT, nil)

	rsmBatch := goresampler.NewResampleBatch2Waves(rsm, 16000, 8000, 16000)

	for i := int16(0); i < 100; i++ {
		err = rsmBatch.AddBatch([]int16{i, i + 1})
	}

	fmt.Println(rsmBatch.UnresampledUngetInAmt())
	// Output:
	// 200 0 0
}

func ExampleResampleBatch2Waves_GetBatchFirstWave() {
	var err error
	defer func() { _ = err }()

	rsmT := goresampler.Resampler2WavesSplineT
	rsm, _, err := goresampler.NewResamplerAuto2Waves(16000, 8000, 16000, rsmT, nil)

	rsmBatch := goresampler.NewResampleBatch2Waves(rsm, 16000, 8000, 16000)

	resampledWave1 := make([]int16, 20)
	resampledWave2 := make([]int16, 40)
	for i := int16(0); rsmBatch.GetBatchFirstWave(resampledWave1) == goresampler.ErrNotEnoughSamples; i += 5 {
		err = rsmBatch.AddBatch([]int16{i, i + 1, i + 2, i + 3, i + 4})
	}
	for {
		if rsmBatch.GetBatchSecondWave(resampledWave2) != nil {
			break
		}
	}

	// fmt.Println(resampledWave) - probably [0 0 2 4 6 8 10 12 14 16 18 20 22 24 26 28 30 32 34 36]

	fmt.Println(len(resampledWave1), len(resampledWave2))
	// Output: 20 40
}

func ExampleResampleBatch2Waves_GetLargeBatchFirstWave() {
	var err error
	defer func() { _ = err }()

	rsmT := goresampler.Resampler2WavesSplineT
	rsm, _, err := goresampler.NewResamplerAuto2Waves(16000, 8000, 16000, rsmT, nil)

	rsmBatch := goresampler.NewResampleBatch2Waves(rsm, 16000, 8000, 16000)

	resampledWave1 := new([]int16)
	*resampledWave1 = make([]int16, 20)
	resampledWave2 := new([]int16)
	*resampledWave2 = make([]int16, 40)
	for i := int16(0); rsmBatch.GetLargeBatchFirstWave(resampledWave1) == goresampler.ErrNotEnoughSamples; i += 5 {
		err = rsmBatch.AddBatch([]int16{i, i + 1, i + 2, i + 3, i + 4})
	}
	for {
		if rsmBatch.GetLargeBatchSecondWave(resampledWave2) != nil {
			break
		}
	}

	// fmt.Println(*resampledWave1) - probably [0 0 2 4 6 8 10 12 14 16 18 20 22 24 26 28 30 32 34 36]

	fmt.Println(len(*resampledWave1), len(*resampledWave2))
	// Output: 20 40
}

func ExampleResampleBatch2Waves_UnresampledUngetInAmt() {
	var err error
	defer func() { _ = err }()

	errRate := 1e-6 // fix err rate not to fail after change of it inside resampler
	rsm, _, err := goresampler.NewResamplerAuto2Waves(8000, 16000, 8000, goresampler.Resampler2WavesSplineT, &errRate)

	rsmBatch := goresampler.NewResampleBatch2Waves(rsm, 8000, 16000, 8000)
	err = rsmBatch.AddBatch(make([]int16, 1000))

	resampledWave := make([]int16, 481)
	err = rsmBatch.GetBatchFirstWave(resampledWave)

	fmt.Println(rsmBatch.UnresampledUngetInAmt())
	// Output:
	// 730 59 270
}

func ExampleResampleBatch2Waves_ResampleAllInBuf() {
	var err error
	defer func() { _ = err }()

	errRate := 1e-6 // fix err rate not to fail after change of it inside resampler
	rsm, _, err := goresampler.NewResamplerAuto2Waves(8000, 16000, 8000, goresampler.Resampler2WavesSplineT, &errRate)

	rsmBatch := goresampler.NewResampleBatch2Waves(rsm, 8000, 16000, 8000)
	err = rsmBatch.AddBatch(make([]int16, 1000))

	resampledWave := make([]int16, 481)
	err = rsmBatch.GetBatchFirstWave(resampledWave)

	fmt.Println(rsmBatch.UnresampledUngetInAmt())
	err = rsmBatch.ResampleAllInBuf() // resample tails - care that may cause not same sound duration - may loose ~ 1/outRate seconds

	fmt.Println(rsmBatch.UnresampledUngetInAmt())
	// Output:
	// 730 59 270
	// 0 1519 1000
}
