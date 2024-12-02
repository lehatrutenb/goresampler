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

func FloatS16ToS16(x float32) int16 {
	return int16(math.Max(math.Min(float64(x), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func FFloatS16ToS16(x float64) int16 {
	return int16(math.Max(math.Min(float64(x), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

// same mark as for S16ToFloat
func Float16ToS16(x float32) int16 {
	return int16(math.Max(math.Min(float64(x*32768.0), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func FFloat16ToS16(x float64) int16 {
	return int16(math.Max(math.Min(float64(x*32768.0), 32767.0), -32768.0) + math.Copysign(0.5, float64(x)))
}

func Float16ToFloatS16(x float32) float32 {
	return float32(math.Max(math.Min(float64(x), 1.0), -1.0) * 32768.0)
}

func FloatS16ToFloat(x float32) float32 {
	return float32(math.Max(math.Min(float64(x), 32768.0), -32768.0) / 32768.0)
}
