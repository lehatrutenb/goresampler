package utils

import "math"

// audio_util.h
// The conversion functions use the following naming convention:
// S16:      int16_t [-32768, 32767]
// Float:    float   [-1.0, 1.0]
// FloatS16: float   [-32768.0, 32768.0]
// Dbfs: float [-20.0*log(10, 32768), 0] = [-90.3, 0]

//math.Copysign(3.2, -1)

// convert to PCM format

// if calc by math so / 32767=(2^15 - 1) is correct, but diff is extr small and it makes calcs faster
// same in webrtc
func S16ToFloat(x int16) float32 {
	return float32(x) / float32(1<<15)
}

func S16ToFloat64(x int16) float64 {
	return float64(x) / float64(1<<15)
}

func FloatS16ToS16(x float32) int16 {
	return int16(math.Max(math.Min(float64(x), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func FFloatS16ToS16(x float64) int16 {
	return int16(math.Max(math.Min(float64(x), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

// same mark as for S16ToFloat
func FloatToS16(x float32) int16 {
	return int16(math.Max(math.Min(float64(x*32768.0), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func Float64ToS16(x float64) int16 {
	return int16(math.Max(math.Min(float64(x*32768.0), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func AFloatToS16(xs []float32) []int16 {
	res := make([]int16, len(xs))
	for i, x := range xs {
		res[i] = FloatToS16(x)
	}
	return res
}

func AFloat64ToS16(xs []float64) []int16 {
	res := make([]int16, len(xs))
	for i, x := range xs {
		res[i] = Float64ToS16(x)
	}
	return res
}

func AS16ToFloat(xs []int16) []float32 {
	res := make([]float32, len(xs))
	for i, x := range xs {
		res[i] = S16ToFloat(x)
	}
	return res
}

func AS16ToFloat64(xs []int16) []float64 {
	res := make([]float64, len(xs))
	for i, x := range xs {
		res[i] = S16ToFloat64(x)
	}
	return res
}

func AS16ToInt(xs []int16) []int {
	res := make([]int, len(xs))
	for i, x := range xs {
		res[i] = int(x)
	}
	return res
}

func FFloatToS16(x float64) int16 {
	return int16(math.Max(math.Min(float64(x*32768.0), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func FloatToFloatS16(x float32) float32 {
	return float32(math.Max(math.Min(float64(x), 1.0), -1.0) * 32768.0)
}

func FloatS16ToFloat(x float32) float32 {
	return float32(math.Max(math.Min(float64(x), 32768.0), -32768.0) / 32768.0)
}
