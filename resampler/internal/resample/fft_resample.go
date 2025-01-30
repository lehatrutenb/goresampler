package resample

import (
	"math"
	"resampler/internal/utils"

	"github.com/mjibson/go-dsp/fft"
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

func normalizeFT(re []float32, im []float32, ni int, n float32) { // ni == int(n)
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

// little->big endian , but not groupped by bytes - swap bits
// len(arr) = 2^lgn
func arrBitReverse(arr []float32) {
	lgn := 0
	{
		cur := 1
		for cur < len(arr) {
			cur *= 2
			lgn++
		}
	}

	st := 0        // straight order
	back := 0      // back order
	mp2 := lgn - 1 // max 2 power of back order - first 0 in bit repr
	toAdd := 1 << mp2
	for i := 0; i+1 < len(arr); i++ {
		st++
		back ^= toAdd
		if mp2 == lgn-1 {
			for mp2 > 0 && ((1<<mp2)&back) != 0 { // find new first 0
				toAdd |= toAdd >> 1
				mp2--
			}
		} else {
			mp2 = lgn - 1
			toAdd = 1 << mp2
		}
		if st < back { // not to swap twice (and we will also swap every element)
			arr[st], arr[back] = arr[back], arr[st]
		}
	}
}

// len(re) = len(im) = 2^k
func forwardFFT(re, im []float32) {
	arrBitReverse(re)
	arrBitReverse(im)
	for layer := 0; (1 << layer) < len(re); layer++ {
		sin := -sin(pi() / float32(int(1)<<layer)) // will use to shift in freq domain
		cos := cos(pi() / float32(int(1)<<layer))
		cRe := float32(1) // currect multipliers not to copy paste code
		cIm := float32(0)

		for i := 0; i < (1 << layer); i++ {
			for j := i; j < len(re); j += (1 << (layer + 1)) { // layer + 1 cause want to merge 2 and jump over them
				jr := j + (1 << layer)
				chRe := re[jr]*cRe - im[jr]*cIm // 'butterfly'
				chIm := re[jr]*cIm + im[jr]*cRe

				re[jr] = re[j] - chRe
				im[jr] = im[j] - chIm
				re[j] += chRe
				im[j] += chIm
			}
			cRe, cIm = cRe*cos-cIm*sin, cRe*sin+cIm*cos
		}
	}
}

func BackwardFFT2(re, im []float32, startLen int) {
	for i := 0; i < len(re); i++ {
		im[i] *= -1
	}
	forwardFFT(re, im)
}

func backwardFFT(re, im []float32, startLen int) {
	for i := 0; i < len(re); i++ {
		im[i] *= -1
	}
	forwardFFT(re, im)
	nf := float32(startLen)
	for i := 0; i < len(re); i++ {
		re[i] /= nf
		im[i] /= -nf
	}
}

// make len eq to 2^k
func setStrictP2Len(arr *[]float32) {
	p2 := 1
	for p2 < len(*arr) {
		p2 <<= 1
	}

	for len(*arr) != p2 {
		*arr = append(*arr, 0)
	}
}

func changeSampleRate(re, im *[]float32, inR, outR int, outLen int) {
	if inR == outR { // TODO rm it for better performance
		return
	}
	if inR > outR {
		*re = (*re)[:outLen]
		*im = (*im)[:outLen]
		return
	}
}

func fix2powAfterChangeFFT(re, im *[]float32) {
	n := len(*re)
	setStrictP2Len(re)
	setStrictP2Len(im)

	curLen2 := len(*re) / 2

	for i := n / 2; i <= curLen2; i++ { // <= < TODO check
		(*re)[i] = 0
		(*im)[i] = 0
	}
}

func fixFreqRulesAfterChangeFFT(re, im []float32) {
	n := len(re)
	n2 := n / 2
	for i := 1; i < n2; i++ {
		re[i+n2] = re[n2-i]
		im[i+n2] = -im[n2-i]
	}
}

func resample(re []float32, outArr []float32, inRate, outRate int, outLen int) {
	startLen := len(re)
	im := make([]float32, len(re))
	in1 := make([]float64, len(re))
	for i := 0; i < len(re); i++ {
		in1[i] = float64(re[i])
	}
	res := fft.FFTReal(in1)
	for i := 0; i < len(re); i++ {
		re[i] = float32(real(res[i]))
		im[i] = float32(imag(res[i]))
	}

	changeSampleRate(&re, &im, inRate, outRate, outLen)

	fixFreqRulesAfterChangeFFT(re, im)

	backwardFFT(re, im, startLen)

	for i := 0; i < outLen; i++ {
		outArr[i] = re[i]
	}
}

func (rsm *FFTResampler) ResampleFFT() {
	if rsm.inRate == 11025 && rsm.outRate == 8000 {
		coefIn := 90317 // 22580
		coefOut := len(rsm.out) / (len(rsm.in) / coefIn)
		for i := 0; i*coefIn < len(rsm.in); i++ {
			resample(rsm.in[i*coefIn:(i+1)*coefIn], rsm.out[i*coefOut:(i+1)*coefOut], rsm.inRate, rsm.outRate, coefOut)
		}
	}
}

func (rsm *FFTResampler) ResampleFT() {
	// downsampling
	re, im := forwardFT(rsm.in)
	n := float32(len(rsm.in))
	normalizeFT(re, im, len(rsm.in), n)
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
