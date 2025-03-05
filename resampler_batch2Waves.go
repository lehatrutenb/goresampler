package goresampler

import (
	"golang.org/x/exp/slices"
)

// ResampleBatch2Waves provides resampling within push (GetBatch/GetLargeBatch) and pull (AddBatch) from input to 2 waves with their own rates
type ResampleBatch2Waves struct {
	in   []int16         // buffered input wave, not yet resampled
	out1 []int16         // buffered first output wave, not yet pulled
	out2 []int16         // buffered secound output wave, not yet pulled
	rsm  Resampler2Waves // resampler that will resample
}

func NewResampleBatch2Waves(rsm Resampler2Waves) ResampleBatch2Waves {
	return ResampleBatch2Waves{make([]int16, 0), make([]int16, 0), make([]int16, 0), rsm}
}

// AddBatch appends given in (input wave) to in buffer
func (rsm *ResampleBatch2Waves) AddBatch(in []int16) error {
	rsm.in = append(rsm.in, in...)
	return nil
}

/*
Why
out = rsm.out[:len(out)]
is fine?

1. It won't be freed by gc, even if later I resize out = out[x:]
2. We won't access that data more
3. GC will free it when ~ out slice is no longer in use

use that func when len(out) is large (~ >1e5)
*/

func (rsm *ResampleBatch2Waves) resampleMore(minRsmAmt1, minRsmAmt2 int) error {
	inAmt, outAmt1, outAmt2 := rsm.rsm.CalcInOutSamplesPerOutAmt(minRsmAmt1, minRsmAmt2)
	if inAmt > len(rsm.in) {
		return ErrNotEnoughSamples
	}

	curOutLen1 := len(rsm.out1)
	curOutLen2 := len(rsm.out2)
	rsm.out1 = slices.Grow(rsm.out1, outAmt1)[:len(rsm.out1)+outAmt1]
	rsm.out2 = slices.Grow(rsm.out2, outAmt2)[:len(rsm.out2)+outAmt2] // not want to use loop there 1. slower 2. not really makes code prettier
	rsm.rsm.Resample(rsm.in[:inAmt], rsm.out1[curOutLen1:curOutLen1+outAmt1], rsm.out2[curOutLen2:curOutLen2+outAmt2])
	rsm.in = rsm.in[inAmt:]
	return nil
}

// GetLargeBatch tries to fill out slice with already resampled and resamples if need
//
// returns ErrNotEnoughSamples if in not large enough to get len(out)
// resampler state after ErrNotEnoughSamples is is not broken - it's expected to get such err
//
// don't want to return ok bool cause out is not filling on err - so get more user attention
//
// Difference between GetBatch and GetLargeBatch just in way to fill out
func (rsm *ResampleBatch2Waves) GetLargeBatchFirstWave(out *[]int16) error {
	bLen := len(*out)
	if bLen > len(rsm.out1) {
		if err := rsm.resampleMore(bLen-len(rsm.out1), 0); err != nil {
			return err
		}
	}
	*out = rsm.out1[:bLen]
	rsm.out1 = rsm.out1[bLen:]
	return nil
}
func (rsm *ResampleBatch2Waves) GetLargeBatchSecondWave(out *[]int16) error {
	bLen := len(*out)
	if bLen > len(rsm.out2) {
		if err := rsm.resampleMore(0, bLen-len(rsm.out2)); err != nil {
			return err
		}
	}
	*out = rsm.out2[:bLen]
	rsm.out2 = rsm.out2[bLen:]
	return nil
}

// GetBatch tries to fill out slice with already resampled and resamples if need
//
// returns ErrNotEnoughSamples if in not large enough to get len(out)
// resampler state after ErrNotEnoughSamples is is not broken - it's expected to get such err
//
// don't want to return ok bool cause out is not filling on err - so get more user attention
func (rsm *ResampleBatch2Waves) GetBatchFirstWave(out []int16) error {
	bLen := len(out)
	if bLen > len(rsm.out1) {
		if err := rsm.resampleMore(bLen-len(rsm.out1), 0); err != nil {
			return err
		}
	}

	copy(out, rsm.out1)
	rsm.out1 = rsm.out1[bLen:]
	return nil
}
func (rsm *ResampleBatch2Waves) GetBatchSecondWave(out []int16) error {
	bLen := len(out)
	if bLen > len(rsm.out2) {
		if err := rsm.resampleMore(0, bLen-len(rsm.out2)); err != nil {
			return err
		}
	}

	copy(out, rsm.out2)
	rsm.out2 = rsm.out2[bLen:]
	return nil
}

// UnresampledUngetInAmt returns len of input and output buffers of ResampleBatch
// outWaveInd - which output len to take (1 or 2 wave)
//
// in  []int16 - buffered input wave, not yet resampled
//
// out []int16  - buffered output wave, not yet pulled
func (rsm ResampleBatch2Waves) UnresampledUngetInAmt(outWaveInd int) (int, int) {
	if outWaveInd == 1 {
		return len(rsm.in), len(rsm.out1)
	}
	return len(rsm.in), len(rsm.out2)
}
