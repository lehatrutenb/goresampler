package goresampler

import (
	"errors"
)

var (
	// ErrUnexpResRate indicates that that ResamplerT not support that conversion
	ErrUnexpResRate = errors.New("got unexpected in rate or out rate to resample")
	// ErrUnexpResamplerType indicates that auto resampler got unsupported rsmT
	ErrUnexpResamplerType   = errors.New("got unexpected resampler type (not const of ResamplerT)")
	ErrUnreadyResamplerType = errors.New("try to get currently not done resampler type")
)

// ResamplerT describes resampler to use in ResamplerAuto
type ResamplerT int

const ResamplerConstExprT ResamplerT = 1
const ResamplerSplineT ResamplerT = 2
const ResamplerFFtT ResamplerT = 3
const ResamplerBestFitT ResamplerT = 4

func (rsmT ResamplerT) String() string {
	switch rsmT {
	case ResamplerConstExprT:
		return "Const_expression_resampler"
	case ResamplerSplineT:
		return "Spline_resampler"
	case ResamplerFFtT:
		return "FFT_resampler"
	case ResamplerBestFitT:
		return "BestFit_resampler"
	default:
		return "Undefined"
	}
}

// Resampler2WavesT describes resamplers that resample to 2 waves simult
type Resampler2WavesT int

const Resampler2WavesSplineT Resampler2WavesT = 10

func (rsmT Resampler2WavesT) GetRsmIns() (ResamplerT, error) {
	switch rsmT {
	case Resampler2WavesSplineT:
		return ResamplerSplineT, nil
	default:
		return ResamplerSplineT, ErrUnreadyResamplerType
	}
}

func (rsmT Resampler2WavesT) String() string {
	switch rsmT {
	case Resampler2WavesSplineT:
		return "Spline_resampler_2waves"
	default:
		return "Undefined"
	}
}

type ResamplerTI interface {
	ResamplerT | Resampler2WavesT
	String() string
}

// resampler that wraps other resamplers and give ability to choose which of them to use
//
// has new type of resampler: ResamplerBestFitT - use Const expression resampler if
// can convert with such inRate, outRate, otherwise use Spline rasmpler
type ResamplerAuto struct {
	inRate  int
	outRate int
	Resampler
}

// returns
// error:
//
// if given rsmT not implement such rate convertion - ErrUnexpResRate
// if given rsmT not fit in known rsm types - ErrUnexpResamplerType
//
// bool:
//
// try to find batch input amt to have less err (0..1) rate than given maxErrRateP
//
// if failed to find such batch to fit maxErrRate,  second arg is false,
// otherwise true (but even with false, resampler is fine to use)
func NewResamplerAuto(inRate, outRate int, rsmT ResamplerT, maxErrRateP *float64) (ResamplerAuto, bool, error) {
	var rsm Resampler = nil
	var ok bool = true

	if inRate == outRate {
		return ResamplerAuto{inRate, outRate, NewRsmNotChange()}, true, nil
	}

	var rsm Resampler = nil
	var ok bool = true
	switch rsmT {
	case ResamplerSplineT:
		rsm, ok = NewResamplerSpline(inRate, outRate, maxErrRateP)
	case ResamplerFFtT:
		if inRate <= outRate {
			return ResamplerAuto{}, false, ErrUnexpResRate
		}
		rsm, ok = NewResamplerFFT(inRate, outRate, maxErrRateP)
	case ResamplerBestFitT:
		switch inRate {
		case 11025, 44100:
			rsm, ok = NewResamplerSpline(inRate, outRate, maxErrRateP)
		default:
			rsmT = ResamplerConstExprT
		}
	}
	if rsmT == ResamplerConstExprT {
		switch outRate {
		case 8000:
			switch inRate {
			case 11000: // not 11025!
				rsm = NewRsm11To8L()
			case 16000:
				rsm = NewRsm16To8L()
			case 44000: // not 44100!
				rsm = NewRsm44To8L()
			case 48000:
				rsm = NewRsm48To8L()
			default:
				return ResamplerAuto{}, false, ErrUnexpResRate
			}
		case 16000:
			switch inRate {
			case 8000:
				rsm = NewRsm8To16L()
			case 11000: // not 11025!
				rsm = NewRsm11To16L()
			case 44000: // not 44100!
				rsm = NewRsm44To16L()
			case 48000:
				rsm = NewRsm48To16L()
			default:
				return ResamplerAuto{}, false, ErrUnexpResRate
			}
		default:
			return ResamplerAuto{}, false, ErrUnexpResRate
		}
	}

	if rsm == nil {
		return ResamplerAuto{}, false, ErrUnexpResamplerType
	}

	return ResamplerAuto{inRate, outRate, rsm}, ok, nil
}

// resampler that wraps other resamplers for 2 waves and give ability to choose which of them to use
type ResamplerAuto2Waves struct {
	inRate   int
	outRate1 int
	outRate2 int
	Resampler2Waves
}

// returns
//
// error:
//
// if given rsmT not fit in known rsm types - ErrUnexpResamplerType
//
// bool:
//
// try to find batch input amt to have less err (0..1) rate than given maxErrRateP
//
// if failed to find such batch to fit maxErrRate,  second arg is false,
// otherwise true (but even with false, resampler is fine to use)
func NewResamplerAuto2Waves(inRate, outRate1, outRate2 int, rsmT Resampler2WavesT, maxErrRateP *float64) (ResamplerAuto2Waves, bool, error) {
	var rsm Resampler2Waves = nil
	var ok bool = true

	switch rsmT {
	case Resampler2WavesSplineT:
		rsm, ok = NewResamplerSpline2Waves(inRate, outRate1, outRate2, maxErrRateP)
	}

	if rsm == nil {
		return ResamplerAuto2Waves{}, false, ErrUnexpResamplerType
	}

	return ResamplerAuto2Waves{inRate, outRate1, outRate2, rsm}, ok, nil
}
