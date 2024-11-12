package testutils

import (
    "log"
    "os"
    "encoding/json"
    "fmt"
)

const SAVE_PATH="../../../plots"

func (tObj *TestObj) Save(dirName string) error {
    tMarsh := TestMarshalledResult{Te: tObj.Tres.Te, SDur: tObj.Tres.SDur}
    buf, err := json.Marshal(tMarsh)
    if err != nil {
        log.Println("failed to marshall results")
        return err
    }

    err = os.WriteFile(fmt.Sprintf("%s/%s/%s:%s", SAVE_PATH, dirName, tObj.Tr, tObj.Tw), buf, 0666)
	if err != nil {
		log.Println("failde to save metrices file")
        return err
	}

    // TODO Draw plots here and save more data

	return nil
}
