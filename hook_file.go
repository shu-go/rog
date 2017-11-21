package rog

import (
	"fmt"
	"io/ioutil"
	"os"
)

type fileHook struct {
	fileName string
	perm     os.FileMode
}

func FileHook(fileName string, perm os.FileMode) fileHook {
	return fileHook{fileName: fileName, perm: perm}
}

func (h fileHook) Run(v ...interface{}) bool {
	ioutil.WriteFile(h.fileName, []byte(fmt.Sprint(v...)), h.perm)
	return false
}
