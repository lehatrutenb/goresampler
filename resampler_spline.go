package goresampler

import (
	"math"
	"slices"

	"github.com/lehatrutenb/goresampler/internal/resampleutils"
	"github.com/lehatrutenb/goresampler/internal/utils"
)

const minInAmt = 30 //  to reduce infl from edges to spline

type borderCond struct {
	c_0, c_n      float32
	mu_0, lamda_n float32
	md_0, md_n    float32 // main diag
}

/*
resampler that provides resampling via splines
*/
type ResamplerSpline struct {
	in          []float32
	outF        []float32 // care will have cap eq to max needed during resampler lifetime
	inRate      int
	outRate     int
	bc          borderCond
	batchInAmt  int
	batchOutAmt int
}

/*
returns configured resampler

if you use New with last arg maxErrRateP=nil - ignore ok value if err doesn't matter (but it can't be large)

try to find batch input amt to have less err (0..1) rate than given maxErrRateP
if failed to find such batch to fit maxErrRate,  second arg is false, otherwise true (but even with false, resampler is fine to use)
*/
func NewResamplerSpline(inRate, outRate int, maxErrRateP *float64) (ResamplerSpline, bool) {
	var maxErrRate = baseTimeErrRate
	if maxErrRateP != nil {
		maxErrRate = *maxErrRateP
	}
	bInAmt, bOutAmt, ok := ResamplerSplineCalcInAmtPerErrRate(maxErrRate, inRate, outRate)
	return ResamplerSpline{inRate: inRate, outRate: outRate, bc: borderCond{0, 0, 0, 0, 2, 2}, batchInAmt: bInAmt, batchOutAmt: bOutAmt}, ok
}

type spline struct {
	ys   []float32 // f(x) in givens xs
	yds  []float32 // f(x)' in given xs
	step float64   // xs are 0, 1/step, 1/(2*step), ...
}

func bool2sfloat(b bool) float64 {
	var i int
	if b {
		i = 1
	} else {
		i = 0
	}
	return (1-float64(i))*2 - 1
}

// have Mx=D where M - three diag A B C where A = [x1] * len(A), C = [x2] * len(C), B = [x3] * len(B)
func solveMatrixEqSimpleDiags(a float32, b float32, c float32, ds []float32, bc borderCond) []float32 {
	sz := len(ds) // everywhere size is same so lets make var for it
	xs := make([]float32, sz)
	alphs := make([]float32, sz)
	betths := make([]float32, sz)

	// calc coefs
	alphs[1] = -bc.mu_0 / bc.md_0
	betths[1] = bc.c_0 / bc.md_0
	for ind := 1; ind+1 < sz; ind++ {
		cur := float64(a*alphs[ind] + b)
		nx := float32(max(math.Abs(cur), 1e-5) * bool2sfloat(math.Signbit(cur))) // protect from zero div
		alphs[ind+1] = -c / nx
		betths[ind+1] = (ds[ind] - a*betths[ind]) / nx
	}
	//calc xs
	xs[sz-1] = (bc.c_n - bc.lamda_n*betths[sz-1]) / (bc.lamda_n*alphs[sz-1] + bc.md_n)
	for ind := sz - 1; ind > 0; ind-- {
		xs[ind-1] = alphs[ind]*xs[ind] + betths[ind]
	}

	return xs
}

func (spline) new(ys []float32, step float64, bc borderCond) spline {
	yds := func() []float32 { // calc discerete diffs
		var lambda float32 = 1.0 / 2
		mu := 1 - lambda

		sz := len(ys)
		cs := make([]float32, sz)        // discrete func diffs in xs
		cs[0], cs[sz-1] = bc.c_0, bc.c_n // unused , but to save math correctness
		for ind := 1; ind+1 < sz; ind++ {
			diff := (ys[ind] - ys[ind-1]) * float32(step)
			cs[ind] = 3 * diff * (2*lambda - 1) // 3 * lamda * diff - 3 * mu * diff but cut
		}
		return solveMatrixEqSimpleDiags(lambda, 2, mu, cs, bc)
	}()
	return spline{ys, yds, step}
}

func (sp spline) calcNewStep(invNewSt float64, amt int) []float32 {
	newYs := make([]float32, amt)
	var st = sp.step
	var st2, st3 float64 = st * st, st * st * st
	for ind := 0; ind < amt; ind++ {
		x := float64(ind) / invNewSt

		il := min(int32(len(sp.ys)-2), max(0, int32(math.Floor(x*sp.step))))
		ir := il + 1
		l := float64(il) / sp.step
		r := float64(ir) / sp.step

		ld := x - l
		rd := x - r

		first := float64(sp.yds[il])*ld*rd*rd*st2 + float64(sp.ys[il])*(2*ld*rd*rd*st3+rd*rd*st2)
		second := float64(sp.yds[ir])*ld*ld*rd*st2 + float64(sp.ys[ir])*(-2*ld*ld*rd*st3+ld*ld*st2)
		newYs[ind] = float32(first + second)
	}

	return newYs
}

func rateToSplineStep(rate int) float64 {
	return 1 / float64(rate)
}

/*
try to find batch input amt to have less err (0..1) rate than given

Calculations:

	maxErr = given err rate
	valExp = inSamplesAmt*outRate / inRate
	valGet = math.Round(valExp)
	minV = min(valExp, valGet)
	maxV = max(valExp, valGet)
	minV*(maxErr+1) >= maxV

return false if failes to find such value < 1e5 and best value found
return true if find such value
*/
func ResamplerSplineCalcInAmtPerErrRate(maxErr float64, inRate int, outRate int) (bInAmt, bOutAmt int, ok bool) {
	bInAmt = minInAmt
	bOutAmt = resampleutils.GetOutAmtPerInAmt(inRate, outRate, bInAmt)
	bErr := 1e9
	for inAmt := minInAmt; inAmt < 1e5; inAmt++ {
		vMin, vMax := resampleutils.GetMinMaxSmplsAmt(inRate, outRate, int64(inAmt))

		if resampleutils.CheckErrMinMax(vMin, vMax, maxErr) {
			return inAmt, resampleutils.GetOutAmtPerInAmt(inRate, outRate, inAmt), true
		}
		if vMin/vMax < bErr {
			bErr = vMin / vMax
			bInAmt = inAmt
		}
	}

	bOutAmt = resampleutils.GetOutAmtPerInAmt(inRate, outRate, bInAmt)
	return bInAmt, bOutAmt, false
}

func (sw ResamplerSpline) CalcNeedSamplesPerOutAmt(outAmt int) int {
	return ((outAmt + sw.batchOutAmt - 1) / sw.batchOutAmt) * sw.batchInAmt
}

// not really need so strict - like inAmt % sw.batchInAmt == 0 , but it's garanted
func (sw ResamplerSpline) calcOutSamplesPerInAmt(inAmt int) int {
	return (inAmt * sw.batchOutAmt) / sw.batchInAmt
}

func (rsm ResamplerSpline) CalcInOutSamplesPerOutAmt(outAmt int) (int, int) {
	in := rsm.CalcNeedSamplesPerOutAmt(outAmt)
	return in, rsm.calcOutSamplesPerInAmt(in)
}

func (sw *ResamplerSpline) preResample(in []int16, outLen int) {
	sw.in = utils.AS16ToFloat(in)
	sw.outF = slices.Grow(sw.outF, outLen)
	sw.outF = sw.outF[:outLen]
}

func (sw *ResamplerSpline) resample(sp spline) {
	sw.outF = sp.calcNewStep(float64(sw.outRate), len(sw.outF))
}

func (sw *ResamplerSpline) postResample(out []int16) {
	copy(out, utils.AFloatToS16(sw.outF))
}

func (sw *ResamplerSpline) calcSpline() spline {
	return spline{}.new(sw.in, float64(sw.inRate), sw.bc)
}

func (sw ResamplerSpline) ResampleAll(in, out []int16) error {
	sw.preResample(in, len(out))
	sw.resample(sw.calcSpline())
	sw.postResample(out)
	return nil
}

func (sw ResamplerSpline) ResampleAll(in, out []int16) error {
	sw.preResample(in, len(out))
	sw.resample(sw.calcSpline())
	sw.postResample(out)
	return nil
}

func (sw ResamplerSpline) Resample(in, out []int16) error {
	{
		cIn, cOut := sw.CalcInOutSamplesPerOutAmt(len(out))
		if cIn != len(in) || cOut != len(out) {
			return ErrIncorrectInLen
		}
	}

	return sw.ResampleAll(in, out)
}

func (rsm ResamplerSpline) Reset() { // currently no state
}
