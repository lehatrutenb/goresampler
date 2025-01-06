package resample

import (
	"errors"
	testutils "resampler/internal/test_utils"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type resamplerSpline struct {
	inRate    int
	outRate   int
	resampled []int16
}

func (resamplerSpline) New(inRate int, outRate int) resamplerSpline {
    return resamplerSpline{inRate, outRate, []int16{}}
}

func (rsm resamplerSpline) Copy() testutils.TestResampler {
    res := new(resamplerSpline)
    *res = rsm.New(rsm.inRate, rsm.outRate)
    return res
}

func (rsm resamplerSpline) String() string {
	return fmt.Sprintf("%d_to_%d_spline_resampler", rsm.inRate, rsm.outRate)
}

func (rsm *resamplerSpline) Resample(inp []int16) error {
	sw := SplineWave{}.New()
	sw.SetInOutWave(inp, rsm.inRate, rsm.outRate)
	sw.ResampleSpline()
	rsm.resampled = sw.GetOutWave()
	return nil
}

func (rsm resamplerSpline) OutLen() int {
	return len(rsm.resampled)
}

func (rsm resamplerSpline) OutRate() int {
    return rsm.outRate
}


func (rsm resamplerSpline) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func TestResampleSpline48To32(t *testing.T) {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 50, 48000, 32000), testutils.TestResampler(&resamplerSpline{inRate: 48000, outRate: 32000}), 10, t)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

/* Not really need anymore? (after test rework)
func TestResampleSpline48To32SmallBorders(t *testing.T) {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 3, 48000, 32000), testutils.TestResampler(&resamplerSpline{inRate: 48000, outRate: 32000}), 10, t)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
*/
