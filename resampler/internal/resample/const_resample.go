package resample

// originally use int16 there but, there is no reason as will do one more type cast
var coefficients48To32 [][]int32 = [][]int32{
	{778, -2050, 1087, 23285, 12903, -3783, 441, 222},
	{222, 441, -3783, 12903, 23285, 1087, -2050, 778},
}

func resample48To32(In []int32) []int32 {
	var tmp int32

	Out := make([]int32, len(In)/3*2)
	for i := 0; i < len(In)/3; i++ {
		if 3*i+8 >= len(In) {
			break
		}
		tmp = 1 << 14
		tmp += coefficients48To32[0][0] * In[3*i+0]
		tmp += coefficients48To32[0][1] * In[3*i+1]
		tmp += coefficients48To32[0][2] * In[3*i+2]
		tmp += coefficients48To32[0][3] * In[3*i+3]
		tmp += coefficients48To32[0][4] * In[3*i+4]
		tmp += coefficients48To32[0][5] * In[3*i+5]
		tmp += coefficients48To32[0][6] * In[3*i+6]
		tmp += coefficients48To32[0][7] * In[3*i+7]
		Out[2*i] = tmp

		tmp = 1 << 14
		tmp += coefficients48To32[1][0] * In[3*i+1]
		tmp += coefficients48To32[1][1] * In[3*i+2]
		tmp += coefficients48To32[1][2] * In[3*i+3]
		tmp += coefficients48To32[1][3] * In[3*i+4]
		tmp += coefficients48To32[1][4] * In[3*i+5]
		tmp += coefficients48To32[1][5] * In[3*i+6]
		tmp += coefficients48To32[1][6] * In[3*i+7]
		tmp += coefficients48To32[1][7] * In[3*i+8]
		Out[2*i+1] = tmp
	}

	for i := 0; i < len(Out); i++ {
		Out[i] = (Out[i] - (1 << 14)) >> 15
	}

	return Out
}
