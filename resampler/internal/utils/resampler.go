package utils

func ResampleWithChannelAmt(resampler func([]int16) []int16, inp []int16, numCh int) []int16 {
    switch numCh {
    case 1:
        return resampler(inp)
    case 2:
        left := make([]int16, len(inp) / 2)
        right := make([]int16, len(inp) / 2)
        for i := 0; i < len(inp); i++ {
            left[i] = inp[i * 2] // TODO use shifts
            right[i] = inp[i * 2 + 1]
        }
        return append(resampler(left), resampler(right)...)
    default:
        res := make([]int16, 0, len(inp)) // questionable solution TODO check if makes worse
        cur := make([]int16, len(inp)/numCh)
        for i := 0; i < numCh; i++ {
            for j := 0; j < len(inp); j++ {
                cur[j] = inp[j * i]
            }
            res = append(res, resampler(cur)...)
        }
        return res
    }
}

type Resampable interface {
    Resample([]int16) error
}

type mtResampable interface {
    GetIthResampler(i int) Resampable
}

func ResampleWithChannelAmtTest(rsm mtResampable, inp []int16, numCh int) error {
    switch numCh {
    case 1:
        return rsm.GetIthResampler(0).Resample(inp)
    case 2:
        left := make([]int16, len(inp) / 2)
        right := make([]int16, len(inp) / 2)
        for i := 0; i * 2 + 1 < len(inp); i++ { // i * 2 is fine but, just to be sure
            left[i] = inp[i * 2] // TODO use shifts
            right[i] = inp[i * 2 + 1]
        }
        err1 := rsm.GetIthResampler(0).Resample(left)
        err2 := rsm.GetIthResampler(1).Resample(right)
        if err1 != nil {
            return err1
        }
        if err2 != nil {
            return err2
        }
    default:
        cur := make([]int16, len(inp)/numCh)
        for i := 0; i < numCh; i++ {
            for j := 0; j < len(inp); j++ {
                cur[j] = inp[j * i]
            }
            err := rsm.GetIthResampler(i).Resample(cur)
            if err != nil {
                return err
            }
        }
        return nil 
    }
    return nil
}
