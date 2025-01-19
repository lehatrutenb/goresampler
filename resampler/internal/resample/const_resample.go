package resample

import "resampler/internal/resample/resamplerl"

func Resample11To8(in []int16, out *[]int16) error {
	window := 220
	var sum int32 = 0
	for i := 0; i < window && len(in)-i >= 0; i++ {
		sum += int32(in[len(in)-i-1])
	}
	for len(in)%220 != 0 {
		in = append(in, int16(sum/int32(window)))
		if len(in) < 220 {
			continue
		}
		sum -= int32(in[len(in)-window])
		sum += int32(in[len(in)-1])
	}

	resamplerl.Resample11To8L(in, out)

	return nil
}
