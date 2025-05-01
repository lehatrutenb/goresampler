package goresampler

import (
	"errors"
	"math"
	"slices"

	"github.com/lehatrutenb/goresampler/internal/resampleutils"
	"github.com/lehatrutenb/goresampler/internal/utils"
)

var (
	// ErrGotIncorrectInOutLen indicates that in new passed windowSize : windowSize%2==1
	ErrGotOddWindowSize = errors.New("got odd window size")
	// ErrGotIncorrectInOutLen indicates that in new passed windowSize, SincMinInAmt : windowSize>SincMinInAmt
	ErrGotWindowSizeLargerSincMinInAmt = errors.New("got windowSize > SincMinInAmt")
	// ErrGotTooLargeErrRateP indicates that with such err it is possible to get bad resampling quality
	ErrGotTooLargeErrRateP = errors.New("got too large maxErrRateP - use ignoreRsmQSpoiling to ignore it")
)

type PrecalcedTablePolicyT int

const (
	PolicyForcePrecalced PrecalcedTablePolicyT = iota
	PolicyForceRuntime
	PolicyAuto
)

// user need such values to better understand algo of sinc resampler
const SincMinInAmt = 128 //  to reduce infl from edges to sinc - min amt of samples to resample
const SincBaseWindowSize = 24
const SincBasePrecalcedTablePolicy = PolicyForcePrecalced

type SincResamplerOptions struct {
	BaseResamplerOptions
	WindowSize           int // care, windowSize % 2 == 0
	MinInAmt             int // care, SincMinInAmt >= windowSize
	PrecalcedTablePolicy PrecalcedTablePolicyT
	ignoreRsmQSpoiling   bool
}

func (opts *SincResamplerOptions) init() {
	opts.Init()
	if opts.WindowSize == 0 {
		opts.WindowSize = SincBaseWindowSize
	}
	if opts.MinInAmt == 0 {
		opts.MinInAmt = SincMinInAmt
	}
}

// set true if want not to throw err on possible large resampling quality spoiling because of large marrErrRateP
// not public not to create it by chance
func (opts SincResamplerOptions) WithIgnoreRsmQSpoiling(newIgnorePolicy bool) SincResamplerOptions {
	opts.ignoreRsmQSpoiling = newIgnorePolicy
	return opts
}

/*
resampler that provides resampling via sinc interpolation
*/
type ResamplerSinc struct {
	in               []float32
	outF             []float32 // care will have cap eq to max needed during resampler lifetime
	sincTable        []float32 // precalced table of sinc coefs grouped by windowSize [...], [...], ... - calc sinc coefs for ind in outBatch for window around [windowSz/2 to left, windowSz/2 to right]
	sincTableLeftInd []int     // precalced indexes in inBatch of start of window (left border) - need to connect inputs&weights
	inRate           int
	outRate          int
	batchInAmt       int
	batchOutAmt      int
	windowSz         int // on vals >= 100 it is better to use ResamplerFFT
}

func (sw *ResamplerSinc) precalcSincTable() {
	sincTable := make([]float32, sw.batchOutAmt*sw.windowSz) // lenOut is always fixed(ws) ! - used in ResamplerSinc.resample
	sincTableLeftInd := make([]int, sw.batchOutAmt)

	invNewSt := float64(sw.outRate)
	invPrevSt := float64(sw.inRate)

	left := -(sw.windowSz / 2) // cur sinc window borders
	right := int(min(int32(-left-1), int32(sw.batchInAmt-1)))
	leftX := float64(left) / invPrevSt
	for outInd := range sincTableLeftInd {
		curX := float64(outInd) / invNewSt
		leftX, left, right = sw.moveIfNeed(leftX, curX, left, right, sw.batchInAmt+sw.windowSz/2-1)

		sincTableLeftInd[outInd] = left
		for fromInd := left; fromInd <= right; fromInd++ {
			weight := float32(sinc((curX - float64(fromInd)/invPrevSt) * invPrevSt))
			sincTable[outInd*sw.windowSz+fromInd-left] = weight
		}
	}

	sw.sincTable = sincTable
	sw.sincTableLeftInd = sincTableLeftInd
}

/*
returns configured resampler

if you use New with last arg opts=nil - ignore first ok value if err in durations doesn't matter (but it can't be large)

try to find batch input amt to have less err (0..1) rate than given maxErrRateP
if failed to find such batch to fit maxErrRate,  second arg is false, otherwise true (but even with false, resampler is still fine to use)
*/
func NewResamplerSinc(inRate, outRate int, opts *SincResamplerOptions) (ResamplerSinc, bool, error) {
	if opts == nil {
		opts = &SincResamplerOptions{}
	}
	opts.init()

	if opts.WindowSize > opts.MinInAmt {
		return ResamplerSinc{}, false, ErrGotWindowSizeLargerSincMinInAmt
	}
	if opts.WindowSize%2 != 0 {
		return ResamplerSinc{}, false, ErrGotOddWindowSize
	}
	if !opts.ignoreRsmQSpoiling && ((opts.MaxErrRateP != nil) && (*opts.MaxErrRateP > BaseTimeErrRate)) {
		return ResamplerSinc{}, false, ErrGotTooLargeErrRateP
	}
	bInAmt, bOutAmt, ok := resampleutils.CalcInAmtPerErrRate(*opts.MaxErrRateP, inRate, outRate, opts.MinInAmt)
	if !(opts.PrecalcedTablePolicy == PolicyForcePrecalced) {
		if (opts.PrecalcedTablePolicy == PolicyForceRuntime) || bOutAmt >= int(1e5) || min(int32(bInAmt), int32(bOutAmt)) < int32(opts.WindowSize) { // check conditions for PreaclcAutoPolicy
			return ResamplerSinc{inRate: inRate, outRate: outRate, batchInAmt: bInAmt, batchOutAmt: bOutAmt, windowSz: opts.WindowSize}, ok, nil
		}
	}

	sw := ResamplerSinc{inRate: inRate, outRate: outRate, batchInAmt: bInAmt, batchOutAmt: bOutAmt, windowSz: opts.WindowSize}
	sw.precalcSincTable()
	return sw, ok, nil
}

func (sw ResamplerSinc) CalcNeedSamplesPerOutAmt(outAmt int64) int64 {
	return ((outAmt + int64(sw.batchOutAmt) - 1) / int64(sw.batchOutAmt)) * int64(sw.batchInAmt)
}

// not really need so strict - like inAmt % sw.batchInAmt == 0 , but it's garanted
func (sw ResamplerSinc) calcOutSamplesPerInAmt(inAmt int64) int64 {
	return (inAmt * int64(sw.batchOutAmt)) / int64(sw.batchInAmt)
}

func (rsm ResamplerSinc) CalcInOutSamplesPerOutAmt(outAmt int64) (int64, int64) {
	in := rsm.CalcNeedSamplesPerOutAmt(outAmt)
	return in, rsm.calcOutSamplesPerInAmt(in)
}

func (sw *ResamplerSinc) preResample(in []int16, outLen int) {
	sw.in = utils.AS16ToFloat(in)
	sw.outF = slices.Grow(sw.outF, outLen)
	sw.outF = sw.outF[:outLen]
}

func sinc(x float64) float64 {
	if math.Abs(x) < 1e-7 {
		return 1
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

func (sw *ResamplerSinc) moveIfNeed(leftX, curX float64, left, right int, maxInd int) (float64, int, int) { // leftX, left, right
	for {
		if math.Abs(leftX-curX) > math.Abs((float64(right+1)/float64(sw.inRate))-curX) && right < maxInd {
			left++
			right++
			leftX = float64(left) / float64(sw.inRate)
		} else {
			break
		}
	}
	return leftX, left, right
}

func (sw *ResamplerSinc) resampleBase() { // base cause fair - without any approx speed ups
	invNewSt := float64(sw.outRate)
	invPrevSt := float64(sw.inRate)

	left := 0 // cur sinc window borders
	right := int(min(int32(sw.windowSz-1), int32(len(sw.in)-1)))
	leftX := float64(left) / invPrevSt

	for i := range len(sw.outF) {
		curX := float64(i) / invNewSt
		leftX, left, right = sw.moveIfNeed(leftX, curX, left, right, len(sw.in)-1)

		for fromInd := left; fromInd <= right; fromInd++ {
			weight := float32(sinc((curX - float64(fromInd)/invPrevSt) * invPrevSt))
			sw.outF[i] += sw.in[fromInd] * weight
		}
	}
}

func (sw *ResamplerSinc) resample() {
	var left, lenIn int32
	for outInd := range len(sw.outF) {
		modInd := int32(outInd % sw.batchOutAmt)
		left = int32(sw.sincTableLeftInd[modInd] + (outInd/sw.batchOutAmt)*sw.batchInAmt)
		lenIn = int32(len(sw.in))
		for i := max(left, -left); i < min(lenIn, int32(sw.windowSz)+left); i++ { // left + i >= 0 as left can (will) be shifted to negatives to overlap prev batch ; i - left < lenOut
			sw.outF[outInd] += sw.in[i] * sw.sincTable[modInd*int32(sw.windowSz)+i-left]
		}
	}
}

func (sw *ResamplerSinc) postResample(out []int16) {
	copy(out, utils.AFloatToS16(sw.outF))
}

func (sw ResamplerSinc) ResampleAll(in, out []int16) error {
	sw.preResample(in, len(out))
	if sw.sincTable != nil {
		sw.resample()
	} else {
		sw.resampleBase()
	}
	sw.postResample(out)
	return nil
}

func (sw ResamplerSinc) Resample(in, out []int16) error {
	{
		cIn, cOut := sw.CalcInOutSamplesPerOutAmt(int64(len(out)))
		if cIn != int64(len(in)) || cOut != int64(len(out)) {
			return ErrIncorrectInLen
		}
	}

	return sw.ResampleAll(in, out)
}

func (rsm ResamplerSinc) Reset() { // currently no state
}
