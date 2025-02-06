package resamplerfft

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
		cSinT, cCosT := math.Sincos(float64(pi() / float32(int(1)<<layer)))
		cSin, cCos := -float32(cSinT), float32(cCosT)
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
			cRe, cIm = cRe*cCos-cIm*cSin, cRe*cSin+cIm*cCos
		}
	}
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

func chirpFilterCoefs(arrSz int) (forwRe, forwIm []float32, backwRe, backwIm []float32) {
	res := make([]complex64, arrSz)
	forwRe, forwIm = make([]float32, arrSz), make([]float32, arrSz)
	backwRe, backwIm = make([]float32, arrSz), make([]float32, arrSz)
	for i := 0; i < len(res); i++ {
		cSin, cCos := float64(0), float64(1)
		if i != 0 {
			cSin, cCos = math.Sincos(math.Pi / float64(arrSz) * float64(i*i))
		}
		forwRe[i], forwIm[i] = float32(cCos), float32(cSin)
		backwRe[i], backwIm[i] = float32(cCos), -float32(cSin)
	}
	return
}

func BluesteinFFT(re []float32) ([]float32, []float32) {
	sLen := len(re)
	fForwRe, fForwIm, fBackwRe, fBackwIm := chirpFilterCoefs(len(re))
	setStrictP2Len(&re)
	if len(re) < 2*sLen-1 {
		re = append(re, 0)
		setStrictP2Len(&re)
	}
	im := make([]float32, len(re))

	for i := 0; i < sLen; i++ {
		im[i] = re[i] * fBackwIm[i]
		re[i] = re[i] * fBackwRe[i]
	}

	convArrRe := make([]float32, len(re))
	convArrIm := make([]float32, len(re))

	for i := 0; i < len(fForwRe); i++ {
		convArrRe[i] = fForwRe[i]
		convArrIm[i] = fForwIm[i]
		if i != 0 {
			convArrRe[len(re)-i] = fForwRe[i]
			convArrIm[len(re)-i] = fForwIm[i]
		}
	}

	forwardFFT(re, im)
	forwardFFT(convArrRe, convArrIm)
	for i := 0; i < len(re); i++ {
		re[i], im[i] = re[i]*convArrRe[i]-im[i]*convArrIm[i], re[i]*convArrIm[i]+im[i]*convArrRe[i]
	}
	backwardFFT(re, im, len(re))

	for i := 0; i < sLen; i++ {
		fRe, fIm := fBackwRe[i], fBackwIm[i]
		re[i], im[i] = re[i]*fRe-im[i]*fIm, re[i]*fIm+im[i]*fRe
	}

	return re[:sLen], im[:sLen]
}

func resample(re []float32, outArr []float32, inRate, outRate int, outLen int) {
	startLen := len(re)

	reCur := make([]float32, startLen)
	for i := 0; i < startLen; i++ {
		reCur[i] = re[i]
	}

	reCur, im := BluesteinFFT(reCur)

	changeSampleRate(&reCur, &im, inRate, outRate, outLen)

	fixFreqRulesAfterChangeFFT(reCur, im)

	backwardFFT(reCur, im, startLen)

	for i := 0; i < outLen; i++ {
		outArr[i] = reCur[i]
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
