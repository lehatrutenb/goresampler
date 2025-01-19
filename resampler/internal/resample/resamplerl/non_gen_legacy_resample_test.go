package resamplerl_test

/*
func TestResample11To8SinWaveL(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outR := 8000
	pathToBaseWaves := "../" + testutils.PATH_TO_BASE_WAVES
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.RealWave{}.New(3, 11000, &outR, &pathToBaseWaves), testutils.TestResampler(&resampler11To8L{}), 10, t, &testutils.TestOpts{false, "../../../../plots"})
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("latest/legacy")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
*/
/*
func TestResample11025To8RealWaveL(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	outR := 8000
	pathToBaseWaves := "../" + testutils.PATH_TO_BASE_WAVES
	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.RealWave{}.New(0, 11025, &outR, &pathToBaseWaves), testutils.TestResampler(&resampler11To8L{}), 10, t, &testutils.TestOpts{true, "../../../../plots"})
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		t.Error(err)
	}
	err = tObj.Save("latest/legacy")
	if !assert.NoError(t, err, "failed to save test results") {
		t.Error(err)
	}
}
*/
