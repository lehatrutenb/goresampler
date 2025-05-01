package resampleutils

import "math"

func GetInAmtPerOutAmt(inRate, outRate, outAmt int) int {
	return int(math.Ceil(float64(outAmt*inRate) / float64(outRate)))
}

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
func CalcInAmtPerErrRate(maxErr float64, inRate int, outRate int, minInAmt int) (bInAmt, bOutAmt int, ok bool) {
	bInAmt = minInAmt
	bErr := 1e9
	for inAmt := minInAmt; inAmt < 1e5; inAmt++ {
		vMin, vMax := GetMinMaxSmplsAmt(inRate, outRate, int64(inAmt))

		if CheckErrMinMax(vMin, vMax, maxErr) {
			return inAmt, GetOutAmtPerInAmt(inRate, outRate, inAmt), true
		}
		if vMin/vMax < bErr {
			bErr = vMin / vMax
			bInAmt = inAmt
		}
	}

	bOutAmt = GetOutAmtPerInAmt(inRate, outRate, bInAmt)
	return bInAmt, bOutAmt, false
}
