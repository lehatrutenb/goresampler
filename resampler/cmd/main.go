package main

import (
	"errors"
	//"log"
	"resampler/internal/resample"
	testutils "resampler/internal/test_utils"
	//"testing"
    "fmt"
	//"github.com/stretchr/testify/assert"
)

type resamplerSpline struct {
    inRate int
    outRate int
    resampled []int16
}

func (rsm resamplerSpline) String() string {
	return fmt.Sprintf("%d_to_%d_spline_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerSpline) Resample(inp []int16) error {
    sw := resample.SplineWave{}.New()
    sw.SetInOutWave(inp, rsm.inRate, rsm.outRate)
	sw.ResampleSpline()
	rsm.resampled = sw.GetOutWave()
    return nil
}

func (rsm resamplerSpline) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func main() {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.Init(0, 5, 48000, 32000), testutils.TestResampler(&resamplerSpline{inRate: 48000, outRate: 32000}), 10)
	tObj.Run()
}
