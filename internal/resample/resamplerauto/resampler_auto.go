package resamplerauto

import (
	"errors"
	"resampler/internal/resample/resamplerce"
	"resampler/internal/resample/resamplerfft"
	"resampler/internal/resample/resampleri"
	"resampler/internal/resample/resamplerspline"
)

var (
	ErrUnexpResRate       = errors.New("got unexpected in rate or out rate to resample")
	ErrUnexpResamplerType = errors.New("got unexpected resampler type (not in {1,2,3})")
)

type ResamplerT int

const ResamplerConstExpr ResamplerT = 1
const ResamplerSpline ResamplerT = 2
const ResamplerFFT ResamplerT = 3
const ResamplerBestFit ResamplerT = 4

func (rsmT ResamplerT) String() string {
	switch rsmT {
	case ResamplerConstExpr:
		return "Const_expression_resampler"
	case ResamplerSpline:
		return "Spline_resampler"
	case ResamplerFFT:
		return "FFT_resampler"
	case ResamplerBestFit:
		return "BestFit_resampler"
	default:
		return "Undefined"
	}
}

// not real auto, but just merged all types
type ResamplerAuto struct {
	inRate  int
	outRate int
	resampleri.Resampler
}

// yeah, it hurts when new return err , but don't want use another ways to create
/*
try to find batch input amt to have less err (0..1) rate than given maxErrRateP
if failed to find such batch to fit maxErrRate,  second arg is false, otherwise true (but even with false, resampler is fine to use)
*/
func New(inRate, outRate int, rsmT ResamplerT, maxErrRateP *float64) (resampleri.Resampler, bool, error) {
	if inRate == outRate {
		return nil, false, ErrUnexpResRate
	}

	switch rsmT {
	case ResamplerSpline:
		rsm, ok := resamplerspline.New(inRate, outRate, maxErrRateP)
		return rsm, ok, nil
	case ResamplerFFT:
		if inRate <= outRate {
			return nil, false, ErrUnexpResRate
		}
		rsm, ok := resamplerfft.New(inRate, outRate, maxErrRateP)
		return rsm, ok, nil
	case ResamplerBestFit:
		switch inRate {
		case 11025, 44100:
			rsm, ok := resamplerspline.New(inRate, outRate, maxErrRateP)
			return rsm, ok, nil
		default:
			rsmT = ResamplerConstExpr
		}
	}
	switch rsmT {
	case ResamplerConstExpr:
		var rsm resampleri.Resampler
		switch outRate {
		case 8000:
			switch inRate {
			case 11000: // not 11025!
				rsm = resamplerce.NewRsm11To8L()
			case 16000:
				rsm = resamplerce.NewRsm16To8L()
			case 44000: // not 44100!
				rsm = resamplerce.NewRsm44To8L()
			case 48000:
				rsm = resamplerce.NewRsm48To8L()
			default:
				return nil, false, ErrUnexpResRate
			}
		case 16000:
			switch inRate {
			case 8000:
				rsm = resamplerce.NewRsm8To16L()
			case 11000: // not 11025!
				rsm = resamplerce.NewRsm11To16L()
			case 44000: // not 44100!
				rsm = resamplerce.NewRsm44To16L()
			case 48000:
				rsm = resamplerce.NewRsm48To16L()
			default:
				return nil, false, ErrUnexpResRate
			}
		default:
			return nil, false, ErrUnexpResRate
		}
		return ResamplerAuto{inRate, outRate, rsm}, true, nil
	default:
		return nil, false, ErrUnexpResRate
	}
}

func (rsm ResamplerAuto) Reset() {
	rsm.Resampler.Reset()
}
