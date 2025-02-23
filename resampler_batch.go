package goresampler

import (
	"errors"

	"golang.org/x/exp/slices"
)

var (
	// ErrNotEnoughSamples indicates that in ResampleBatch not enough buffered data
	// to get requested samples amount
	ErrNotEnoughSamples = errors.New("need more samples to get that size of batch")
)

// ResampleBatch provides resampling within push (GetBatch/GetLargeBatch) and pull (AddBatch)
type ResampleBatch struct {
	in  []int16   // buffered input wave, not yet resampled
	out []int16   // buffered output wave, not yet pulled
	rsm Resampler // resampler that will resample
}

func NewResampleBatch(rsm Resampler) ResampleBatch {
	return ResampleBatch{make([]int16, 0), make([]int16, 0), rsm}
}

// AddBatch appends given in (input wave) to in buffer
func (rsm *ResampleBatch) AddBatch(in []int16) error {
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

func (rsm *ResampleBatch) resampleMore(minRsmAmt int) error {
	inAmt, outAmt := rsm.rsm.CalcInOutSamplesPerOutAmt(minRsmAmt)
	if inAmt > len(rsm.in) {
		return ErrNotEnoughSamples
	}

	curOutLen := len(rsm.out)
	rsm.out = slices.Grow(rsm.out, outAmt)[:len(rsm.out)+outAmt]
	rsm.rsm.Resample(rsm.in[:inAmt], rsm.out[curOutLen:curOutLen+outAmt])
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
func (rsm *ResampleBatch) GetLargeBatch(out *[]int16) error {
	bLen := len(*out)
	if bLen > len(rsm.out) {
		if err := rsm.resampleMore(bLen - len(rsm.out)); err != nil {
			return err
		}
	}
	*out = rsm.out[:bLen]
	rsm.out = rsm.out[bLen:]
	return nil
}

// GetBatch tries to fill out slice with already resampled and resamples if need
//
// returns ErrNotEnoughSamples if in not large enough to get len(out)
// resampler state after ErrNotEnoughSamples is is not broken - it's expected to get such err
//
// don't want to return ok bool cause out is not filling on err - so get more user attention
func (rsm *ResampleBatch) GetBatch(out []int16) error {
	bLen := len(out)
	if bLen > len(rsm.out) {
		if err := rsm.resampleMore(bLen - len(rsm.out)); err != nil {
			return err
		}
	}

	copy(out, rsm.out)
	rsm.out = rsm.out[bLen:]
	return nil
}

// UnresampledUngetInAmt returns len of input and output buffers of ResampleBatch
//
// in  []int16 - buffered input wave, not yet resampled
//
// out []int16  - buffered output wave, not yet pulled
func (rsm ResampleBatch) UnresampledUngetInAmt() (int, int) {
	return len(rsm.in), len(rsm.out)
}
