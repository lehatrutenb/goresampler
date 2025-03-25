//go:build !NoBenchmarks

package benchmark

import (
	"errors"
	"fmt"
	"testing"

	goresampler "github.com/lehatrutenb/goresampler"
	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"
	"github.com/lehatrutenb/goresampler/internal/utils"
)

type BenchmarkerBatch struct {
	b        *testing.B
	rsmT     goresampler.Resampler2WavesT
	inRate   int
	outRate1 int
	outRate2 int
	rsmIns   goresampler.Resampler2Waves
	rsm      goresampler.ResampleBatch2Waves
	wave1    testutils.CutWave
	in       []int16
}

func (bmr BenchmarkerBatch) New(rsmT goresampler.Resampler2WavesT, inRate, outRate1, outRate2 int, b *testing.B) BenchmarkerBatch {
	if min(outRate1, outRate2)*2 != max(outRate1, outRate2) {
		b.Error(errors.New("expected to get 8000 and 16000 as outRates"))
		b.FailNow()
	}
	if err := testutils.CheckRsmCompAb(rsmT, inRate, outRate1); err != nil {
		b.Error(err)
		b.FailNow()
	}
	if err := testutils.CheckRsmCompAb(rsmT, inRate, outRate2); err != nil {
		b.Error(err)
		b.FailNow()
	}

	rsm, _, err := goresampler.NewResamplerAuto2Waves(inRate, outRate1, outRate2, rsmT, nil)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	rsmBatch := goresampler.NewResampleBatch2Waves(rsm, inRate, outRate1, outRate2)

	rsmInsT, err := rsmT.GetRsmIns()
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	wave := RealWaves[fmt.Sprintf("%d:%d:%d", rsmInsT, inRate, outRate1)]
	in, err := testutils.GetFullInWave(wave)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	return BenchmarkerBatch{b, rsmT, inRate, outRate1, outRate2, rsm, rsmBatch, wave, in}
}

func (bmr BenchmarkerBatch) setup() ([]int16, []int16, [2][]int16, [2][]int16) {
	in1 := utils.GetWithStep(bmr.in, 0, 2)
	in2 := utils.GetWithStep(bmr.in, 1, 2)
	outLen1, outLen2 := testutils.CalcMinOutSamplesPerInAmt2Waves(len(in1), bmr.rsmIns)
	out1 := [2][]int16{make([]int16, outLen1), make([]int16, outLen2)}
	out2 := [2][]int16{make([]int16, outLen1), make([]int16, outLen2)}
	return in1, in2, out1, out2
}

func (bmr BenchmarkerBatch) resample(in1, in2 []int16, out1, out2 [2][]int16) {
	var err1, err2, err3 error
	bmr.b.ResetTimer()

	for i := 0; i < bmr.b.N; i++ {
		err1 = bmr.rsm.AddBatch(in1)
		err2 = bmr.rsm.GetBatchFirstWave(out1[0])
		err3 = bmr.rsm.GetBatchSecondWave(out1[1])
	}

	bmr.b.StopTimer()

	bmr.chkErr(err1)
	bmr.chkErr(err2)
	bmr.chkErr(err3)

	err1 = bmr.rsm.AddBatch(in2)
	err2 = bmr.rsm.GetBatchFirstWave(out2[0])
	err3 = bmr.rsm.GetBatchSecondWave(out2[1])

	bmr.chkErr(err1)
	bmr.chkErr(err2)
	bmr.chkErr(err3)
}

func (bmr BenchmarkerBatch) chkErr(err error) {
	if err != nil {
		bmr.b.Error(err)
		bmr.b.FailNow()
	}
}

func benchBatchResampler(rsmT goresampler.Resampler2WavesT, inRate, outRate1, outRate2 int, b *testing.B) {
	bmr := BenchmarkerBatch{}.New(rsmT, inRate, outRate1, outRate2, b)
	bmr.resample(bmr.setup())
}

func BenchmarkBatchSpline11025_8000(b *testing.B) { benchBatchResampler(10, 11025, 8000, 16000, b) }
func BenchmarkBatchSpline16000_8000(b *testing.B) { benchBatchResampler(10, 16000, 8000, 16000, b) }
func BenchmarkBatchSpline44100_8000(b *testing.B) { benchBatchResampler(10, 44100, 8000, 16000, b) }
func BenchmarkBatchSpline48000_8000(b *testing.B) { benchBatchResampler(10, 48000, 8000, 16000, b) }
