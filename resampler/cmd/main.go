package main

import (
	"errors"
	"log"
	"resampler/internal/resample"
	testutils "resampler/internal/test_utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

type resampler48To32 struct {
	resampled []int16
}

func (resampler48To32) String() string {
	return "48000_to_32000_resampler"
}

func (rsm *resampler48To32) Resample(inp []int16) error {
	rsm.resampled = resample.Resample48To32(inp)
	return nil
}

func (rsm resampler48To32) Get(ind int) (int16, error) {
	if ind >= len(rsm.resampled) {
		return 0, errors.New("out of bounds")
	}
	return rsm.resampled[ind], nil
}

func main() {
	var t *testing.T = &testing.T{}

	var tObj testutils.TestObj = testutils.TestObj{}.New(testutils.SinWave{}.Init(0, 5, 48000, 32000), testutils.TestResampler(&resampler48To32{}), 5)
	err := tObj.Run()
	if !assert.NoError(t, err, "failed to run resampler") {
		log.Println(err)
	}

	err = tObj.Save("latest")
	if !assert.NoError(t, err, "failed to save test results") {
		log.Println(err)
	}
}
