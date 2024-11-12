package testutils

import (
    "errors"
    "math"
    "fmt"
)

// convert to PCM format
func cF32toI32(x float32) int32 {
    return int32(math.Round(float64(x) * float64(1<<31)))
}

func cF64toI32(x float64) int32 {
    return int32(math.Round(x * float64(1<<31)))
}

type SinWave struct {
    leftB float64
    rightB float64
    inResRate int
    outResRate int
    inSampleAmt int
    outSampleAmt int
}

func Create(leftB, rightB float64, inResRate, outResRate, inSampleAmt, outSampleAmt int) TestWave {
    return SinWave {leftB, rightB, inResRate, outResRate, inSampleAmt, outSampleAmt}
}

func (_ SinWave) New(_ int) {}

func (sw SinWave) InLen() int {
    return sw.inSampleAmt
}

func (sw SinWave) OutLen() int {
    return sw.outSampleAmt
}

func (sw SinWave) GetIn(ind int) (error, int32) {
    if ind >= sw.InLen() {
        return errors.New("out of bounds"), 0
    }

    x := float64(ind)*(sw.rightB - sw.leftB)/float64(sw.inResRate)
    return nil, cF64toI32(math.Sin(x))
}

func (sw SinWave) GetOut(ind int) (error, int32) {
    if ind >= sw.OutLen() {
        return errors.New("out of bounds"), 0
    }

    x := float64(ind)*(sw.rightB - sw.leftB)/float64(sw.outResRate)
    return nil, cF64toI32(math.Sin(x))
}

func (sw SinWave) String() string {
    return fmt.Sprintf("SinWave:[%v; %v]|f%vto%vsr|%vsec",
        sw.leftB, sw.rightB, sw.inResRate, sw.outResRate, sw.inSampleAmt/sw.inResRate)
}
