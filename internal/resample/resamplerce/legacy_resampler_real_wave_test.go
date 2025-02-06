package resamplerce_test

import (
	testutils "resampler/internal/test_utils"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestResample8To16L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 8000, &outRate, nil), 0, 440000), ResamplerLTest{}.New(8000, 16000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample11To8L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 11000, &outRate, nil), 0, 605000), ResamplerLTest{}.New(11000, 8000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample11To16L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 11000, &outRate, nil), 0, 605000), ResamplerLTest{}.New(11000, 16000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample16To8L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 16000, &outRate, nil), 0, 880000), ResamplerLTest{}.New(16000, 8000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample44To8L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 44000, &outRate, nil), 0, 2420000), ResamplerLTest{}.New(44000, 8000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample44To16L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 44000, &outRate, nil), 0, 2420000), ResamplerLTest{}.New(44000, 16000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To8L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 8000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 48000, &outRate, nil), 0, 2640000), ResamplerLTest{}.New(48000, 8000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}

func TestResample48To16L_RealWave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outRate := 16000
	pathToBaseWaves := "../../../../base_waves/"
	_ = outRate
	_ = pathToBaseWaves
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.CutWave{}.New(testutils.RealWave{}.New(0, 48000, &outRate, nil), 0, 2640000), ResamplerLTest{}.New(48000, 16000), 10, t, testutils.TestOpts{}.NewDefault().WithCrSF(true))
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("rsm_const")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
