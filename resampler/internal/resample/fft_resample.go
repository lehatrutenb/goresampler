package resample

import (
	"math"
	"resampler/internal/utils"
)

type FFTResampler struct {
	in      []float32
	out     []float32
	inRate  int
	outRate int
}

func (FFTResampler) New(in []int16, inRate int, outRate int) *FFTResampler {
	return &FFTResampler{in: utils.AS16ToFloat(in), out: make([]float32, len(in)*outRate/inRate), inRate: inRate, outRate: outRate}
}

func (rsm FFTResampler) GetOutWave() []int16 {
	return utils.AFloatToS16(rsm.out)
}

func cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

func sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func pi() float32 {
	return float32(math.Pi)
}

func forwardFT(in []float32) ([]float32, []float32) {
	re := make([]float32, len(in)/2+1)
	im := make([]float32, len(in)/2+1)
	for j := 0; j < len(re); j++ {
		for i, v := range in {
			re[j] += v * cos(2*pi()*float32(j*i)/float32(len(in)))
			im[j] += v * sin(2*pi()*float32(j*i)/float32(len(in)))
		}
		im[j] *= -1
	}
	return re, im
}

func normalize(re []float32, im []float32, ni int, n float32) { // ni == int(n)
	for i := 0; i < len(re); i++ {
		re[i] /= n
		re[i] *= 2
		im[i] /= n
		im[i] *= -2
	}
	re[0] /= 2
	re[ni/2] /= 2
}

func backwardFT(re []float32, im []float32, n float32) []float32 {
	out := make([]float32, (len(re)-1)*2)
	for i := 0; i < len(out); i++ {
		for j := 0; j < len(re); j++ {
			out[i] += re[j] * cos(2*pi()*float32(j*i)/n)
			out[i] += im[j] * sin(2*pi()*float32(j*i)/n)
		}
	}
	return out
}

func (rsm *FFTResampler) Resample() {
	/*p2 := 1
	for ; p2 < len(rsm.in); p2 *= 2 {
	}
	for i := len(rsm.in); i < p2; i++ {
		rsm.in = append(rsm.in, 0)
	}

	re, im := forwardFT(rsm.in)
	curLen := len(re) - 1
	for i := 0; i < curLen; i++ {
		re = append(re, 0)
		im = append(im, 0)
	}

	rsm.out = backwardFT(re, im)[:len(rsm.out)]*/

	// downsampling
	re, im := forwardFT(rsm.in)
	n := float32(len(rsm.in))
	normalize(re, im, len(rsm.in), n)
	// len(rsm.out) == (len(re) - 1) * 2 => len(rsm.out) / 2 + 1 = len(re)
	if rsm.outRate == rsm.inRate/2 {
		re = re[:len(rsm.out)/2+1]
		im = im[:len(rsm.out)/2+1]
	}
	if rsm.outRate == 8000 && rsm.inRate == 11025 {
		re = re[:len(rsm.out)/2+1]
		im = im[:len(rsm.out)/2+1]
	}
	rsm.out = backwardFT(re, im, float32(len(rsm.out)))
}
