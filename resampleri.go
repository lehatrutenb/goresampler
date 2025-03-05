package goresampler

const baseTimeErrRate = 1e-6

// Resampler provides user resampler funcs
type Resampler interface {
	// Resample resamples all data from inWave and save result in outWave
	// len(inWave) and len(outWave) must be equal to any pair got as return of Resampler.CalcInOutSamplesPerOutAmt()
	Resample(inWave []int16, outWave []int16) error

	// CalcNeedSamplesPerOutAmt returns min len(inWave) to get at least outAmt samples as outWave
	CalcNeedSamplesPerOutAmt(outAmt int) (inLen int)

	// Calcs len(inWave) and len(outWave) to get at least outAmt samples after resampling
	// it calls CalcNeedSamplesPerOutAmt inside
	CalcInOutSamplesPerOutAmt(outAmt int) (inLen int, outLen int)

	// Reset clears resample state, make it ready to resample another wave
	Reset()

	// calcOutSamplesPerInAmt returns outLen per inLen
	// not want to make that func public cause some resamplers (fft) want to get only correct inAmt
	//  - result of CalcNeedSamplesPerOutAmt
	calcOutSamplesPerInAmt(inAmt int) (outLen int)
}

// Resampler2Waves provides user resampler funcs that resamples simultaneously
type Resampler2Waves interface {
	// Resample resamples all data from inWave and save result in outWave1 and outWave2
	// len(inWave) and len(outWave) must be equal to any pair got as return of Resampler.CalcInOutSamplesPerOutAmt()
	Resample(inWave []int16, outWave1, outWave2 []int16) error

	// CalcNeedSamplesPerOutAmt returns min len(inWave) to get at least outAmt1 and outAmt2 samples as outWave1 and outWave2
	CalcNeedSamplesPerOutAmt(outAmt1, outAmt2 int) (inLen int)

	// Calcs len(inWave) and len(outWave) to get at least outAmt samples after resampling per 2 waves
	// it calls CalcNeedSamplesPerOutAmt inside
	CalcInOutSamplesPerOutAmt(outAmt1, outAmt2 int) (inLen int, outLen1, outLen2 int)

	// Reset clears resample state, make it ready to resample another wave
	Reset()

	// calcOutSamplesPerInAmt returns outLen per inLen
	// not want to make that func public cause some resamplers (fft) want to get only correct inAmt
	//  - result of CalcNeedSamplesPerOutAmt
	calcOutSamplesPerInAmt(inAmt int) (outLen1, outLen2 int)
}
