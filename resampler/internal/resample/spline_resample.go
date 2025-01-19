package resample

import (
	"math"
	"resampler/internal/utils"
)

type WaveInt16 struct {
	data  []int16
	sRate int
}

type borderCond struct {
	c_0, c_n      float32
	mu_0, lamda_n float32
	md_0, md_n    float32 // main diag
}

type SplineWave struct {
	in      []float32
	out     []float32
	inRate  int
	outRate int
	bc      borderCond
}

func (SplineWave) New() SplineWave {
	return SplineWave{bc: borderCond{0, 0, 0, 0, 2, 2}}
}

// splitted to add ability to speed up if have many runs

func (sw *SplineWave) SetInWave(in []int16, rate int) {
	sw.in = utils.AS16ToFloat(in)
	sw.inRate = rate
}

func (sw *SplineWave) SetOutWave(rate int) {
	sw.outRate = rate
	sw.out = make([]float32, len(sw.in)*rate/sw.inRate)
}

func (sw *SplineWave) SetInOutWave(in []int16, inRate int, outRate int) {
	sw.SetInWave(in, inRate)
	sw.SetOutWave(outRate)
}

func (sw SplineWave) GetOutWave() []int16 {
	return utils.AFloatToS16(sw.out)
}

type Spline struct {
	ys   []float32 // f(x) in givens xs
	yds  []float32 // f(x)' in given xs
	step float32   // xs are 0, step, 2*step, ...
}

// don't really need that func currently cause step is same between all xs, but written not to return to it later
// have Mx=D where M - three diag A B C - "метод прогонки"
func _solveMatrixEq(as []float32, bs []float32, cs []float32, ds []float32) []float32 { // TODO work with possible divide by 0 - return err as result or better another solution?
	sz := len(as) // everywhere size is same so lets make var for it
	xs := make([]float32, sz)
	alphs := make([]float32, sz)
	betths := make([]float32, sz)

	// calc coefs
	alphs[1] = -cs[0] / bs[0]
	betths[1] = ds[0] / bs[0]
	for ind := 1; ind+1 < sz; ind++ {
		nx := as[ind]*alphs[ind] + bs[ind]
		alphs[ind+1] = -cs[ind] / nx
		betths[ind+1] = (ds[ind] - as[ind]*betths[ind]) / nx
	}
	//calc xs
	xs[sz-1] = (ds[sz-1] - as[sz-1]*betths[sz-1]) / (as[sz-1]*alphs[sz-1] + bs[sz-1])
	for ind := sz - 1; ind > 0; ind-- {
		xs[ind-1] = alphs[ind]*xs[ind] + betths[ind]
	}

	return xs
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

func (_ Spline) New(ys []float32, step float32, bc borderCond) Spline {
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

func (sp Spline) _calcNewY(x float32) float32 {
	il := min(int32(len(sp.ys)-2), max(0, int32(math.Floor(float64(x/sp.step)))))
	ir := il + 1
	l := sp.step * float32(il)
	r := sp.step * float32(ir)

	ld := x - l
	rd := x - r

	st := sp.step
	st2 := st * st

	first := sp.yds[il]*ld*rd*rd/st2 + sp.ys[il]*(2*ld*rd*rd/(st2*st)+rd*rd/st2)
	second := sp.yds[ir]*ld*ld*rd/st2 + sp.ys[ir]*(-2*ld*ld*rd/(st2*st)+ld*ld/st2)
	return first + second
}

func (sp Spline) calcNewStep(newSt float32, amt int) []float32 {
	newYs := make([]float32, amt)
	var st float32 = sp.step
	var st2, st3 float32 = st * st, st * st * st
	for ind := 0; ind < amt; ind++ {
		x := float32(ind) * newSt

		il := min(int32(len(sp.ys)-2), max(0, int32(math.Floor(float64(x/sp.step)))))
		ir := il + 1
		l := sp.step * float32(il)
		r := sp.step * float32(ir)

		ld := x - l
		rd := x - r

		first := sp.yds[il]*ld*rd*rd/st2 + sp.ys[il]*(2*ld*rd*rd/st3+rd*rd/st2)
		second := sp.yds[ir]*ld*ld*rd/st2 + sp.ys[ir]*(-2*ld*ld*rd/st3+ld*ld/st2)
		newYs[ind] = first + second
	}
	return newYs
}

func rateToSplineStep(rate int) float32 {
	return 1 / float32(rate)
}

func (sw *SplineWave) resampleSpline(sp Spline) {
	sw.out = sp.calcNewStep(rateToSplineStep(sw.outRate), len(sw.out))
}

// divided to speed up if want to convert to many rates - to build spline once

func (sw *SplineWave) ResampleSpline() {
	sw.resampleSpline(Spline{}.New(sw.in, rateToSplineStep(sw.inRate), sw.bc))
}
