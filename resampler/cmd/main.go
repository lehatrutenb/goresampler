package main

import (
	"errors"
	"testing"

	//"log"
	"resampler/internal/resample"
	"resampler/internal/resample/resamplerl"
	testutils "resampler/internal/test_utils"

	//"testing"
	"fmt"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/assert"
)

type resamplerSpline struct {
	inRate    int
	outRate   int
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

type resampler44To8L struct {
	resampled []int16
}

func (resampler44To8L) Copy() testutils.TestResampler {
	return new(resampler44To8L)
}
func (resampler44To8L) String() string {
	return "44100_to_8000_resamplerL"
}
func (rsm *resampler44To8L) Resample(inp []int16) error {
	return resamplerl.Resample44To8L(inp, &rsm.resampled)
}
func (rsm resampler44To8L) OutLen() int {
	return len(rsm.resampled)
}
func (resampler44To8L) OutRate() int {
	return 8000
}
func (rsm resampler44To8L) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func main() {
	t := &testing.T{}
	/*defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()*/
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 44100, 8000), testutils.TestResampler(&resampler44To8L{}), 10, t, &testutils.TestOpts{false, "../../../../plots"})
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("latest/legacy")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
