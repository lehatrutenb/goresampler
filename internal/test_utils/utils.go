package testutils

import (
	"errors"
	"fmt"
	"sync"

	"github.com/lehatrutenb/goresampler"
	"github.com/lehatrutenb/goresampler/internal/utils"
)

var (
	ErrUnimplemented    = errors.New("currently unimplemented resampling")
	ErrNotExpResampling = errors.New("called not expected resampling cfg")
)

func CalcMinOutSamplesPerInAmt(inAmt int64, rsm goresampler.Resampler) int64 {
	l, r := int64(0), inAmt*2+int64(1e7) // cause max multiplier in rsm 8->16 // add 1e7 to go over restrictions
	for l+1 < r {
		mid := (l + r) / 2
		if rsm.CalcNeedSamplesPerOutAmt(mid) < inAmt {
			l = mid
		} else {
			r = mid
		}
	}
	return utils.GetSecondReturnedVal(rsm.CalcInOutSamplesPerOutAmt(r))
}

func CalcMinOutSamplesPerInAmt2Waves(inAmt int64, rsm goresampler.Resampler2Waves) (int64, int64) {
	l, r := int64(0), inAmt*2+int64(1e7) // cause max multiplier in rsm 8->16 // add 1e7 to go over restrictions
	for l+1 < r {
		mid := (l + r) / 2
		if rsm.CalcNeedSamplesPerOutAmt(mid, mid*2) < inAmt {
			l = mid
		} else {
			r = mid
		}
	}
	_, out1, out2 := rsm.CalcInOutSamplesPerOutAmt(r, r*2)
	return out1, out2
}

func calcMinInSamplesAmt(inAmt int64, rsm goresampler.Resampler) int64 {
	return rsm.CalcNeedSamplesPerOutAmt(CalcMinOutSamplesPerInAmt(inAmt, rsm))
}

func GetWaveName[T goresampler.ResamplerTI](rsmT T, inRate, outRate int) string {
	return fmt.Sprintf("%d:%d:%d", rsmT, inRate, outRate)
}

func loadRealWave(samplesAmt int, rsmT goresampler.ResamplerT, waveInd, inRate, outRate int, res map[string]CutWave, mtx *sync.Mutex, gr *sync.WaitGroup, path string) {
	defer gr.Done()

	wave := CutWave{}.New(RealWave{}.New(waveInd, inRate, &outRate, &path), 0, samplesAmt).(CutWave)

	mtx.Lock()
	defer mtx.Unlock()
	res[GetWaveName(rsmT, inRate, outRate)] = wave
}

/*
func to conc load waves (as it is slow)
samplesAmt - min samples amt in wave will have
samplesDurS - min duration in wave will have
notStrictAmt - exect duration in wave will have (without any resamplers rools)
*/
func LoadAllRealWaves(waveInd int, pathToBaseWaves *string, samplesAmt *int, samplesDurS *int, notStrictAmt *int) map[string]CutWave { // rsmT:inRate:outRate
	mtx := &sync.Mutex{}
	res := make(map[string]CutWave)
	if pathToBaseWaves == nil {
		pathToBaseWaves = new(string)
		*pathToBaseWaves = PATH_TO_BASE_WAVES
	}

	gr := &sync.WaitGroup{}
	for _, outRate := range []int{8000, 16000} {
		for _, inRate := range []int{8000, 11000, 11025, 16000, 44000, 44100, 48000} {
			for _, rsmT := range []goresampler.ResamplerT{goresampler.ResamplerConstExprT, goresampler.ResamplerSplineT, goresampler.ResamplerFFtT, goresampler.ResamplerSincT} {
				if notStrictAmt != nil && rsmT != goresampler.ResamplerConstExprT && CheckRsmCompAb(goresampler.ResamplerConstExprT, inRate, outRate) == nil { // if not strict set - wave not depends on rsm type
					continue
				}
				if CheckRsmCompAb(rsmT, inRate, outRate) != nil {
					continue
				}
				gr.Add(1)
				rsm, _, err := goresampler.NewResamplerAuto[goresampler.BaseResamplerOptions](inRate, outRate, rsmT, nil)
				if err != nil {
					panic(err)
				}
				if notStrictAmt != nil {
					go loadRealWave(*notStrictAmt, rsmT, waveInd, inRate, outRate, res, mtx, gr, *pathToBaseWaves)
				} else if samplesAmt != nil {
					go loadRealWave(int(calcMinInSamplesAmt(int64(*samplesAmt), rsm)), rsmT, waveInd, inRate, outRate, res, mtx, gr, *pathToBaseWaves) // to resample > x frames
				} else if samplesDurS != nil {
					go loadRealWave(int(calcMinInSamplesAmt(int64(*samplesDurS)*int64(inRate), rsm)), rsmT, waveInd, inRate, outRate, res, mtx, gr, *pathToBaseWaves) // to resample > x frames
				} else {
					panic("expected one of samplesAmt or samplesDurS or notStrictAmt not nil")
				}
			}
		}
	}
	gr.Wait()
	return res
}

func CheckRsmCompAb[T goresampler.ResamplerTI](rsmInd T, inRate, outRate int) error {
	if rsmInd.String() == goresampler.ResamplerFFtT.String() && outRate > inRate {
		return ErrUnimplemented
	}
	if rsmInd.String() == goresampler.ResamplerConstExprT.String() && (inRate == 11025 || inRate == 44100) {
		return ErrNotExpResampling
	}
	return nil
}
