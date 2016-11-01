package spm

import (
	"os"
	"os/exec"

	"github.com/kvz/logstreamer"
)

type Job struct {
	Name    string
	Command string
	Logger  Logger
	LogFile *os.File

	NotifyEnd chan bool `json:"-"`
	Cmd       *exec.Cmd `json:"-"`
}

type Logger struct {
	Out *logstreamer.Logstreamer
	Err *logstreamer.Logstreamer
}
