package resampleutils

import "math"

func GetOutAmtPerInAmt(inRate, outRate, inAmt int) int {
	return int(math.Ceil(float64(inAmt*outRate) / float64(inRate)))
}

// TODO rename that func
// calc real amt of samples in output wave and round it
func GetMinMaxSmplsAmt(inRate, outRate int, inAmt int64) (float64, float64) {
	valExp := float64(inAmt*int64(outRate)) / float64(inRate)
	valGet := math.Round(valExp)

	return min(valExp, valGet), max(valExp, valGet)
}

func CheckErrMinMax(minV, maxV, maxErrRate float64) bool {
	return maxV <= minV*(maxErrRate+1)
}
