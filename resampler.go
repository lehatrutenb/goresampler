// Package goresampler provides resampling structs (resampler_*) for changing sample rate
// of given sound waves
package goresampler

const BaseTimeErrRate float64 = 1e-6
const MaxResamplingBatchLen = int(1e7) // 1e7 cause resmapling for 2e9/1e7 multiplying sounds useless

type BaseResamplerOptions struct {
	MaxErrRateP *float64
}

func (opts *BaseResamplerOptions) Init() {
	if opts.MaxErrRateP == nil {
		baseTimeErrRate := BaseTimeErrRate
		opts.MaxErrRateP = &baseTimeErrRate
	}
}
