package testutils

import (
	"errors"
	"fmt"
    "github.com/go-audio/wav"
    "os"
	"math"
	"resampler/internal/utils"
    "github.com/go-audio/audio"
)

type SinWave struct {
	leftB        float64
	rightB       float64
	inResRate    int
	outResRate   int
	inSampleAmt  int
	outSampleAmt int
}

func (SinWave) New(leftB, rightB float64, inResRate, outResRate int) TestWave {
	inSampleAmt := int(math.Floor((rightB - leftB) * float64(inResRate))) // no math behind, but floor not have lager segment in theory
	outSampleAmt := int(math.Floor((rightB - leftB) * float64(outResRate)))
	return SinWave{leftB, rightB, inResRate, outResRate, inSampleAmt, outSampleAmt}
}

func (SinWave) Seed(int) {}

func (sw SinWave) InLen() int {
	return sw.inSampleAmt
}

func (sw SinWave) OutLen() int {
	return sw.outSampleAmt
}

func (sw SinWave) InRate() int {
	return sw.inResRate
}

func (sw SinWave) OutRate() int {
	return sw.outResRate
}

func (SinWave) WithResampled() bool {
    return true
}

func (SinWave) NumChannels() int {
    return 1
}

func (sw SinWave) GetIn(ind int) (int16, error) {
	if ind >= sw.InLen() {
		return 0, errors.New("out of bounds")
	}

	x := float64(ind) * (sw.rightB - sw.leftB) / float64(sw.inResRate)
	return utils.FFloatToS16(math.Sin(x)), nil
}

func (sw SinWave) GetOut(ind int) (int16, error) {
	if ind >= sw.OutLen() {
		return 0, errors.New("out of bounds")
	}

	x := float64(ind) * (sw.rightB - sw.leftB) / float64(sw.outResRate)
	return utils.FFloatToS16(math.Sin(x)), nil
}

func (sw SinWave) String() string {
	return fmt.Sprintf("SinWave:[%v; %v]|f%vto%vsr|%vsec",
		sw.leftB, sw.rightB, sw.inResRate, sw.outResRate, sw.inSampleAmt/sw.inResRate)
}

type RealWave struct {
    fName string
    buf *audio.IntBuffer
}

var PATH_TO_BASE_WAVES = "../../../base_waves/"

// don't want to return err there not to caught it in test code
func (RealWave) New(fInd int, rate int) TestWave {
    var realWaveFiles map[int]string = map[int]string{0:"base1",1:"base2",2:"base3"}

    fName := realWaveFiles[fInd]
    f, err := os.Open(fmt.Sprintf("%s%s/%s_%d.wav", PATH_TO_BASE_WAVES, fName, fName, rate))
    defer f.Close()
    if err != nil {
        panic(err)
    }
    dec := wav.NewDecoder(f)
    dec.ReadInfo()
    intBuff, err := dec.FullPCMBuffer()
    if err != nil {
        panic(err)
    }
    
    if intBuff.SourceBitDepth != 16 {
        panic(errors.New("expected to get wave with int16 inside"))
    }

    return RealWave{fName, intBuff}
}

func (RealWave) Seed(int) {}

func (rw RealWave) InLen() int {
    return rw.buf.NumFrames() * rw.buf.Format.NumChannels  // == len(buf.Data)
}

func (RealWave) OutLen() int {
    panic(errors.New("call unimplemented func"))
    //return -1
}

func (rw RealWave) InRate() int {
	return rw.buf.Format.SampleRate
}

func (RealWave) OutRate() int {
    panic(errors.New("call unimplemented func"))
	//return -1
}

func (rw RealWave) NumChannels() int {
    return rw.buf.Format.NumChannels
}

func (RealWave) WithResampled() bool {
    return false
}

func (rw RealWave) GetIn(ind int) (int16, error) {
	if ind >= rw.InLen() {
		return 0, errors.New("out of bounds")
	}

	return int16(rw.buf.Data[ind]), nil
}

func (RealWave) GetOut(ind int) (int16, error) {
    return -1, errors.New("call unimplemented func")
}

func (rw RealWave) String() string {
	return fmt.Sprintf("RealWave:%s", rw.fName)
}

