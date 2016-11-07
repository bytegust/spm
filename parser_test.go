package spm

import (
	"strings"
	"testing"
)

var procfile = `
# task: echo "comment line"
chord: make dev
# start redis
redis: redis-server
`

func TestParser(t *testing.T) {
	p := NewParser(strings.NewReader(procfile))
	jobs, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	job, job1 := jobs[0], jobs[1]

	if job.Name != "chord" {
		t.Error("wrong job name")
	}

	if job.Command != "make dev" {
		t.Error("wrong command")
	}

	if job1.Name != "redis" {
		t.Error("wrong job name")
	}
	if job1.Command != "redis-server" {
		t.Error("wrong command")
	}
}
