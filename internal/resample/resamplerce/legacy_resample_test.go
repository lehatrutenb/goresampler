package resamplerce_test

import (
	"errors"
	"fmt"
	resamplerce "resampler/internal/resample/resamplerce"
	resampleri "resampler/internal/resample/resampleri"
	testutils "resampler/internal/test_utils"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

type ResamplerLTest struct {
	inRate    int
	outRate   int
	rsm       resampleri.Resampler
	resampled []int16
}

func (ResamplerLTest) New(inRate, outRate int) *ResamplerLTest {
	rsm, err := resamplerce.NewAutoResampler(inRate, outRate)
	if err != nil {
		panic(err)
	}
	res := new(ResamplerLTest)
	*res = ResamplerLTest{inRate, outRate, rsm, nil}
	return res
}

func (rsm ResamplerLTest) Copy() testutils.TestResampler {
	return ResamplerLTest{}.New(rsm.inRate, rsm.outRate)
}
func (rsm ResamplerLTest) String() string {
	return fmt.Sprintf("%d_to_%d_resamplerL", rsm.inRate, rsm.outRate)
}
func (rsm *ResamplerLTest) Resample(inp []int16) error {
	rsm.resampled = make([]int16, rsm.rsm.CalcOutSamplesPerInAmt(len(inp)))
	return rsm.rsm.Resample(inp, rsm.resampled)
}
func (rsm ResamplerLTest) OutLen() int {
	return len(rsm.resampled)
}
func (rsm ResamplerLTest) OutRate() int {
	return rsm.outRate
}
func (rsm ResamplerLTest) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func TestResample8To16L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 8000, 16000), ResamplerLTest{}.New(8000, 16000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample11To8L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 11000, 8000), ResamplerLTest{}.New(11000, 8000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample11To16L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 11000, 16000), ResamplerLTest{}.New(11000, 16000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample16To8L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 16000, 8000), ResamplerLTest{}.New(16000, 8000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample44To8L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 44000, 8000), ResamplerLTest{}.New(44000, 8000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample44To16L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 44000, 16000), ResamplerLTest{}.New(44000, 16000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To8L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 48000, 8000), ResamplerLTest{}.New(48000, 8000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To16L_SinWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.New(0, 55, 48000, 16000), ResamplerLTest{}.New(48000, 16000), 10, t, testutils.TestOpts{}.NewDefault())
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
