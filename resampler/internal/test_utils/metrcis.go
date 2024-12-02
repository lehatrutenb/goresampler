package testutils

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"resampler/internal/utils"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const SAVE_PATH = "../../../plots"

func createPointsFromData(wave []int16) plotter.XYs {
	pts := make(plotter.XYs, len(wave))
	for i, p := range wave {
		pts[i] = plotter.XY{X: float64(i), Y: float64(utils.S16ToFloat(p))}
	}
	return pts
}

func (tObj TestObj) drawPlots(baseFName string) error {
	// TODO save more data

	p := plot.New()

	p.Title.Text = baseFName
	p.X.Label.Text = "X"
	p.Y.Label.Text = "PCM"

	sCorr, err := plotter.NewScatter(createPointsFromData(tObj.Tres.CorrectW))
	if err != nil {
		log.Println("failed to create scatter")
		return err
	}
	sCorr.GlyphStyle.Radius = vg.Points(1)
	sCorr.GlyphStyle.Color = color.RGBA{R: 255}
	p.Add(sCorr)

	sRes, err := plotter.NewScatter(createPointsFromData(tObj.Tres.Resampeled))
	if err != nil {
		log.Println("failed to create scatter")
		return err
	}
	sRes.GlyphStyle.Radius = vg.Points(1)
	sRes.GlyphStyle.Color = color.RGBA{B: 255}
	p.Add(sCorr, sRes)

	err = plotutil.AddScatters(p, "Correct", createPointsFromData(tObj.Tres.CorrectW), "Resampled", createPointsFromData(tObj.Tres.Resampeled))
	if err != nil {
		panic(err)
	}

	if err := p.Save(30*vg.Inch, 12*vg.Inch, fmt.Sprintf("%s:plot.png", baseFName)); err != nil {
		panic(err)
	}
	return nil
}

func (tObj TestObj) fastDrawPlots(baseFName string) error {
	// TODO save more data

	p := plot.New()

	p.Title.Text = baseFName
	p.X.Label.Text = "X"
	p.Y.Label.Text = "PCM"

	err := plotutil.AddScatters(p, "Correct", createPointsFromData(tObj.Tres.CorrectW), "Resampled", createPointsFromData(tObj.Tres.Resampeled))
	if err != nil {
		panic(err)
	}

	if err := p.Save(30*vg.Inch, 12*vg.Inch, fmt.Sprintf("%s:plot.png", baseFName)); err != nil {
		panic(err)
	}
	return nil
}

func (tObj TestObj) Save(dirName string) error {
	tMarsh := TestResultZipped{Te: tObj.Tres.Te, SDur: tObj.Tres.SDur}
	bufL, err := json.Marshal(tObj.Tres)
	if err != nil {
		log.Println("failed to marshall results")
		return err
	}

	bufS, err := json.Marshal(tMarsh)
	if err != nil {
		log.Println("failed to marshall results")
		return err
	}

	baseFName := fmt.Sprintf("%s/%s/%s:%s:small", SAVE_PATH, dirName, tObj.Tr, tObj.Tw)
	err = os.WriteFile(fmt.Sprintf("%s:small", baseFName), bufS, 0666)
	if err != nil {
		log.Println("failed to save metrices file")
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s:large", baseFName), bufL, 0666)
	if err != nil {
		log.Println("failed to save metrices file")
		return err
	}

	tObj.fastDrawPlots(baseFName)
	return nil
}
