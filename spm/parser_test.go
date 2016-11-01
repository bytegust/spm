package spm

import (
	"strings"
	"testing"
)

var procfile = `chord: wait for tcp 127.0.0.1:6379: \
cd $GOPATH/src/github.com/bytegust/chord && make dev
`

func TestParser(t *testing.T) {
	p := NewParser(strings.NewReader(procfile))
	jobs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	job, job1 := jobs[0], jobs[1]

	if job.Name != "a" {
		t.Error("wrong job name")
	}
	if job.Command != "ENV1=value1 program1 arg1 && program2" {
		t.Error("wrong command")
	}
	if job.WaitSockets[0].Type != "tcp" {
		t.Error("wrong protocol")
	}
	if job.WaitSockets[0].Addr != "localhost:6379" {
		t.Error("wrong address")
	}

	if job1.Name != "b" {
		t.Error("wrong job name")
	}
	if job1.Command != "program3 arg3 arg4" {
		t.Error("wrong command")
	}
	if job1.WaitSockets != nil {
		t.Error("wrong socket info")
	}
}
