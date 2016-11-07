package spm

import (
	"os/exec"
)

type Job struct {
	Name    string
	Command string
	Logging *Logging

	NotifyEnd chan bool `json:"-"`
	Cmd       *exec.Cmd `json:"-"`
}
