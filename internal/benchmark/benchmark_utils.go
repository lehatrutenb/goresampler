package benchmark

var inRateConv = map[int]int{8000: 0, 11000: 1, 11025: 1, 16000: 2, 44000: 3, 44100: 3, 48000: 4}
var outRateConv = map[int]int{8000: 0, 16000: 1}

var rInRateConv = map[int][]int{0: {8000, 8000, 8000}, 1: {11000, 11025, 11025}, 2: {16000, 16000, 16000}, 3: {44000, 44100, 44100}, 4: {48000, 48000, 48000}}
var rOutRateConv = map[int]int{0: 8000, 1: 16000}
