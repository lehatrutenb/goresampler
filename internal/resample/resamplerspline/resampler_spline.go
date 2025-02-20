package resamplerspline

import (
	"math"

	"github.com/lehatrutenb/go_resampler/internal/resampleutils"
	"github.com/lehatrutenb/go_resampler/internal/utils"
)

const minInAmt = 30 //  to reduce infl from edges to spline
const baseTimeErrRate = 1e-6

type WaveInt16 struct {
	data  []int16
	sRate int
}

type borderCond struct {
	c_0, c_n      float32
	mu_0, lamda_n float32
	md_0, md_n    float32 // main diag
}

type ResamplerSpline struct {
	in          []float32
	inRate      int
	outRate     int
	bc          borderCond
	batchInAmt  int
	batchOutAmt int
}

/*
if you use New with last maxErrRateP=nil - ignore ok value if err doesn't matter (but it can't be large)
return configured resampler

try to find batch input amt to have less err (0..1) rate than given maxErrRateP
if failed to find such batch to fit maxErrRate,  second arg is false, otherwise true (but even with false, resampler is fine to use)
*/
func New(inRate, outRate int, maxErrRateP *float64) (ResamplerSpline, bool) {
	var maxErrRate = baseTimeErrRate
	if maxErrRateP != nil {
		maxErrRate = *maxErrRateP
	}
	bInAmt, bOutAmt, ok := CalcInAmtPerErrRate(maxErrRate, inRate, outRate)
	return ResamplerSpline{inRate: inRate, outRate: outRate, bc: borderCond{0, 0, 0, 0, 2, 2}, batchInAmt: bInAmt, batchOutAmt: bOutAmt}, ok
}

type Spline struct {
	ys   []float32 // f(x) in givens xs
	yds  []float32 // f(x)' in given xs
	step float32   // xs are 0, step, 2*step, ...
}

// have Mx=D where M - three diag A B C where A = [x1] * len(A), C = [x2] * len(C), B = [x3] * len(B)
func solveMatrixEqSimpleDiags(a float32, b float32, c float32, ds []float32, bc borderCond) []float32 {
	/*
	   TODO work with possible divide by 0 - return err as result or better another solution?
	   TODO speed up
	*/

	sz := len(ds) // everywhere size is same so lets make var for it
	xs := make([]float32, sz)
	alphs := make([]float32, sz)
	betths := make([]float32, sz)

	// calc coefs
	alphs[1] = -bc.mu_0 / bc.md_0
	betths[1] = bc.c_0 / bc.md_0
	for ind := 1; ind+1 < sz; ind++ {
		nx := a*alphs[ind] + b
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

func (Spline) New(ys []float32, step float32, bc borderCond) Spline {
	// TODO check to make step int

	yds := func() []float32 { // calc discerete diffs
		var lambda float32 = 1 / 2
		mu := 1 - lambda

		sz := len(ys)
		cs := make([]float32, sz)        // discrete func diffs in xs
		cs[0], cs[sz-1] = bc.c_0, bc.c_n // unused , but to save math correctness
		for ind := 1; ind+1 < sz; ind++ {
			diff := (ys[ind] - ys[ind-1]) / step
			cs[ind] = 3 * diff * (2*lambda - 1) // 3 * lamda * diff - 3 * mu * diff but cut
		}
		return solveMatrixEqSimpleDiags(lambda, 2, mu, cs, bc)
	}()
	return Spline{ys, yds, step}
}

func (sp Spline) calcNewStep(newSt float64, amt int) []float32 {
	newYs := make([]float32, amt)
	var st float64 = float64(sp.step)
	var st2, st3 float64 = st * st, st * st * st
	for ind := 0; ind < amt; ind++ {
		x := float64(ind) * float64(newSt)

		il := min(int32(len(sp.ys)-2), max(0, int32(math.Floor(float64(x/float64(sp.step))))))
		ir := il + 1
		l := float64(sp.step) * float64(il)
		r := float64(sp.step) * float64(ir)

		ld := x - l
		rd := x - r

		first := float64(sp.yds[il])*ld*rd*rd/st2 + float64(sp.ys[il])*(2*ld*rd*rd/st3+rd*rd/st2)
		second := float64(sp.yds[ir])*ld*ld*rd/st2 + float64(sp.ys[ir])*(-2*ld*ld*rd/st3+ld*ld/st2)
		newYs[ind] = float32(first + second)
	}

	return newYs
}

func rateToSplineStep(rate int) float64 {
	return 1 / float64(rate)
}

func (sw *ResamplerSpline) resample(sp Spline, out *[]float32) {
	*out = sp.calcNewStep(rateToSplineStep(sw.outRate), len(*out))
}

/*
try to find batch input amt to have less err (0..1) rate than given
(check calcs inside)

return false if failes to find such value < 1e5 and best value found
return true if find such value
*/
func CalcInAmtPerErrRate(maxErr float64, inRate int, outRate int) (bInAmt, bOutAmt int, ok bool) {
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

	return bInAmt, bOutAmt, false
}

func (sw ResamplerSpline) CalcNeedSamplesPerOutAmt(outAmt int) int {
	return ((outAmt + sw.batchOutAmt - 1) / sw.batchOutAmt) * sw.batchInAmt
}

// not really need so strict - like inAmt % sw.batchInAmt == 0 , but it's garanted
func (sw ResamplerSpline) calcOutSamplesPerInAmt(inAmt int) int {
	return (inAmt / sw.batchInAmt) * sw.batchOutAmt
}

func (rsm ResamplerSpline) CalcInOutSamplesPerOutAmt(outAmt int) (int, int) {
	in := rsm.CalcNeedSamplesPerOutAmt(outAmt)
	return in, rsm.calcOutSamplesPerInAmt(in)
}

func (sw ResamplerSpline) Resample(in, out []int16) error {
	sw.in = utils.AS16ToFloat(in)
	outF := make([]float32, len(out))
	sw.resample(Spline{}.New(sw.in, float32(rateToSplineStep(sw.inRate)), sw.bc), &outF)
	copy(out, utils.AFloatToS16(outF))
	return nil
}

func (rsm ResamplerSpline) Reset() { // TODO logically should be empty but not tested
	panic("UNIMPLEMENTED")
}
