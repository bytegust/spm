package spm

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

type Parser struct {
	r io.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{r: r}
}

type Job struct {
	Name     string
	Commands []Command
}

type Command struct {
	Env map[string]string
	Cmd []string
}

func (p *Parser) Parse() (jobs []Job, err error) {
	reader := bufio.NewReader(p.r)

PARSE:
	for {
		job := Job{}
		var lines []byte

		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				return jobs, nil
			}
			if err != nil {
				return jobs, err
			}

			// trim leading and trailing spaces
			line = bytes.TrimSpace(line)

			if len(line) == 0 {
				continue PARSE
			} else if len(line) > 0 && line[len(line)-1] == '\\' {
				lines = append(lines, line[:len(line)-1]...)
				continue
			} else {
				lines = append(lines, line...)
			}

			sp := strings.SplitN(string(lines), ":", 2)
			if len(sp) < 2 {
				return jobs, errors.New("spm: missing command")
			}

			job.Name = sp[0]
			commandsStr := sp[1]

			if job.Name == "" {
				return jobs, errors.New("spm: invalid name")
			}

			commandsSlice := strings.Split(commandsStr, " && ")
			for _, c := range commandsSlice {
				cmd := Command{Env: make(map[string]string)}
				exprs := strings.Split(c, " ")

				for _, expr := range exprs {
					expr = strings.Trim(expr, " ")
					if expr == "" {
						continue
					}

					// if it's an env var
					sl := strings.Split(expr, "=")
					if len(sl) == 2 {
						cmd.Env[sl[0]] = sl[1]
						continue
					}

					cmd.Cmd = append(cmd.Cmd, expr)
				}

				job.Commands = append(job.Commands, cmd)
			}

			jobs = append(jobs, job)
			break
		}
	}
}
