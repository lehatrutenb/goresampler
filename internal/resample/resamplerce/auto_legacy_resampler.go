package resamplerce

import (
	"errors"
	"resampler/internal/resample/resampleri"
)

var ErrUnexpResRate = errors.New("got unexpected in rate or out rate to resample")

type ResamplerL struct {
	inRate  int
	outRate int
	resampleri.Resampler
}

// yeah, it hurts when new return err , but don't want use another ways to create
func NewAutoResampler(inRate, outRate int) (resampleri.Resampler, error) {
	var rsm resampleri.Resampler
	switch outRate {
	case 8000:
		switch inRate {
		case 8000:
			return nil, ErrUnexpResRate
		case 11000: // not 11025!
			rsm = NewRsm11To8L()
		case 16000:
			rsm = NewRsm16To8L()
		case 44000: // not 44100!
			rsm = NewRsm44To8L()
		case 48000:
			rsm = NewRsm48To8L()
		default:
			return nil, ErrUnexpResRate
		}
	case 16000:
		switch inRate {
		case 8000:
			rsm = NewRsm8To16L()
		case 11000: // not 11025!
			rsm = NewRsm11To16L()
		case 16000:
			return nil, ErrUnexpResRate
		case 44000: // not 44100!
			rsm = NewRsm44To16L()
		case 48000:
			rsm = NewRsm48To16L()
		default:
			return nil, ErrUnexpResRate
		}
	default:
		return nil, ErrUnexpResRate
	}
	return ResamplerL{inRate, outRate, rsm}, nil
}
