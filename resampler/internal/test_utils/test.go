package testutils

import (
	"log"
	"math"
	"time"
)

type TestWave interface {
	New(seed int)
	GetIn(index int) (int16, error)
	GetOut(index int) (int16, error)
	InLen() int
	OutLen() int
	String() string
}

// don't mind with interfaces in tests - don't believe that affect will change cmps
type TestResampler interface {
	Resample([]int16) error
	Get(index int) (int16, error)
	String() string
}

type TestErr struct { // площадь процентного различия больтше 10 5
	SqProc20    float64 `json:"Proc of xs with error higher 20%"`
	SqProc10    float64 `json:"Proc of xs with error higher 10%"`
	SqProc5     float64 `json:"Proc of xs with error higher 5%"`
	SqProc1     float64 `json:"Proc of xs with error higher 1%"`
	ErrSqed     float64 `json:"ErrSqed (x-corr)^2"`
	ErrMeanSqed float64 `json:"ErrSqed ((x-corr)^2)/output_samples_amt"`
}

type TestResultZipped struct {
	Te   TestErr
	SDur time.Duration `json:"SDurMS"`
}

type TestResult struct {
	Te         TestErr
	Resampeled []int16
	InWave     []int16
	CorrectW   []int16
	SDur       time.Duration `json:"SDurMS"` // summary duration
}

type TestObj struct {
	Tw     TestWave
	Tr     TestResampler
	Tres   TestResult
	RunAmt int // to measure time (in struct to later divide) - but not write large values not to affect results
}

func (TestObj) New(tw TestWave, tr TestResampler, runAmt int) TestObj {
	return TestObj{
		Tw: tw, Tr: tr, Tres: TestResult{}, RunAmt: runAmt,
	}
}

func (tErr *TestErr) recalcErr(got, corr int16) {
	tErr.ErrSqed += float64(got-corr) * float64(got-corr)

	log.Println(math.Abs(float64(got-corr)), math.Abs(float64(corr))/100.0*5.0)
	if math.Abs(float64(got-corr)) > math.Abs(float64(corr))/100.0*1.0 {
		tErr.SqProc1++
	}
	if math.Abs(float64(got-corr)) > math.Abs(float64(corr))/100.0*5.0 {
		tErr.SqProc5++
	}
	if math.Abs(float64(got-corr)) > math.Abs(float64(corr))/100.0*10.0 {
		tErr.SqProc10++
	}
	if math.Abs(float64(got-corr)) > math.Abs(float64(corr))/100.0*20.0 {
		tErr.SqProc20++
	}
}

// will update testObj.Tres
func (tObj *TestObj) Run() error {
	var err error
	tObj.Tw.New(1)

	inWave := make([]int16, tObj.Tw.InLen())
	for i := 0; i < len(inWave); i++ {
		inWave[i], err = tObj.Tw.GetIn(i)
		if err != nil {
			log.Println("failed to get input wave")
			return err
		}
	}

	sT := time.Now() // TODO check if there are another better opts to measure

	for runInd := 0; runInd < tObj.RunAmt; runInd++ { // run as much as said
		err = tObj.Tr.Resample(inWave)
		if err != nil {
			log.Println("failed to resample")
			return err
		}

		// if sm realization is lazy/...
		for i := 0; i < tObj.Tw.OutLen(); i++ {
			_, _ = tObj.Tr.Get(i)
		}
	}

	sE := time.Now()

	outWave := make([]int16, tObj.Tw.OutLen())
	CorrectW := make([]int16, tObj.Tw.OutLen())
	for i := 0; i < tObj.Tw.OutLen(); i++ { // cmp results
		got, err1 := tObj.Tr.Get(i)
		corr, err2 := tObj.Tw.GetOut(i)
		if err1 != nil {
			log.Println("failed to get output wave")
			return err1
		}
		if err2 != nil {
			log.Println("failed to get correct output wave")
			return err2
		}

		tObj.Tres.Te.recalcErr(got, corr)
		outWave[i] = got
		CorrectW[i] = corr
	}

	tObj.Tres.Te.ErrMeanSqed = tObj.Tres.Te.ErrSqed / float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc1 /= float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc5 /= float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc10 /= float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc20 /= float64(tObj.Tw.OutLen())

	tObj.Tres.Resampeled = outWave
	tObj.Tres.InWave = inWave
	tObj.Tres.CorrectW = CorrectW
	tObj.Tres.SDur = time.Duration((sE.Sub(sT) / time.Duration(tObj.RunAmt)).Milliseconds()) // divide, no?

	return nil
}
