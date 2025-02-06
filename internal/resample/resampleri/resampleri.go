package resampleri

type Resampler interface {
	Resample([]int16, []int16) error
	CalcNeedSamplesPerOutAmt(int) int
	CalcOutSamplesPerInAmt(int) int
}
