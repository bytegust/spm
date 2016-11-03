package spm

import (
	"os"
	"os/exec"
)

type Job struct {
	Name     string
	Command  string
	Logfile  *os.File
	LogColor int

	NotifyEnd chan bool `json:"-"`
	Cmd       *exec.Cmd `json:"-"`
}
