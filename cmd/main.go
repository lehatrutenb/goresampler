package main

import (
	"fmt"
	resamplerce "resampler/internal/resample/resamplerce"
	testutils "resampler/internal/test_utils"
	"resampler/internal/utils"
	"sync"
)

const README_TABLE_MIN_TIME_S = 30 // in secs
const PATH_TO_BASE_WAVES = "./base_waves/"
const OUTPUT_PATH = "./test/readme_audio"

func loadRealWave(rsmT, inRate, outRate int, res map[string]testutils.CutWave, mtx *sync.Mutex, gr *sync.WaitGroup) {
	defer gr.Done()

	path := PATH_TO_BASE_WAVES
	var need int
	switch rsmT {
	case 0:
		rsm, err := resamplerce.NewAutoResampler(inRate, outRate)
		if err != nil { // got unexp rate for const resamplers
			return
		}
		need = rsm.CalcNeedSamplesPerOutAmt(README_TABLE_MIN_TIME_S * outRate)
	}
	wave := testutils.CutWave{}.New(testutils.RealWave{}.New(0, inRate, &outRate, &path), 0, need).(testutils.CutWave)

	mtx.Lock()
	defer mtx.Unlock()
	res[fmt.Sprintf("%d:%d", inRate, outRate)] = wave
}

func loadAllRealWaves() (map[string]testutils.CutWave, error) { // inRate:outRate
	mtx := &sync.Mutex{}
	res := make(map[string]testutils.CutWave)

	gr := &sync.WaitGroup{}
	for _, outRate := range []int{8000, 16000} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			gr.Add(1)
			go loadRealWave(0, inRate, outRate, res, mtx, gr)
		}
	}
	gr.Wait()
	return res, nil
}

func createReadmeAudioTable() error {
	waves, err := loadAllRealWaves()
	if err != nil {
		return err
	}
	for i, outRate := range []int{8000, 16000} {
		for j, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for rsmInd := 0; rsmInd < 3; rsmInd++ {
				if rsmInd > 0 {
					continue
				}
				if inRate == outRate {
					continue
				}
				if rsmInd == 0 && inRate == 11025 || inRate == 44100 {
					continue
				}
				if rsmInd != 0 && inRate == 11000 || inRate == 44000 {
					continue
				}

				rsm, err := resamplerce.NewAutoResampler(inRate, outRate)
				if err != nil {
					return err
				}
				wave := waves[fmt.Sprintf("%d:%d", inRate, outRate)]

				in, err := wave.GetFullInWave()
				if err != nil {
					return err
				}

				in1 := utils.GetWithStep(in, 0, 2)
				in2 := utils.GetWithStep(in, 1, 2)

				out1 := make([]int16, rsm.CalcOutSamplesPerInAmt(len(in1)))
				out2 := make([]int16, rsm.CalcOutSamplesPerInAmt(len(in2)))
				if err = rsm.Resample(in1, out1); err != nil {
					return err
				}
				if err = rsm.Resample(in2, out2); err != nil {
					return err
				}

				out := utils.Merge2Channels(out1, out2)

				err = testutils.SaveSoundFile(fmt.Sprintf("%s/%d%d%d.wav", OUTPUT_PATH, 0, i, j), wave.NumChannels(), outRate, out)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	err := createReadmeAudioTable()
	if err != nil {
		panic(err)
	}
}
