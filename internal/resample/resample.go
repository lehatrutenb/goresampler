package resample

import (
	"errors"
	"resampler/internal/resample/resampleri"

	"golang.org/x/exp/slices"
)

var (
	ErrNotEnoughSamples = errors.New("need more samples to get that size of batch")
)

type ResamplerBatch struct {
	in  []int16
	out []int16
	rsm resampleri.Resampler
}

func New(rsm resampleri.Resampler) ResamplerBatch {
	return ResamplerBatch{make([]int16, 0), make([]int16, 0), rsm}
}

func (rsm *ResamplerBatch) AddBatch(in []int16) error {
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

func (rsm *ResamplerBatch) resampleMore(minRsmAmt int) error {
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

func (rsm *ResamplerBatch) GetLargeBatch(out *[]int16) error {
	bLen := len(*out)
	if bLen > len(rsm.out) {
		rsm.resampleMore(bLen - len(rsm.out))
	}
	*out = rsm.out[:bLen]
	rsm.out = rsm.out[bLen:]
	return nil
}

func (rsm *ResamplerBatch) GetBatch(out []int16) error {
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

func (rsm ResamplerBatch) UnresampledInAmt() int {
	return len(rsm.in)
}
