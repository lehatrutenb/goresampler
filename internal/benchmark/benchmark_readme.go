package benchmark

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/lehatrutenb/go_resampler/internal/resample/resamplerauto"
	testutils "github.com/lehatrutenb/go_resampler/internal/test_utils"

	"github.com/nao1215/markdown"
)

const PATH_TO_AUDO_URLS = "./internal/benchmark/audio_urls"

func CreateReadmeAudioTable() error {
	data, err := os.ReadFile(PATH_TO_AUDO_URLS)
	if err != nil {
		return err
	}

	// create order same as audio urls appear in file
	order := make([][3]int, 0, 3*5*2)
	for _, rsmT := range []resamplerauto.ResamplerT{resamplerauto.ResamplerConstExpr, resamplerauto.ResamplerSpline, resamplerauto.ResamplerFFT} {
		for _, outRate := range []int{8000, 16000} {
			for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
				if testutils.CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				order = append(order, [3]int{int(rsmT) - 1, inRateConv[inRate], outRateConv[outRate]})
			}
		}
	}

	urls := make([][]string, 4*2)
	for i := range urls {
		urls[i] = make([]string, 5)
	}

	dataS := strings.ReplaceAll(string(data), "\n\n", "\n")
	for i := 0; i < 20; i++ {
		dataS = strings.ReplaceAll(dataS, "\n\n", "\n")
	}
	splittedData := strings.Split(dataS, "\n")
	slices.Reverse(splittedData)
	ffmpeg8000 := splittedData[len(splittedData)-2]
	ffmpeg16000 := splittedData[len(splittedData)-1]
	splittedData = splittedData[:len(splittedData)-2]

	elPref := "src="
	elSuff := ">"
	// set urls to specific places in table
	for ind, url := range splittedData {
		i, j, k := order[ind][0], order[ind][1], order[ind][2]

		row := j + k*5 - 1
		if row >= 7 {
			row--
		}
		urls[row][i+1] = elPref + url + elSuff

		if k == 0 {
			urls[row][4] = elPref + ffmpeg8000 + elSuff
		} else {
			urls[row][4] = elPref + ffmpeg16000 + elSuff
		}

		urls[row][0] = fmt.Sprintf("%d to %d", rInRateConv[j][1], rOutRateConv[k])
	}

	buf := make([]byte, 1e6)
	md := bytes.NewBuffer(buf)
	err = markdown.NewMarkdown(md).H2("Resample results").
		Table(markdown.TableSet{
			Header: []string{"/", resamplerauto.ResamplerT(1).String(), resamplerauto.ResamplerT(2).String(), resamplerauto.ResamplerT(3).String(), "FFMPEG resampling"},
			Rows:   urls,
		}).Build()

	if err != nil {
		return err
	}

	table, err := md.ReadString('~') // base lib not support such long lines, but github does - so some hand work
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	table = strings.ReplaceAll(table, "src=", "<video src=")
	table = strings.ReplaceAll(table, ">", "> </video>")
	fmt.Println(table)

	return nil
}
