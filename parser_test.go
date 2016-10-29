package spm

import (
	"strings"
	"testing"
)

var procfile = `a: ENV1=value1 program1 arg1 && program2
b: program3 arg3 arg4
`

func TestA(t *testing.T) {
	p := NewParser(strings.NewReader(procfile))
	jobs, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	job := jobs[0]
	job1 := jobs[1]

	if job.Name != "a" {
		t.Error("wrong job name")
	}
	if job.Commands[0].Env["ENV1"] != "value1" {
		t.Error("wrong env var value")
	}
	if job.Commands[0].Cmd[0] != "program1" {
		t.Error("wrong cmd")
	}
	if job.Commands[0].Cmd[1] != "arg1" {
		t.Error("wrong cmd")
	}
	if job.Commands[1].Cmd[0] != "program2" {
		t.Error("wrong cmd")
	}

	if job1.Name != "b" {
		t.Error("wrong job name")
	}
	if job1.Commands[0].Cmd[0] != "program3" {
		t.Error("wrong cmd")
	}
	if job1.Commands[0].Cmd[1] != "arg3" {
		t.Error("wrong cmd")
	}
	if job1.Commands[0].Cmd[2] != "arg4" {
		t.Error("wrong cmd")
	}
}
