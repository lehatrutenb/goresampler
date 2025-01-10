package resamplerl

import (
	"errors"
)

var (
	ErrIncorrectInLen = errors.New("expected strict input len, got not matched")
)

func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

func S32ToS16Cut(x int32) int16 {
	return int16(max(min(x, 32767), -32768))
}

// filter coefs
var kResampleAllpass1 []uint16 = []uint16{3284, 24441, 49528}
var kResampleAllpass2 []uint16 = []uint16{12199, 37471, 60255}

func mulAccum1(a uint16, b, c int32) int32 {
	return (c + (b>>16)*int32(a) + int32((uint32(b&0x0000FFFF)*uint32(a))>>16))
}

func mulAccum2(a uint16, b, c int32) int32 {
	return mulAccum1(a, b, c)
}

// originally use int16 there but, there is no reason as will do one more type cast
var coefficients48To32 [][]int32 = [][]int32{
	{778, -2050, 1087, 23285, 12903, -3783, 441, 222},
	{222, 441, -3783, 12903, 23285, 1087, -2050, 778},
}

// k == len(In)/3 == len(Out)/2
func Resample48To32L(in []int32, out []int32, k int32) {
	var tmp, i int32

	for i = 0; i < k; i++ {
		if 3*i+8 >= k*3 {
			break
		}
		tmp = 0
		tmp += coefficients48To32[0][0] * in[3*i+0]
		tmp += coefficients48To32[0][1] * in[3*i+1]
		tmp += coefficients48To32[0][2] * in[3*i+2]
		tmp += coefficients48To32[0][3] * in[3*i+3]
		tmp += coefficients48To32[0][4] * in[3*i+4]
		tmp += coefficients48To32[0][5] * in[3*i+5]
		tmp += coefficients48To32[0][6] * in[3*i+6]
		tmp += coefficients48To32[0][7] * in[3*i+7]
		out[2*i] = tmp

		tmp = 0
		tmp += coefficients48To32[1][0] * in[3*i+1]
		tmp += coefficients48To32[1][1] * in[3*i+2]
		tmp += coefficients48To32[1][2] * in[3*i+3]
		tmp += coefficients48To32[1][3] * in[3*i+4]
		tmp += coefficients48To32[1][4] * in[3*i+5]
		tmp += coefficients48To32[1][5] * in[3*i+6]
		tmp += coefficients48To32[1][6] * in[3*i+7]
		tmp += coefficients48To32[1][7] * in[3*i+8]
		out[2*i+1] = tmp
	}
}

func InitStateResample16To8L(st1 *[]int32, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	*st1 = make([]int32, 8)
}

func ResetResample16To8L(st1 []int32) {
	for i := 0; i < len(st1); i++ {
		st1[i] = 0
	}
}

func Resample16To8L(in []int16, out []int16) {
	var st1 []int32
	InitStateResample16To8L(&st1, &out, len(in)/2)

	DownsampleBy2L(in, len(in), out, st1)
}

// DownsampleBy2 - downsample in by 2 to out with filster state - fSt at the beginning
func DownsampleBy2L(in []int16, inLen int, out []int16, fSt []int32) {
	var tmp1, tmp2, diff, in32, out32 int32
	state0 := fSt[0]
	state1 := fSt[1]
	state2 := fSt[2]
	state3 := fSt[3]
	state4 := fSt[4]
	state5 := fSt[5]
	state6 := fSt[6]
	state7 := fSt[7]

	for i := 0; i < (inLen >> 1); i++ {
		in32 = int32(in[i*2]) * (1 << 10)
		diff = in32 - state1
		tmp1 = mulAccum1(kResampleAllpass2[0], diff, state0)
		state0 = in32
		diff = tmp1 - state2
		tmp2 = mulAccum2(kResampleAllpass2[1], diff, state1)
		state1 = tmp1
		diff = tmp2 - state3
		state3 = mulAccum2(kResampleAllpass2[2], diff, state2)
		state2 = tmp2

		// upper allpass filter
		in32 = int32(in[i*2+1]) * (1 << 10)
		diff = in32 - state5
		tmp1 = mulAccum1(kResampleAllpass1[0], diff, state4)
		state4 = in32
		diff = tmp1 - state6
		tmp2 = mulAccum1(kResampleAllpass1[1], diff, state5)
		state5 = tmp1
		diff = tmp2 - state7
		state7 = mulAccum2(kResampleAllpass1[2], diff, state6)
		state6 = tmp2

		// add two allpass outputs, divide by two and round
		out32 = (state3 + state7 + 1024) >> 11

		// limit amplitude to prevent wrap-around, and write to output array
		out[i] = S32ToS16Cut(out32)
	}

	fSt[0] = state0
	fSt[1] = state1
	fSt[2] = state2
	fSt[3] = state3
	fSt[4] = state4
	fSt[5] = state5
	fSt[6] = state6
	fSt[7] = state7
}

func InitStateResample8To16L(st1 *[]int32, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	*st1 = make([]int32, 8)
}

func ResetResample8To16L(st1 []int32) {
	for i := 0; i < len(st1); i++ {
		st1[i] = 0
	}
}

func Resample8To16L(in []int16, out []int16) {
	var st1 []int32
	InitStateResample8To16L(&st1, &out, len(in)*2)

	UpsampleBy2L(in, len(in), out, st1)
}

func UpsampleBy2L(in []int16, inLen int, out []int16, fSt []int32) {
	var tmp1, tmp2, diff, in32, out32 int32
	state0 := fSt[0]
	state1 := fSt[1]
	state2 := fSt[2]
	state3 := fSt[3]
	state4 := fSt[4]
	state5 := fSt[5]
	state6 := fSt[6]
	state7 := fSt[7]

	for i := 0; i < inLen; i++ {
		// lower allpass filter
		in32 = int32(in[i]) * (1 << 10)
		diff = in32 - state1
		tmp1 = mulAccum1(kResampleAllpass1[0], diff, state0)
		state0 = in32
		diff = tmp1 - state2
		tmp2 = mulAccum1(kResampleAllpass1[1], diff, state1)
		state1 = tmp1
		diff = tmp2 - state3
		state3 = mulAccum2(kResampleAllpass1[2], diff, state2)
		state2 = tmp2

		// round limit amplitude to prevent wrap-around write to output array
		out32 = (state3 + 512) >> 10
		out[i*2] = S32ToS16Cut(out32)

		// upper allpass filter
		diff = in32 - state5
		tmp1 = mulAccum1(kResampleAllpass2[0], diff, state4)
		state4 = in32
		diff = tmp1 - state6
		tmp2 = mulAccum2(kResampleAllpass2[1], diff, state5)
		state5 = tmp1
		diff = tmp2 - state7
		state7 = mulAccum2(kResampleAllpass2[2], diff, state6)
		state6 = tmp2

		// round limit amplitude to prevent wrap-around write to output array
		out32 = (state7 + 512) >> 10
		out[i*2+1] = S32ToS16Cut(out32)
	}

	fSt[0] = state0
	fSt[1] = state1
	fSt[2] = state2
	fSt[3] = state3
	fSt[4] = state4
	fSt[5] = state5
	fSt[6] = state6
	fSt[7] = state7
}

// it is int16 legacy, but used as int32
var kResampleAllpass [][]int32 = [][]int32{
	{821, 6110, 12382},
	{3050, 9368, 15063},
}

// lowpass filter
// input:  int16_t
// output: int32_t (normalized, not saturated)
// state:  filter state array; length = 8
func LPBy2ShortToIntL(in []int16, inLen int, out []int32, fSt []int32) {
	var tmp0, tmp1, diff int32
	var i int

	inLen = inLen >> 1

	// lower allpass filter: odd input -> even output samples
	// initial state of polyphase delay element
	tmp0 = fSt[12]
	for i = 0; i < inLen; i++ {
		diff = tmp0 - fSt[1]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = fSt[0] + diff*kResampleAllpass[1][0]
		fSt[0] = tmp0
		diff = tmp1 - fSt[2]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = fSt[1] + diff*kResampleAllpass[1][1]
		fSt[1] = tmp1
		diff = tmp0 - fSt[3]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		fSt[3] = fSt[2] + diff*kResampleAllpass[1][2]
		fSt[2] = tmp0

		// scale down, round and store
		out[i<<1] = fSt[3] >> 1
		tmp0 = (int32(in[1+(i<<1)]) << 15) + (1 << 14) // 1 + (i<<1) check 159 line (before for)
	}

	// upper allpass filter: even input -> even output samples
	for i = 0; i < inLen; i++ {
		tmp0 = (int32(in[i<<1]) << 15) + (1 << 14)
		diff = tmp0 - fSt[5]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = fSt[4] + diff*kResampleAllpass[0][0]
		fSt[4] = tmp0
		diff = tmp1 - fSt[6]
		// scale down and round
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = fSt[5] + diff*kResampleAllpass[0][1]
		fSt[5] = tmp1
		diff = tmp0 - fSt[7]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		fSt[7] = fSt[6] + diff*kResampleAllpass[0][2]
		fSt[6] = tmp0

		// average the two allpass outputs, scale down and store
		out[i<<1] = (out[i<<1] + (fSt[7] >> 1)) >> 15
	}

	// switch to odd output samples
	// lower allpass filter: even input -> odd output samples
	for i = 0; i < inLen; i++ {
		tmp0 = (int32(in[i<<1]) << 15) + (1 << 14)
		diff = tmp0 - fSt[9]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = fSt[8] + diff*kResampleAllpass[1][0]
		fSt[8] = tmp0
		diff = tmp1 - fSt[10]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = fSt[9] + diff*kResampleAllpass[1][1]
		fSt[9] = tmp1
		diff = tmp0 - fSt[11]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		fSt[11] = fSt[10] + diff*kResampleAllpass[1][2]
		fSt[10] = tmp0

		// scale down, round and store
		out[1+i<<1] = fSt[11] >> 1
	}

	// upper allpass filter: odd input -> odd output samples
	for i = 0; i < inLen; i++ {
		tmp0 = (int32(in[1+i<<1]) << 15) + (1 << 14)
		diff = tmp0 - fSt[13]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = fSt[12] + diff*kResampleAllpass[0][0]
		fSt[12] = tmp0
		diff = tmp1 - fSt[14]
		// scale down and round
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = fSt[13] + diff*kResampleAllpass[0][1]
		fSt[13] = tmp1
		diff = tmp0 - fSt[15]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		fSt[15] = fSt[14] + diff*kResampleAllpass[0][2]
		fSt[14] = tmp0

		// average the two allpass outputs, scale down and store
		out[1+i<<1] = (out[1+i<<1] + (fSt[15] >> 1)) >> 15
	}
}

//   decimator
// input:  int32_t (shifted 15 positions to the left, + offset 16384) OVERWRITTEN!
// output: int16_t (saturated) (of length len/2)
// state:  filter state array; length = 8

// void RTC_NO_SANITIZE("signed-integer-overflow")  // bugs.webrtc.org/5486 TODO so care there is an UB
func DownBy2IntToShortL(in []int32, inLen int, out []int16, fSt []int32) {
	var tmp0, tmp1, diff int32
	var i int

	inLen = inLen >> 1

	// lower allpass filter (operates on even input samples)
	for i = 0; i < inLen; i++ {
		tmp0 = in[i<<1]
		diff = tmp0 - fSt[1]
		// UBSan: -1771017321 - 999586185 cannot be represented in type 'int'

		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = fSt[0] + diff*kResampleAllpass[1][0]
		fSt[0] = tmp0
		diff = tmp1 - fSt[2]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = fSt[1] + diff*kResampleAllpass[1][1]
		fSt[1] = tmp1
		diff = tmp0 - fSt[3]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		fSt[3] = fSt[2] + diff*kResampleAllpass[1][2]
		fSt[2] = tmp0

		// divide by two and store temporarily
		in[i<<1] = (fSt[3] >> 1)
	}

	// upper allpass filter (operates on odd input samples)
	for i = 0; i < inLen; i++ {
		tmp0 = in[1+i<<1]
		diff = tmp0 - fSt[5]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = fSt[4] + diff*kResampleAllpass[0][0]
		fSt[4] = tmp0
		diff = tmp1 - fSt[6]
		// scale down and round
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = fSt[5] + diff*kResampleAllpass[0][1]
		fSt[5] = tmp1
		diff = tmp0 - fSt[7]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		fSt[7] = fSt[6] + diff*kResampleAllpass[0][2]
		fSt[6] = tmp0

		// divide by two and store temporarily
		in[1+i<<1] = (fSt[7] >> 1)
	}

	// combine allpass outputs
	for i = 0; i < inLen; i += 2 {
		// divide by two, add both allpass outputs and round
		tmp0 = (in[i<<1] + in[(i<<1)+1]) >> 15
		tmp1 = (in[(i<<1)+2] + in[(i<<1)+3]) >> 15
		if tmp0 > 32767 { // 0x00007FFF
			tmp0 = 32767
		}
		if tmp0 < -32768 { // 0xFFFF8000
			tmp0 = -32768
		}
		out[i] = int16(tmp0) // TODO looks like need to S32ToS16Cut
		if tmp1 > 32767 {    // 0x00007FFF
			tmp1 = 32767
		}
		if tmp1 < -32768 { // 0xFFFF8000
			tmp1 = -32768
		}
		out[i+1] = int16(tmp1) // TODO looks like need to S32ToS16Cut
	}
}

//	interpolator
//
// input:  int16_t
// output: int32_t (normalized, not saturated) (of length len*2)
// state:  filter state array; length = 8
func UpsampleBy2ShortToIntL(in []int16, inLen int32, out []int32, state []int32) {
	var tmp0, tmp1, diff, i int32

	// upper allpass filter (generates odd output samples)
	for i = 0; i < inLen; i++ {
		tmp0 = (int32(in[i]) << 15) + (1 << 14)
		diff = tmp0 - state[5]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = state[4] + diff*kResampleAllpass[0][0]
		state[4] = tmp0
		diff = tmp1 - state[6]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[5] + diff*kResampleAllpass[0][1]
		state[5] = tmp1
		diff = tmp0 - state[7]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[7] = state[6] + diff*kResampleAllpass[0][2]
		state[6] = tmp0

		// scale down, round and store
		out[i<<1] = state[7] >> 15
	}

	// lower allpass filter (generates even output samples)
	for i = 0; i < inLen; i++ {
		tmp0 = (int32(in[i]) << 15) + (1 << 14)
		diff = tmp0 - state[1]
		// scale down and round
		diff = (diff + (1 << 13)) >> 14
		tmp1 = state[0] + diff*kResampleAllpass[1][0]
		state[0] = tmp0
		diff = tmp1 - state[2]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[1] + diff*kResampleAllpass[1][1]
		state[1] = tmp1
		diff = tmp0 - state[3]
		// scale down and truncate
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[3] = state[2] + diff*kResampleAllpass[1][2]
		state[2] = tmp0

		// scale down, round and store
		out[1+i<<1] = state[3] >> 15
	}
}

type state48To16 struct {
	S_48_48 []int32 // 16
	S_48_32 []int32 // 8
	S_32_16 []int32 // 8
}

func resample48To16L(in []int16, out []int16, state state48To16, tmpmem []int32) {
	const inLen = 480
	///// 48 --> 48(LP) /////
	// int16_t  in[480]
	// int32_t out[480]
	/////
	LPBy2ShortToIntL(in, inLen, tmpmem[16:], state.S_48_48)

	///// 48 --> 32 /////
	// int32_t  in[480]
	// int32_t out[320]
	/////
	// copy state to and from input array
	//memcpy(tmpmem + 8, state->S_48_32, 8 * sizeof(int32_t))
	//memcpy(state->S_48_32, tmpmem + 488, 8 * sizeof(int32_t))
	for i := 0; i < 8; i++ {
		tmpmem[8+i] = state.S_48_32[i]
	}
	for i := 0; i < 8; i++ {
		state.S_48_32[i] = tmpmem[488+i]
	}

	Resample48To32L(tmpmem[8:], tmpmem, 160)

	///// 32 --> 16 /////
	// int32_t  in[320]
	// int16_t out[160]
	/////
	DownBy2IntToShortL(tmpmem, 320, out, state.S_32_16)
}

//
// fractional resampling filters
//   Fout = 11/16 * Fin
//   Fout =  8/11 * Fin
//

// compute two inner-products and store them to output array
func ResampDotProductL(in1 []int32, in2 []int32, in2Len int32, coef_ptr []int16, out1 []int32, out2 []int32) {
	var tmp1 int32 = 16384
	var tmp2 int32 = 16384
	var coef int16

	coef = coef_ptr[0]
	tmp1 += int32(coef) * in1[0]
	tmp2 += int32(coef) * in2[in2Len-1]

	coef = coef_ptr[1]
	tmp1 += int32(coef) * in1[1]
	tmp2 += int32(coef) * in2[in2Len-2]

	coef = coef_ptr[2]
	tmp1 += int32(coef) * in1[2]
	tmp2 += int32(coef) * in2[in2Len-3]

	coef = coef_ptr[3]
	tmp1 += int32(coef) * in1[3]
	tmp2 += int32(coef) * in2[in2Len-4]

	coef = coef_ptr[4]
	tmp1 += int32(coef) * in1[4]
	tmp2 += int32(coef) * in2[in2Len-5]

	coef = coef_ptr[5]
	tmp1 += int32(coef) * in1[5]
	tmp2 += int32(coef) * in2[in2Len-6]

	coef = coef_ptr[6]
	tmp1 += int32(coef) * in1[6]
	tmp2 += int32(coef) * in2[in2Len-7]

	coef = coef_ptr[7]
	tmp1 += int32(coef) * in1[7]
	tmp2 += int32(coef) * in2[in2Len-8]

	coef = coef_ptr[8]
	out1[0] = tmp1 + int32(coef)*in1[8]
	out2[0] = tmp2 + int32(coef)*in2[in2Len-9]
}

var kCoefficients44To32 [][]int16 = [][]int16{
	{117, -669, 2245, -6183, 26267, 13529, -3245, 845, -138},
	{-101, 612, -2283, 8532, 29790, -5138, 1789, -524, 91},
	{50, -292, 1016, -3064, 32010, 3933, -1147, 315, -53},
	{-156, 974, -3863, 18603, 21691, -6246, 2353, -712, 126},
}

//   Resampling ratio: 8/11
// input:  int32_t (normalized, not saturated) :: size 11 * K
// output: int32_t (shifted 15 positions to the left, + offset 16384) :: size  8 * K
//      K: number of blocks

func Resample44To32L(in []int32, out []int32, k int32) {
	/////////////////////////////////////////////////////////////
	// Filter operation:
	//
	// Perform resampling (11 input samples -> 8 output samples);
	// process in sub blocks of size 11 samples.
	var tmp int32
	var i int32

	for i = 0; i < k; i++ {
		tmp = 1 << 14

		// first output sample
		out[i*8+0] = (in[3] << 15) + tmp

		// sum and accumulate filter coefficients and input samples
		tmp += int32(kCoefficients44To32[3][0]) * in[i*11+5]
		tmp += int32(kCoefficients44To32[3][1]) * in[i*11+6]
		tmp += int32(kCoefficients44To32[3][2]) * in[i*11+7]
		tmp += int32(kCoefficients44To32[3][3]) * in[i*11+8]
		tmp += int32(kCoefficients44To32[3][4]) * in[i*11+9]
		tmp += int32(kCoefficients44To32[3][5]) * in[i*11+10]
		tmp += int32(kCoefficients44To32[3][6]) * in[i*11+11]
		tmp += int32(kCoefficients44To32[3][7]) * in[i*11+12]
		tmp += int32(kCoefficients44To32[3][8]) * in[i*11+13]
		out[i*8+4] = tmp

		// sum and accumulate filter coefficients and input samples
		ResampDotProductL(in[i*11+0:], in[:1+i*11+17], 1+i*11+17, kCoefficients44To32[0], out[i*8+1:], out[i*8+7:])

		// sum and accumulate filter coefficients and input samples
		ResampDotProductL(in[i*11+2:], in[:1+i*11+15], 1+i*11+15, kCoefficients44To32[1], out[i*8+2:], out[i*8+6:])

		// sum and accumulate filter coefficients and input samples
		ResampDotProductL(in[i*11+3:], in[:1+i*11+14], 1+i*11+14, kCoefficients44To32[2], out[i*8+3:], out[i*8+5:])
	}
}

type state22To16 struct {
	S_22_44 []int32 // 8
	S_44_32 []int32 // 8
	S_32_16 []int32 // 8
}

// number of subblocks; options: 1, 2, 4, 5, 10
const SUB_BLOCKS_22_16 = 5

// 22 -> 16 resampler
func Resample22To16L(in []int16, out []int16, state state22To16, tmpmem []int32) {
	var i int32

	// process two blocks of 10/SUB_BLOCKS_22_16 ms (to reduce temp buffer size)
	for i = 0; i < SUB_BLOCKS_22_16; i++ {
		///// 22 --> 44 /////
		// int16_t  in[220/SUB_BLOCKS_22_16]
		// int32_t out[440/SUB_BLOCKS_22_16]
		/////
		UpsampleBy2ShortToIntL(in[i*220/SUB_BLOCKS_22_16:], 220/SUB_BLOCKS_22_16, tmpmem[16:], state.S_22_44)

		///// 44 --> 32 /////
		// int32_t  in[440/SUB_BLOCKS_22_16]
		// int32_t out[320/SUB_BLOCKS_22_16]
		/////
		// copy state to and from input array
		tmpmem[8] = state.S_44_32[0]
		tmpmem[9] = state.S_44_32[1]
		tmpmem[10] = state.S_44_32[2]
		tmpmem[11] = state.S_44_32[3]
		tmpmem[12] = state.S_44_32[4]
		tmpmem[13] = state.S_44_32[5]
		tmpmem[14] = state.S_44_32[6]
		tmpmem[15] = state.S_44_32[7]
		state.S_44_32[0] = tmpmem[440/SUB_BLOCKS_22_16+8]
		state.S_44_32[1] = tmpmem[440/SUB_BLOCKS_22_16+9]
		state.S_44_32[2] = tmpmem[440/SUB_BLOCKS_22_16+10]
		state.S_44_32[3] = tmpmem[440/SUB_BLOCKS_22_16+11]
		state.S_44_32[4] = tmpmem[440/SUB_BLOCKS_22_16+12]
		state.S_44_32[5] = tmpmem[440/SUB_BLOCKS_22_16+13]
		state.S_44_32[6] = tmpmem[440/SUB_BLOCKS_22_16+14]
		state.S_44_32[7] = tmpmem[440/SUB_BLOCKS_22_16+15]

		Resample44To32L(tmpmem[8:], tmpmem, 40/SUB_BLOCKS_22_16)

		///// 32 --> 16 /////
		// int32_t  in[320/SUB_BLOCKS_22_16]
		// int32_t out[160/SUB_BLOCKS_22_16]
		/////
		DownBy2IntToShortL(tmpmem, 320/SUB_BLOCKS_22_16, out[i*160/SUB_BLOCKS_22_16:], state.S_32_16)
	}
}

type State22To8 struct {
	S_22_22 []int32 // 16
	S_22_16 []int32 // 8
	S_16_8  []int32 // 8
}

// number of subblocks; options: 1, 2, 5, 10
const SUB_BLOCKS_22_8 = 2

// 22 -> 8 resampler
func Resample22To8L(in []int16, out []int16, state State22To8, tmpmem []int32) {
	var i int32

	// process two blocks of 10/SUB_BLOCKS_22_8 ms (to reduce temp buffer size)
	for i = 0; i < SUB_BLOCKS_22_8; i++ {
		///// 22 --> 22 lowpass /////
		// int16_t  in[220/SUB_BLOCKS_22_8]
		// int32_t out[220/SUB_BLOCKS_22_8]
		/////
		LPBy2ShortToIntL(in[i*220/SUB_BLOCKS_22_8:], 220/SUB_BLOCKS_22_8, tmpmem[16:], state.S_22_22)

		///// 22 --> 16 /////
		// int32_t  in[220/SUB_BLOCKS_22_8]
		// int32_t out[160/SUB_BLOCKS_22_8]
		/////
		// copy state to and from input array
		tmpmem[8] = state.S_22_16[0]
		tmpmem[9] = state.S_22_16[1]
		tmpmem[10] = state.S_22_16[2]
		tmpmem[11] = state.S_22_16[3]
		tmpmem[12] = state.S_22_16[4]
		tmpmem[13] = state.S_22_16[5]
		tmpmem[14] = state.S_22_16[6]
		tmpmem[15] = state.S_22_16[7]
		state.S_22_16[0] = tmpmem[220/SUB_BLOCKS_22_8+8]
		state.S_22_16[1] = tmpmem[220/SUB_BLOCKS_22_8+9]
		state.S_22_16[2] = tmpmem[220/SUB_BLOCKS_22_8+10]
		state.S_22_16[3] = tmpmem[220/SUB_BLOCKS_22_8+11]
		state.S_22_16[4] = tmpmem[220/SUB_BLOCKS_22_8+12]
		state.S_22_16[5] = tmpmem[220/SUB_BLOCKS_22_8+13]
		state.S_22_16[6] = tmpmem[220/SUB_BLOCKS_22_8+14]
		state.S_22_16[7] = tmpmem[220/SUB_BLOCKS_22_8+15]

		Resample44To32L(tmpmem[8:], tmpmem, 20/SUB_BLOCKS_22_8)

		///// 16 --> 8 /////
		// int32_t in[160/SUB_BLOCKS_22_8]
		// int32_t out[80/SUB_BLOCKS_22_8]
		/////
		DownBy2IntToShortL(tmpmem, 160/SUB_BLOCKS_22_8, out[i*80/SUB_BLOCKS_22_8:], state.S_16_8)
	}
}

func InitStateResample48To8L(st1 *state48To16, st2 *[]int32, tmpMem *[]int32, tmp *[]int16, inLen int, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	st1.S_48_48 = make([]int32, 16)
	st1.S_48_32 = make([]int32, 8)
	st1.S_32_16 = make([]int32, 8)
	*st2 = make([]int32, 8)

	*tmpMem = make([]int32, 496)
	*tmp = make([]int16, inLen/3)
}

func ResetResample48To8L(st1 state48To16, st2 []int32) {
	for i := 0; i < 8; i++ {
		st1.S_48_32[i] = 0
		st1.S_32_16[i] = 0
		st2[i] = 0
	}
	for i := 0; i < 16; i++ {
		st1.S_48_48[i] = 0
	}
}

func Resample48To8L(in []int16, out []int16) error {
	if len(in)%480 != 0 {
		return ErrIncorrectInLen
	}

	var st1 state48To16
	var st2, tmpMem []int32
	var tmp []int16
	InitStateResample48To8L(&st1, &st2, &tmpMem, &tmp, len(in), &out, len(in)/6)
	for i := 0; i < len(in); i += 480 {
		resample48To16L(in[i:], tmp[i/3:], st1, tmpMem)
	}

	DownsampleBy2L(tmp, len(in)/3, out, st2)

	return nil
}

func InitStateResample48To16L(st1 *state48To16, tmpMem *[]int32, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	st1.S_48_48 = make([]int32, 16)
	st1.S_48_32 = make([]int32, 8)
	st1.S_32_16 = make([]int32, 8)
	*tmpMem = make([]int32, 496)
}

func ResetResample48To16L(st1 state48To16) {
	for i := 0; i < 8; i++ {
		st1.S_48_32[i] = 0
		st1.S_32_16[i] = 0
	}
	for i := 0; i < 16; i++ {
		st1.S_48_48[i] = 0
	}
}

func Resample48To16L(in []int16, out []int16) error {
	if len(in)%480 != 0 {
		return ErrIncorrectInLen
	}

	var st1 state48To16
	var tmpMem []int32
	InitStateResample48To16L(&st1, &tmpMem, &out, len(in)/3)
	for i := 0; i < len(in); i += 480 {
		resample48To16L(in[i:], out[i/3:], st1, tmpMem)
	}

	return nil
}

func InitStateResample11To8L(st1 *state22To16, tmpMem *[]int32, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	st1.S_22_44 = make([]int32, 8)
	st1.S_44_32 = make([]int32, 8)
	st1.S_32_16 = make([]int32, 8)
	*tmpMem = make([]int32, 104)
}

func ResetResample11To8L(st1 state22To16) {
	for i := 0; i < 8; i++ {
		st1.S_22_44[i] = 0
		st1.S_44_32[i] = 0
		st1.S_32_16[i] = 0
	}
}

func Resample11To8L(in []int16, out []int16) error {
	if len(in)%220 != 0 {
		return ErrIncorrectInLen
	}

	var st1 state22To16
	var tmpMem []int32
	InitStateResample11To8L(&st1, &tmpMem, &out, (len(in)*8)/11)
	for i := 0; i < len(in); i += 220 {
		Resample22To16L(in[i:], out[(i*8)/11:], st1, tmpMem)
	}

	return nil
}

func InitStateResample11To16L(st1 *[]int32, st2 *state22To16, tmpMem *[]int32, tmp *[]int16, inLen int, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	*st1 = make([]int32, 8)
	st2.S_22_44 = make([]int32, 8)
	st2.S_44_32 = make([]int32, 8)
	st2.S_32_16 = make([]int32, 8)
	*tmpMem = make([]int32, 104)
	*tmp = make([]int16, inLen*2)
}

func ResetResample11To16L(st1 []int32, st2 state22To16) {
	for i := 0; i < 8; i++ {
		st1[i] = 0
		st2.S_22_44[i] = 0
		st2.S_44_32[i] = 0
		st2.S_32_16[i] = 0
	}
}

func Resample11To16L(in []int16, out []int16) error {
	if len(in)%110 != 0 {
		return ErrIncorrectInLen
	}

	var st1 []int32
	var st2 state22To16
	var tmpMem []int32
	var tmp []int16
	InitStateResample11To16L(&st1, &st2, &tmpMem, &tmp, len(in), &out, (len(in)*16)/11)

	UpsampleBy2L(in, len(in), tmp, st1)
	for i := 0; i < len(in)*2; i += 220 {
		Resample22To16L(tmp[i:], out[(i/220)*160:], st2, tmpMem)
	}

	return nil
}

func InitStateResample44To8L(st1 *[]int32, st2 *State22To8, tmpMem *[]int32, tmp *[]int16, inLen int, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	*st1 = make([]int32, 8)
	st2.S_22_22 = make([]int32, 16)
	st2.S_22_16 = make([]int32, 8)
	st2.S_16_8 = make([]int32, 8)
	*tmpMem = make([]int32, 126)
	*tmp = make([]int16, (inLen*4)/11)
}

func ResetResample44to8L(st1 []int32, st2 State22To8) {
	for i := 0; i < 16; i++ {
		st2.S_22_22[i] = 0
	}
	for i := 0; i < 8; i++ {
		st1[i] = 0
		st2.S_22_16[i] = 0
		st2.S_16_8[i] = 0
	}
}

func Resample44To8L(in []int16, out []int16) error {
	if len(in)%220 != 0 {
		return ErrIncorrectInLen
	}

	var st1 []int32
	var st2 State22To8
	var tmpMem []int32
	var tmp []int16
	InitStateResample44To8L(&st1, &st2, &tmpMem, &tmp, len(in), &out, (len(in)*2)/11)

	for i := 0; i < len(in)*2; i += 220 {
		Resample22To8L(in[i:], tmp[(i*4)/11:], st2, tmpMem)
	}
	DownsampleBy2L(tmp, (len(in)*4)/11, out, st1)

	return nil
}

func InitStateResample44To16L(st1 *State22To8, tmpMem *[]int32, out *[]int16, outLen int) {
	*out = make([]int16, outLen)
	st1.S_22_22 = make([]int32, 16)
	st1.S_22_16 = make([]int32, 8)
	st1.S_16_8 = make([]int32, 8)
	*tmpMem = make([]int32, 126)
}

func ResetResample44To16L(st1 State22To8) {
	for i := 0; i < 16; i++ {
		st1.S_22_22[i] = 0
	}
	for i := 0; i < 8; i++ {
		st1.S_22_16[i] = 0
		st1.S_16_8[i] = 0
	}
}

func Resample44To16L(in []int16, out []int16) error {
	if len(in)%220 != 0 {
		return ErrIncorrectInLen
	}

	var st1 State22To8
	var tmpMem []int32
	InitStateResample44To16L(&st1, &tmpMem, &out, (len(in)*4)/11)

	for i := 0; i < len(in)*2; i += 220 {
		Resample22To8L(in[i:], out[(i*4)/11:], st1, tmpMem)
	}

	return nil
}
