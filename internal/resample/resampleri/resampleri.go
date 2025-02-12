package resampleri

type Resampler interface {
	Resample([]int16, []int16) error
	CalcInOutSamplesPerOutAmt(int) (int, int) // in, out
	CalcNeedSamplesPerOutAmt(int) int
	// CalcOutSamplesPerInAmt(int) int // not want to make that func public cause some rasmplers (fft) want to get only correct inAmt - that returned CalcNeedSamplesPerOutAmt
}

func GetSecondReturnedVal[T any](_, b T) T {
	return b
}
