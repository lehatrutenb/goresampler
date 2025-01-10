package testutils

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"resampler/internal/utils"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const SAVE_PATH = "../../../plots"
const DRAW_USING_GO = false

func createPointsFromData(wave []int16) plotter.XYs {
	pts := make(plotter.XYs, len(wave))
	for i, p := range wave {
		pts[i] = plotter.XY{X: float64(i), Y: float64(utils.S16ToFloat(p))}
	}
	return pts
}

func (tObj TestObj) drawPlots(baseFName string) error {
	// TODO save more data

	maxDataLength := min(100000, len(tObj.Tres.Resampeled)-1) // to make it fast

	for i := 0; i < tObj.Tw.NumChannels(); i++ {
		p := plot.New()

		p.Title.Text = baseFName + string(i) + "Channel"
		p.X.Label.Text = "X"
		p.Y.Label.Text = "PCM"
		if tObj.Tw.WithResampled() {
			sCorr, err := plotter.NewScatter(createPointsFromData(utils.GetWithStep(tObj.Tres.CorrectW, i, tObj.Tw.NumChannels()))[:maxDataLength])
			if err != nil {
				tObj.t.Error("failed to create scatter")
				return err
			}
			sCorr.GlyphStyle.Radius = vg.Points(1)
			sCorr.GlyphStyle.Color = color.RGBA{R: 255}
			p.Add(sCorr)
		}

		sRes, err := plotter.NewScatter(createPointsFromData(utils.GetWithStep(tObj.Tres.Resampeled, i, tObj.Tw.NumChannels()))[:maxDataLength])
		if err != nil {
			tObj.t.Error("failed to create scatter")
			return err
		}
		sRes.GlyphStyle.Radius = vg.Points(1)
		sRes.GlyphStyle.Color = color.RGBA{B: 255}
		p.Add(sRes)

		if err := p.Save(30*vg.Inch, 12*vg.Inch, fmt.Sprintf("%s:plot%dch.png", baseFName, i)); err != nil {
			tObj.t.Error("failed to save plots")
			return err
		}
	}
	return nil
}

func (tObj TestObj) fastDrawPlots(baseFName string) error {
	// TODO save more data

	maxDataLength := min(100000, len(tObj.Tres.Resampeled)-1) // to make it fast
	p := plot.New()

	p.Title.Text = baseFName
	p.X.Label.Text = "X"
	p.Y.Label.Text = "PCM"

	err := plotutil.AddScatters(p, "Correct", createPointsFromData(tObj.Tres.CorrectW[:maxDataLength]), "Resampled", createPointsFromData(tObj.Tres.Resampeled[:maxDataLength]))
	if err != nil {
		tObj.t.Error("failed to add scatters in plots")
		return err
	}

	if err := p.Save(30*vg.Inch, 12*vg.Inch, fmt.Sprintf("%s:plot.png", baseFName)); err != nil {
		tObj.t.Error("failed to save plots")
		return err
	}
	return nil
}

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

func (tObj TestObj) saveSoundFile(baseFName string) error {
	buf1 := audio.IntBuffer{Format: &audio.Format{NumChannels: tObj.Tw.NumChannels(), SampleRate: tObj.Tw.InRate()}, Data: utils.AS16ToInt(tObj.Tres.InWave), SourceBitDepth: 16}
	buf2 := audio.IntBuffer{Format: &audio.Format{NumChannels: tObj.Tw.NumChannels(), SampleRate: tObj.Tr.OutRate()}, Data: utils.AS16ToInt(tObj.Tres.Resampeled), SourceBitDepth: 16}

	if err := createSoundFile(baseFName+"_inWave.wav", &buf1); err != nil {
		return err
	}

	if err := createSoundFile(baseFName+"_OutWave.wav", &buf2); err != nil {
		return err
	}

	if tObj.Tw.WithResampled() {
		buf3 := audio.IntBuffer{Format: &audio.Format{NumChannels: tObj.Tw.NumChannels(), SampleRate: tObj.Tw.OutRate()}, Data: utils.AS16ToInt(tObj.Tres.CorrectW), SourceBitDepth: 16}
		if err := createSoundFile(baseFName+"_CorrectOutWave.wav", &buf3); err != nil {
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

	baseFName := fmt.Sprintf("%s/%s/%s:%s", SAVE_PATH, dirName, tObj.Tr, tObj.Tw)
	err = os.WriteFile(fmt.Sprintf("%s:small", baseFName), bufS, 0666)
	if err != nil {
		tObj.t.Error("failed to save metrics file")
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s:large", baseFName), bufL, 0666) // Care: large is keyword
	if err != nil {
		tObj.t.Error("failed to save metrics file")
		return err
	}

	if DRAW_USING_GO {
		if err := tObj.drawPlots(baseFName); err != nil {
			tObj.t.Error("failed to plot using go")
			return err
		}
	}
	if err := tObj.saveSoundFile(baseFName); err != nil {
		tObj.t.Error("failed to save wav files")
		return err
	}
	return nil
}
