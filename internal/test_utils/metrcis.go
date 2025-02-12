package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"resampler/internal/utils"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const SAVE_PATH = "../../../test"
const REPORTS_SUFFIX = "reports"
const AUDIO_SUFFIX = "audio"
const LARGE_FILES_SUFFIX = "reports_large"
const DRAW_USING_GO = false

func createSoundFile(fName string, buf *audio.IntBuffer) error {
	f, err := os.Create(fName)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := wav.NewEncoder(f, buf.Format.SampleRate, buf.SourceBitDepth, buf.Format.NumChannels, 1) // 16 - bitDepth, 1 - numChannels, 1 - not compressed
	if err = enc.Write(buf); err != nil {
		return err
	}
	return enc.Close()
}

func SaveSoundFile(fName string, numCh, sR int, data []int16) error {
	buf := audio.IntBuffer{Format: &audio.Format{NumChannels: numCh, SampleRate: sR}, Data: utils.AS16ToInt(data), SourceBitDepth: 16}
	if err := createSoundFile(fName, &buf); err != nil {
		return err
	}

	return nil
}

func (tObj TestObj) saveSoundData(baseFName string) error {
	if err := SaveSoundFile(baseFName+"_inWave.wav", tObj.Tw.NumChannels(), tObj.Tw.InRate(), tObj.Tres.InWave); err != nil {
		return err
	}

	if err := SaveSoundFile(baseFName+"_OutWave.wav", tObj.Tw.NumChannels(), tObj.Tw.OutRate(), tObj.Tres.Resampeled); err != nil {
		return err
	}

	if tObj.Tw.WithResampled() {
		if err := SaveSoundFile(baseFName+"_CorrectOutWave.wav", tObj.Tw.NumChannels(), tObj.Tw.OutRate(), tObj.Tres.CorrectW); err != nil {
			return err
		}
	}

	return nil
}

func (tObj TestObj) Save(dirName string) error {
	tMarsh := TestResultZipped{Te: tObj.Tres.Te, SDur: tObj.Tres.SDur}
	bufL, err := json.Marshal(tObj.Tres)
	if err != nil {
		tObj.t.Error("failed to marshall results")
		return err
	}

	bufS, err := json.Marshal(tMarsh)
	if err != nil {
		tObj.t.Error("failed to marshall results")
		return err
	}

	reportsFName := fmt.Sprintf("%s/%s/%s/%s:%s", tObj.opts.OutPlotPath, REPORTS_SUFFIX, dirName, tObj.Tr, tObj.Tw)
	err = os.WriteFile(fmt.Sprintf("%s:small", reportsFName), bufS, 0666)
	if err != nil {
		tObj.t.Error("failed to save metrics file")
		return err
	}

	reportsLargeFName := fmt.Sprintf("%s/%s/%s/%s:%s", tObj.opts.OutPlotPath, LARGE_FILES_SUFFIX, dirName, tObj.Tr, tObj.Tw)
	err = os.WriteFile(fmt.Sprintf("%s:large", reportsLargeFName), bufL, 0666) // Care: large is keyword
	if err != nil {
		tObj.t.Error("failed to save metrics file")
		return err
	}

	if !tObj.opts.ToCrSF {
		return nil
	}

	audioFName := fmt.Sprintf("%s/%s/%s/%s:%s", tObj.opts.OutPlotPath, AUDIO_SUFFIX, dirName, tObj.Tr, tObj.Tw)
	if err := tObj.saveSoundData(audioFName); err != nil {
		tObj.t.Error("failed to save wav files")
		return err
	}
	return nil
}
