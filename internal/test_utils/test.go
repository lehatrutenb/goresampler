package testutils

import (
	"math"
	"resampler/internal/utils"
	"testing"
	"time"
)

type TestWave interface {
	Seed(seed int)
	GetIn(index int) (int16, error)
	GetOut(index int) (int16, error)
	InLen() int
	OutLen() int
	String() string
	InRate() int
	OutRate() int
	WithResampled() bool
	NumChannels() int
}

// don't mind with interfaces in tests - don't believe that affect will change cmps
type TestResampler interface {
	Copy() TestResampler
	Resample(inWave []int16) error
	Get(index int) (int16, error)
	OutLen() int
	OutRate() int
	String() string
}

type MTTestResampler struct {
	trs []TestResampler
}

type TestErr struct { // площадь процентного различия больше 10 5
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
	Te          TestErr
	Resampeled  []int16
	InWave      []int16
	CorrectW    []int16
	SDur        time.Duration `json:"MSDurMS"` // mean summary duration
	NumChannels int           // to be able to parse logs later
	InRate      int
	OutRate     int
}

type TestOpts struct {
	ToCrSF        bool
	OutPlotPath   string
	failOnHighErr bool
	CreateReports bool
	SFName        string
}

type TestObj struct {
	Tw     TestWave
	Tr     MTTestResampler
	Tres   TestResult
	RunAmt int // to measure time (in struct to later divide) - but not write large values not to affect results
	t      *testing.T
	opts   TestOpts
}

func (MTTestResampler) Seed(int) {
}

func (MTTestResampler) New(tr TestResampler, chAmt int) MTTestResampler {
	trs := make([]TestResampler, chAmt)
	for i := 0; i < len(trs); i++ {
		trs[i] = tr.Copy()
	}
	return MTTestResampler{trs}
}

func (tr MTTestResampler) Copy() TestResampler {
	res := new(MTTestResampler)
	*res = tr.New(tr.trs[0], len(tr.trs))
	return res
}

func (tr MTTestResampler) GetIthResampler(i int) utils.Resampable {
	return tr.trs[i]
}

func (tr *MTTestResampler) Resample(in []int16) error {
	return utils.ResampleWithChannelAmtTest(tr, in, len(tr.trs))
}

func (tr MTTestResampler) Get(ind int) (int16, error) {
	return tr.trs[ind%len(tr.trs)].Get(ind / len(tr.trs))
}

func (tr MTTestResampler) OutLen() int {
	return tr.trs[0].OutLen() * len(tr.trs)
}

func (tr MTTestResampler) OutRate() int {
	return tr.trs[0].OutRate()
}

func (tr MTTestResampler) String() string {
	return tr.trs[0].String()
}

func (TestOpts) NewDefault() *TestOpts {
	res := new(TestOpts)
	res.OutPlotPath = SAVE_PATH
	res.ToCrSF = false
	res.failOnHighErr = true
	res.CreateReports = true
	res.SFName = ""
	return res
}

func (TestOpts) New(toCrSoundF bool, outPlotPath string) *TestOpts {
	res := new(TestOpts)
	res.OutPlotPath = SAVE_PATH
	res.ToCrSF = false
	res.failOnHighErr = true
	res.CreateReports = true
	res.SFName = ""
	return res
}

func (to *TestOpts) WithPlotPath(outPlotPath string) *TestOpts {
	to.OutPlotPath = outPlotPath
	return to
}

func (to *TestOpts) WithCrSF(toCrSF bool) *TestOpts {
	to.ToCrSF = toCrSF
	return to
}

func (to *TestOpts) NotFailOnHighErr() *TestOpts {
	to.failOnHighErr = false
	return to
}

func (to *TestOpts) SetCreateReports(toCreate bool) *TestOpts {
	to.CreateReports = toCreate
	return to
}

func (to *TestOpts) SetSFName(name string) *TestOpts {
	to.SFName = name
	return to
}

// New - expected nil as default opts value
func (TestObj) New(tw TestWave, tr TestResampler, runAmt int, t *testing.T, opts *TestOpts) TestObj {
	if opts == nil {
		opts = &TestOpts{ToCrSF: false, OutPlotPath: SAVE_PATH}
	}
	return TestObj{
		Tw: tw, Tr: MTTestResampler{}.New(tr, tw.NumChannels()), Tres: TestResult{}, RunAmt: runAmt, t: t, opts: *opts,
	}
}

func (tErr *TestErr) recalcErr(got, corr int16) {
	tErr.ErrSqed += float64(got-corr) * float64(got-corr)

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
	tObj.Tw.Seed(1)

	inWave := make([]int16, tObj.Tw.InLen())
	for i := 0; i < len(inWave); i++ {
		inWave[i], err = tObj.Tw.GetIn(i)
		if err != nil {
			tObj.t.Error("")
			tObj.t.Error("failed to get input wave")
			return err
		}
	}

	sT := time.Now() // TODO check if there are another better opts to measure

	for runInd := 0; runInd < tObj.RunAmt; runInd++ { // run as much as said
		err = utils.ResampleWithChannelAmtTest(tObj.Tr, inWave, tObj.Tw.NumChannels())
		if err != nil {
			tObj.t.Error("failed to resample")
			return err
		}

		// if sm realization is lazy/...
		for i := 0; i < tObj.Tr.OutLen(); i++ {
			_, _ = tObj.Tr.Get(i)
		}
	}

	sE := time.Now()

	outWave := make([]int16, tObj.Tr.OutLen())
	CorrectW := make([]int16, tObj.Tr.OutLen())

	for i := 0; i < tObj.Tr.OutLen(); i++ { // cmp results
		got, err1 := tObj.Tr.Get(i)
		if err1 != nil {
			tObj.t.Error("failed to get output wave")
			return err1
		}
		outWave[i] = got

		if tObj.Tw.WithResampled() && i < tObj.Tw.OutLen() { // care
			corr, err2 := tObj.Tw.GetOut(i)
			if err2 != nil {
				tObj.t.Errorf("failed to get correct output wave ; ind: %d ; waveLen: %d ; resampled waveLen: %d", i, tObj.Tw.OutLen(), tObj.Tr.OutLen())
				return err2
			}

			tObj.Tres.Te.recalcErr(got, corr)
			CorrectW[i] = corr
		}
	}

	tObj.Tres.Resampeled = outWave
	tObj.Tres.InWave = inWave
	tObj.Tres.SDur = time.Duration((sE.Sub(sT) / time.Duration(tObj.RunAmt)).Milliseconds()) // divide, no?
	tObj.Tres.InRate = tObj.Tw.InRate()
	tObj.Tres.OutRate = tObj.Tr.OutRate()
	tObj.Tres.NumChannels = tObj.Tw.NumChannels()

	if !tObj.Tw.WithResampled() {
		return nil
	}

	tObj.Tres.Te.ErrMeanSqed = tObj.Tres.Te.ErrSqed / float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc1 /= float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc5 /= float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc10 /= float64(tObj.Tw.OutLen())
	tObj.Tres.Te.SqProc20 /= float64(tObj.Tw.OutLen())

	// small check for resampler correctness
	if tObj.opts.failOnHighErr && tObj.Tres.Te.SqProc20 >= 0.2 { // if too large error too often
		tObj.t.Logf("too large error too often: tObj.Tres.Te.SqProc20=%f", tObj.Tres.Te.SqProc20)
		tObj.t.Fail()
	}

	tCorr := float64(tObj.Tw.OutLen()) / float64(tObj.Tw.OutRate())
	tGot := float64(len(tObj.Tres.Resampeled)) / float64(tObj.Tres.OutRate)
	if math.Abs(tCorr-tGot) >= 1e-5 { // if got just math error in time of resampled
		tObj.t.Logf("too large time difference: correct:%f got:%f", tCorr, tGot)
		tObj.t.Fail()
	}

	tObj.Tres.CorrectW = CorrectW

	return nil
}
