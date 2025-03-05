package goresampler

type ResamplerSpline2Waves struct {
	rsm1 ResamplerSpline
	rsm2 ResamplerSpline
}

/*
returns configured resampler

if you use New with last arg maxErrRateP=nil - ignore ok value if err doesn't matter (but it can't be large)

try to find batch input amt to have less err (0..1) rate than given maxErrRateP
if failed to find such batch to fit maxErrRate,  second arg is false, otherwise true (but even with false, resampler is fine to use)
*/
func NewResamplerSpline2Waves(inRate, outRate1, outRate2 int, maxErrRateP *float64) (ResamplerSpline2Waves, bool) {
	rsm1, ok1 := NewResamplerSpline(inRate, outRate1, maxErrRateP)
	rsm2, ok2 := NewResamplerSpline(inRate, outRate2, maxErrRateP)
	return ResamplerSpline2Waves{rsm1, rsm2}, ok1 && ok2
}

func (sw ResamplerSpline2Waves) CalcNeedSamplesPerOutAmt(outAmt1, outAmt2 int) int {
	return max(sw.rsm1.CalcNeedSamplesPerOutAmt(outAmt1), sw.rsm2.CalcNeedSamplesPerOutAmt(outAmt2))
}

// not really need so strict - like inAmt % sw.batchInAmt == 0 , but it's garanted
func (sw ResamplerSpline2Waves) calcOutSamplesPerInAmt(inAmt int) (int, int) {
	return sw.rsm1.calcOutSamplesPerInAmt(inAmt), sw.rsm2.calcOutSamplesPerInAmt(inAmt)
}

func (rsm ResamplerSpline2Waves) CalcInOutSamplesPerOutAmt(outAmt1, outAmt2 int) (int, int, int) {
	in1 := rsm.rsm1.CalcNeedSamplesPerOutAmt(outAmt1)
	in2 := rsm.rsm2.CalcNeedSamplesPerOutAmt(outAmt2)
	in := max(in1, in2)
	return in, rsm.rsm1.calcOutSamplesPerInAmt(in), rsm.rsm2.calcOutSamplesPerInAmt(in)
}

func (sw ResamplerSpline2Waves) Resample(in, out1, out2 []int16) error {
	{
		cIn, cOut1, cOut2 := sw.CalcInOutSamplesPerOutAmt(len(out1), len(out2))
		if cIn != len(in) || cOut1 != len(out1) || cOut2 != len(out2) {
			return ErrIncorrectInLen
		}
	}

	sw.rsm1.preResample(in, len(out1))
	sw.rsm2.preResample(in, len(out2))
	spl := sw.rsm1.calcSpline()
	sw.rsm1.resample(spl)
	sw.rsm2.resample(spl)
	sw.rsm1.postResample(out1)
	sw.rsm2.postResample(out2)
	return nil
}

func (rsm ResamplerSpline2Waves) Reset() { // TODO logically should be empty but not tested
	panic("UNIMPLEMENTED")
}
