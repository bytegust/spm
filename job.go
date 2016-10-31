package spm

import "os/exec"

type Job struct {
	Name    string
	Command string

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
