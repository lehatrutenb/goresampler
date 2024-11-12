package testutils

import (
    "time"
    "log"
)

type TestWave interface {
    New(seed int);
    GetIn(index int) (error, int32);
    GetOut(index int) (error, int32);
    InLen() int;
    OutLen() int;
};

// don't mind with interfaces in tests - don't believe that affect will change cmps
type TestResample interface {
    Resample([]int32) error;
    Get(index int) (error, int32);
};

type TestErr struct {
    ErrSqed float64
};

type TestResult struct {
    Te TestErr
    Resampeled []int32
    SDur time.Duration // summary duration
};

type TestObj struct {
    Tw TestWave
    Tr TestResample
    Tres TestResult
    RunAmt int // to measure time (in struct to later divide)
};

func (_ TestObj) New(tw TestWave, tr TestResample) TestObj {
    return TestObj {
        Tw: tw, Tr: tr,
    }
}

// will update testObj.Tres
func (tObj *TestObj) Run() error {
    var err error
    tObj.Tw.New(1)

    inWave := make([]int32, tObj.Tw.InLen())
    for i := 0; i < len(inWave); i++ {
        err, inWave[i] = tObj.Tw.GetIn(i)
        if err != nil {
            log.Println("failed to get input wave")
            return err
        }
    }

    sT := time.Now() // TODO check if there are another better opts to measure

    for runInd := 0; runInd < tObj.RunAmt; runInd++ {
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

    outWave := make([]int32, tObj.Tw.OutLen())
    var errSqed float64
    for i := 0; i < tObj.Tw.OutLen(); i++ {
        err1, got := tObj.Tr.Get(i)
        err2, corr := tObj.Tw.GetOut(i)
        if err1 != nil {
            log.Println("failed to get output wave")
            return err1
        }
        if err2 != nil {
            log.Println("failed to get correct output wave")
            return err2
        }
        errSqed += float64(got-corr)*float64(got-corr)
        outWave[i] = got
    }

    tObj.Tres.Te.ErrSqed = errSqed
    tObj.Tres.Resampeled = outWave
    tObj.Tres.SDur = sE.Sub(sT)

    return nil 
}
