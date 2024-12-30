package resample

import (
	"errors"
	testutils "resampler/internal/test_utils"
	"testing"

	"github.com/stretchr/testify/assert"
    "fmt"
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
    sw := SplineWave{}.New()
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

func TestResampleSpline48To32(t *testing.T) {
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.Init(0, 5, 48000, 32000), testutils.TestResampler(&resamplerSpline{inRate: 48000, outRate: 32000}), 10)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
