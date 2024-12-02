package resample

// originally use int16 there but, there is no reason as will do one more type cast
var coefficients48To32 [][]int32 = [][]int32{
	{778, -2050, 1087, 23285, 12903, -3783, 441, 222},
	{222, 441, -3783, 12903, 23285, 1087, -2050, 778},
}

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

func Resample48To32L(In []int16) []int16 { // legacy version
	var tmp int32

	Out := make([]int16, len(In)/3*2)
	for i := 0; i < len(In)/3; i++ {
		if 3*i+8 >= len(In) {
			break
		}
		tmp = 1 << 14
		tmp += coefficients48To32[0][0] * int32(In[3*i+0])
		tmp += coefficients48To32[0][1] * int32(In[3*i+1])
		tmp += coefficients48To32[0][2] * int32(In[3*i+2])
		tmp += coefficients48To32[0][3] * int32(In[3*i+3])
		tmp += coefficients48To32[0][4] * int32(In[3*i+4])
		tmp += coefficients48To32[0][5] * int32(In[3*i+5])
		tmp += coefficients48To32[0][6] * int32(In[3*i+6])
		tmp += coefficients48To32[0][7] * int32(In[3*i+7])
		Out[2*i] = int16((tmp - (1 << 14)) >> 15) // also in legacy version such conversion would be done later in downsample/upsample funcs

		tmp = 1 << 14
		tmp += coefficients48To32[1][0] * int32(In[3*i+1])
		tmp += coefficients48To32[1][1] * int32(In[3*i+2])
		tmp += coefficients48To32[1][2] * int32(In[3*i+3])
		tmp += coefficients48To32[1][3] * int32(In[3*i+4])
		tmp += coefficients48To32[1][4] * int32(In[3*i+5])
		tmp += coefficients48To32[1][5] * int32(In[3*i+6])
		tmp += coefficients48To32[1][6] * int32(In[3*i+7])
		tmp += coefficients48To32[1][7] * int32(In[3*i+8])
		Out[2*i+1] = int16((tmp - (1 << 14)) >> 15)
	}

	return Out
}

func Resample48To32(In []int16) []int16 {
	var tmp int32

	Out := make([]int16, len(In)/3*2)
	for i := 0; i < len(In)/3; i++ {
		if 3*i+8 >= len(In) {
			break
		}
		tmp = 0
		tmp += coefficients48To32[0][0] * int32(In[3*i+0])
		tmp += coefficients48To32[0][1] * int32(In[3*i+1])
		tmp += coefficients48To32[0][2] * int32(In[3*i+2])
		tmp += coefficients48To32[0][3] * int32(In[3*i+3])
		tmp += coefficients48To32[0][4] * int32(In[3*i+4])
		tmp += coefficients48To32[0][5] * int32(In[3*i+5])
		tmp += coefficients48To32[0][6] * int32(In[3*i+6])
		tmp += coefficients48To32[0][7] * int32(In[3*i+7])
		Out[2*i] = int16(max(min(tmp>>15, 32767), -32768))

		tmp = 0
		tmp += coefficients48To32[1][0] * int32(In[3*i+1])
		tmp += coefficients48To32[1][1] * int32(In[3*i+2])
		tmp += coefficients48To32[1][2] * int32(In[3*i+3])
		tmp += coefficients48To32[1][3] * int32(In[3*i+4])
		tmp += coefficients48To32[1][4] * int32(In[3*i+5])
		tmp += coefficients48To32[1][5] * int32(In[3*i+6])
		tmp += coefficients48To32[1][6] * int32(In[3*i+7])
		tmp += coefficients48To32[1][7] * int32(In[3*i+8])
		Out[2*i+1] = int16(max(min(tmp>>15, 32767), -32768))
	}

	return Out
}
