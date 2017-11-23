package rog

import (
	"fmt"
	"os/exec"
)

type annoyingHook struct {
}

func AnnoyingHook() annoyingHook {
	return annoyingHook{}
}

func (h annoyingHook) Run(v ...interface{}) bool {
	args := make([]string, len(v)+1)
	args[0] = "*"
	for i, a := range v {
		args[i+1] = fmt.Sprint(a)
	}
	cmd := exec.Command("msg", args...)
	cmd.Start()
	return false
}
