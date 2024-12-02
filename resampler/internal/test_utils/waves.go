package testutils

import (
	"errors"
	"fmt"
	"math"
	"resampler/internal/utils"
)

type SinWave struct {
	leftB        float64
	rightB       float64
	inResRate    int
	outResRate   int
	inSampleAmt  int
	outSampleAmt int
}

func (SinWave) Init(leftB, rightB float64, inResRate, outResRate int) TestWave {
	inSampleAmt := int(math.Floor((rightB - leftB) * float64(inResRate))) // no math behind, but floor not have latger segment in theory
	outSampleAmt := int(math.Floor((rightB - leftB) * float64(outResRate)))
	return SinWave{leftB, rightB, inResRate, outResRate, inSampleAmt, outSampleAmt}
}

func (SinWave) New(_ int) {}

func (sw SinWave) InLen() int {
	return sw.inSampleAmt
}

func (sw SinWave) OutLen() int {
	return sw.outSampleAmt
}

func (sw SinWave) GetIn(ind int) (int16, error) {
	if ind >= sw.InLen() {
		return 0, errors.New("out of bounds")
	}

	x := float64(ind) * (sw.rightB - sw.leftB) / float64(sw.inResRate)
	return utils.FFloat16ToS16(math.Sin(x)), nil
}

func (sw SinWave) GetOut(ind int) (int16, error) {
	if ind >= sw.OutLen() {
		return 0, errors.New("out of bounds")
	}

	x := float64(ind) * (sw.rightB - sw.leftB) / float64(sw.outResRate)
	return utils.FFloat16ToS16(math.Sin(x)), nil
}

func (sw SinWave) String() string {
	return fmt.Sprintf("SinWave:[%v; %v]|f%vto%vsr|%vsec",
		sw.leftB, sw.rightB, sw.inResRate, sw.outResRate, sw.inSampleAmt/sw.inResRate)
}
