package benchmark

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	testutils "github.com/lehatrutenb/goresampler/internal/test_utils"
	"github.com/lehatrutenb/goresampler/internal/utils"
	"github.com/lehatrutenb/goresampler/resamplerauto"
	"github.com/lehatrutenb/goresampler/resampleri"
)

var MIN_RESAMPLE_DURATION_S int // min duration of input wave in secs
var MIN_SAMPLES_AMT int         // min amt of samples in input wave
// var RERUN_AMT = 100             // amt of times to resample each wave

var PATH_TO_BASE_WAVES = "../../base_waves/"

const OUTPUT_PATH = "../../test/readme_audio/"

var RealWaves map[string]testutils.CutWave

type Benchmarker struct {
	b       *testing.B
	rsmT    resamplerauto.ResamplerT
	inRate  int
	outRate int
	rsm     resampleri.Resampler
	wave    testutils.CutWave
	in      []int16
}

func (bmr Benchmarker) New(rsmT resamplerauto.ResamplerT, inRate, outRate int, b *testing.B) Benchmarker {
	if err := testutils.CheckRsmCompAb(rsmT, inRate, outRate); err != nil {
		b.Error(err)
		b.FailNow()
	}

	rsm, _, err := resamplerauto.New(inRate, outRate, rsmT, nil)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	wave := RealWaves[fmt.Sprintf("%d:%d:%d", rsmT, inRate, outRate)]
	in, err := wave.GetFullInWave()
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	return Benchmarker{b, rsmT, inRate, outRate, rsm, wave, in}
}

func (bmr Benchmarker) setup() ([]int16, []int16, []int16, []int16) {
	in1 := utils.GetWithStep(bmr.in, 0, 2)
	in2 := utils.GetWithStep(bmr.in, 1, 2)
	out1 := make([]int16, testutils.CalcMinOutSamplesPerInAmt(len(in1), bmr.rsm))
	out2 := make([]int16, testutils.CalcMinOutSamplesPerInAmt(len(in2), bmr.rsm))
	return in1, in2, out1, out2
}

func (bmr Benchmarker) resample(in1, in2, out1, out2 []int16) []int16 {
	var err1, err2 error
	bmr.b.ResetTimer()

	for i := 0; i < bmr.b.N; i++ {
		err1 = bmr.rsm.Resample(in1, out1) // CARE COUNT ONLY 1 CHANNEL RESAMPLING
	}

	bmr.b.StopTimer()

	err2 = bmr.rsm.Resample(in2, out2)

	bmr.chkErr(err1)
	bmr.chkErr(err2)
	return utils.Merge2Channels(out1, out2)
}

func (bmr Benchmarker) tearDown(out []int16) {
	pref := (int(bmr.rsmT)-1)*9 + inRateConv[bmr.inRate] + outRateConv[bmr.outRate]*5 // just try to index all audios
	pref -= pref / 8

	err := testutils.SaveSoundFile(fmt.Sprintf("%s/%d_%s_%dto%d.mp4", OUTPUT_PATH, pref, bmr.rsmT, bmr.inRate, bmr.outRate), bmr.wave.NumChannels(), bmr.outRate, out) // mp4 not breakes anything - just to load to git
	bmr.chkErr(err)
	err = testutils.SaveSoundFile(fmt.Sprintf("%s/listenable/%d_%s_%dto%d.wav", OUTPUT_PATH, pref, bmr.rsmT, bmr.inRate, bmr.outRate), bmr.wave.NumChannels(), bmr.outRate, out) // to have an ability to just listen
	bmr.chkErr(err)
}

func (bmr Benchmarker) chkErr(err error) {
	if err != nil {
		bmr.b.Error(err)
		bmr.b.FailNow()
	}
}

func benchResampler(rsmT resamplerauto.ResamplerT, inRate, outRate int, b *testing.B) {
	bmr := Benchmarker{}.New(rsmT, inRate, outRate, b)
	bmr.tearDown(bmr.resample(bmr.setup()))
}

func BenchmarkConstExpr11000_8000(b *testing.B) { benchResampler(1, 11000, 8000, b) }
func BenchmarkConstExpr16000_8000(b *testing.B) { benchResampler(1, 16000, 8000, b) }
func BenchmarkConstExpr44000_8000(b *testing.B) { benchResampler(1, 44000, 8000, b) }
func BenchmarkConstExpr48000_8000(b *testing.B) { benchResampler(1, 48000, 8000, b) }

func BenchmarkConstExpr8000_16000(b *testing.B)  { benchResampler(1, 8000, 16000, b) }
func BenchmarkConstExpr11000_16000(b *testing.B) { benchResampler(1, 11000, 16000, b) }
func BenchmarkConstExpr44000_16000(b *testing.B) { benchResampler(1, 44000, 16000, b) }
func BenchmarkConstExpr48000_16000(b *testing.B) { benchResampler(1, 48000, 16000, b) }

func BenchmarkSpline11025_8000(b *testing.B) { benchResampler(2, 11025, 8000, b) }
func BenchmarkSpline16000_8000(b *testing.B) { benchResampler(2, 16000, 8000, b) }
func BenchmarkSpline44100_8000(b *testing.B) { benchResampler(2, 44100, 8000, b) }
func BenchmarkSpline48000_8000(b *testing.B) { benchResampler(2, 48000, 8000, b) }

func BenchmarkSpline8000_16000(b *testing.B)  { benchResampler(2, 8000, 16000, b) }
func BenchmarkSpline11025_16000(b *testing.B) { benchResampler(2, 11025, 16000, b) }
func BenchmarkSpline44100_16000(b *testing.B) { benchResampler(2, 44100, 16000, b) }
func BenchmarkSpline48000_16000(b *testing.B) { benchResampler(2, 48000, 16000, b) }

func BenchmarkFFT11025_8000(b *testing.B) { benchResampler(3, 11025, 8000, b) }
func BenchmarkFFT16000_8000(b *testing.B) { benchResampler(3, 16000, 8000, b) }
func BenchmarkFFT44100_8000(b *testing.B) { benchResampler(3, 44100, 8000, b) }
func BenchmarkFFT48000_8000(b *testing.B) { benchResampler(3, 48000, 8000, b) }

// func BenchmarkFFT8000_16000(b *testing.B)  { benchResampler(2, 8000, 16000, b) }
// func BenchmarkFFT11025_16000(b *testing.B) { benchResampler(2, 11025, 16000, b) }
func BenchmarkFFT44100_16000(b *testing.B) { benchResampler(3, 44100, 16000, b) }
func BenchmarkFFT48000_16000(b *testing.B) { benchResampler(3, 48000, 16000, b) }

func TestMain(t *testing.M) {
	MIN_SAMPLES_AMT = 0
	MIN_RESAMPLE_DURATION_S = 0
	waveInd := 0
	var err error

	for _, arg := range os.Args {
		if ind := strings.Index(arg, "="); ind != -1 {
			if strings.HasPrefix(arg, "minsamplesamt") {
				MIN_SAMPLES_AMT, err = strconv.Atoi(arg[ind+1:])
				if err != nil {
					log.Fatal("failed to parse minsamplesamt")
				}
			}
			if strings.HasPrefix(arg, "minsampledurationins") {
				MIN_RESAMPLE_DURATION_S, err = strconv.Atoi(arg[ind+1:])
				if err != nil {
					log.Fatal("failed to parse minsampledurationins")
				}
			}
		}
		if arg == "customwave" {
			waveInd = 3
		}
	}
	if MIN_SAMPLES_AMT == 0 {
		log.Println("failed to get min samples amt for benchmark - use 60 sec")
		MIN_RESAMPLE_DURATION_S = 60
	}

	if MIN_SAMPLES_AMT == 0 {
		RealWaves = testutils.LoadAllRealWaves(waveInd, &PATH_TO_BASE_WAVES, nil, &MIN_RESAMPLE_DURATION_S, nil)
	} else {
		RealWaves = testutils.LoadAllRealWaves(waveInd, &PATH_TO_BASE_WAVES, &MIN_SAMPLES_AMT, nil, nil)
	}
	t.Run()
}
