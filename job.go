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

	// WaitSockets holds socket information to wait their network availability
	// before running this job.
	WaitSockets []WaitSocket

	NotifyEnd chan bool `json:"-"`
	Cmd       *exec.Cmd `json:"-"`
}

type WaitSocket struct {
	Type string // tcp or udp
	Addr string // ip:port
}

type Logger struct {
	Out *logstreamer.Logstreamer
	Err *logstreamer.Logstreamer
}
