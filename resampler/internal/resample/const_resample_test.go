package resample_test

import (
	"errors"
	"resampler/internal/resample"
	testutils "resampler/internal/test_utils"
)

type resampler11To8 struct {
	resampled []int16
}

func (resampler11To8) Copy() testutils.TestResampler {
	return new(resampler11To8)
}
func (resampler11To8) String() string {
	return "11000_to_8000_resampler"
}
func (rsm *resampler11To8) Resample(inp []int16) error {
	return resample.Resample11To8(inp, &rsm.resampled)
}
func (rsm resampler11To8) OutLen() int {
	return len(rsm.resampled)
}
func (resampler11To8) OutRate() int {
	return 8000
}
func (rsm resampler11To8) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

/*
func TestResample11025To8RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outR := 8000
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.RealWave{}.New(0, 11025, &outR, &testutils.PATH_TO_BASE_WAVES), testutils.TestResampler(&resampler11To8{}), 10, t, &testutils.TestOpts{false, "../../../plots"})
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
