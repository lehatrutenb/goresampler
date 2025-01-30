package testutils

import (
	"errors"
	"fmt"
	"math"
	"os"
	"resampler/internal/utils"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
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
	fName  string
	inBuf  *audio.IntBuffer
	outBuf *audio.IntBuffer
}

var PATH_TO_BASE_WAVES = "../../../base_waves/"

// changed panic to fatal to catch in tests
func parseWaveOrSkip(fName *string) *audio.IntBuffer {
	if fName == nil {
		return nil
	}

	f, err := os.Open(*fName)
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

	return intBuff
}

// don't want to return err there not to caught it in test code
func (RealWave) New(fInd int, inRate int, outRate *int, pathToBaseWaves *string) TestWave {
	if pathToBaseWaves == nil {
		pathToBaseWaves = new(string)
		*pathToBaseWaves = PATH_TO_BASE_WAVES
	}
	var realWaveFiles map[int]string = map[int]string{0: "base1", 1: "base2", 2: "base3", 3: "base4"}

	fName := realWaveFiles[fInd]
	fInName := fmt.Sprintf("%s%s/%s_%d.wav", *pathToBaseWaves, fName, fName, inRate)
	var fOutName *string = nil
	if outRate != nil {
		fOutName = new(string)
		*fOutName = fmt.Sprintf("%s%s/%s_%d.wav", *pathToBaseWaves, fName, fName, *outRate)
	}

	return RealWave{fName, parseWaveOrSkip(&fInName), parseWaveOrSkip(fOutName)}
}

func (RealWave) Seed(int) {}

func (rw RealWave) InLen() int {
	return len(rw.inBuf.Data) // == rw.inBuf.NumFrames() * rw.inBuf.Format.NumChannels
}

func (rw RealWave) OutLen() int {
	if rw.outBuf == nil {
		panic(errors.New("call unimplemented func"))
	}

	return rw.outBuf.NumFrames() * rw.outBuf.Format.NumChannels // == len(buf.Data)
}

func (rw RealWave) InRate() int {
	return rw.inBuf.Format.SampleRate
}

func (rw RealWave) OutRate() int {
	if rw.outBuf == nil {
		panic(errors.New("call unimplemented func"))
	}
	return rw.outBuf.Format.SampleRate
}

func (rw RealWave) NumChannels() int {
	return rw.inBuf.Format.NumChannels
}

func (rw RealWave) WithResampled() bool {
	return rw.outBuf != nil
}

func (rw RealWave) GetIn(ind int) (int16, error) {
	if ind >= rw.InLen() {
		return 0, errors.New("out of bounds")
	}

	return int16(rw.inBuf.Data[ind]), nil
}

func (rw RealWave) GetOut(ind int) (int16, error) {
	if rw.outBuf == nil {
		panic(errors.New("call unimplemented func"))
	}
	if ind >= rw.OutLen() {
		return 0, errors.New("out of bounds")
	}

	return int16(rw.outBuf.Data[ind]), nil
}

func (rw RealWave) String() string {
	return fmt.Sprintf("RealWave:%s_%d", rw.fName, rw.InRate())
}

type CutWave struct {
	tw      TestWave
	prefCut int
	cutAmt  int
}

/*
prefCut - amt of samples to cut at the beginning
cutAmt - amt of samples to save after prefCut (not to cut)
*/
func (CutWave) New(w TestWave, prefCut int, cutAmt int) TestWave {
	res := CutWave{tw: w, prefCut: prefCut, cutAmt: cutAmt}
	// why use float there - want to cut not only perfect dividable waves but with errors too ; count ceil to be sure that math round error won't cause overflow
	if prefCut*w.NumChannels()+res.InLen() > w.InLen() || int(math.Ceil(float64(prefCut*w.NumChannels())*float64(res.tw.OutRate())/float64(res.tw.InRate())))+res.OutLen() > w.OutLen() {
		panic("got incorrect cut params - too large for wave len")
	}
	return res
}

func (CutWave) Seed(int) {}

func (rw CutWave) InLen() int {
	return rw.cutAmt * rw.tw.NumChannels()
}

func (rw CutWave) OutLen() int {
	return int(math.Floor(float64(rw.InLen()) * float64(rw.tw.OutRate()) / float64(rw.tw.InRate())))
}

func (rw CutWave) InRate() int {
	return rw.tw.InRate()
}

func (rw CutWave) OutRate() int {
	return rw.tw.OutRate()
}

func (rw CutWave) NumChannels() int {
	return rw.tw.NumChannels()
}

func (rw CutWave) WithResampled() bool {
	return rw.tw.WithResampled()
}

func (rw CutWave) GetIn(ind int) (int16, error) {
	pref := rw.prefCut * rw.tw.NumChannels()
	if ind >= rw.InLen() {
		return 0, errors.New("out of bounds")
	}

	return rw.tw.GetIn(pref + ind)
}

func (rw CutWave) GetOut(ind int) (int16, error) {
	pref := int(math.Round(float64(rw.prefCut*rw.tw.NumChannels()) * float64(rw.tw.OutRate()) / float64(rw.tw.InRate())))
	if ind >= rw.OutLen() {
		return 0, errors.New("out of bounds")
	}

	return rw.tw.GetOut(pref + ind)
}

func (rw CutWave) String() string {
	return fmt.Sprintf("CutWave:%s_[%d:%d]", rw.tw.String(), rw.prefCut, rw.prefCut+rw.cutAmt)
}
