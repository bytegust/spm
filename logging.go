package spm

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mattn/go-isatty"
)

type Logging struct {
	Prefix   string
	LogColor int
	Logfile  *os.File
}

func NewLogging(name string) (*Logging, error) {
	fileName := fmt.Sprintf("/tmp/spm_%s.log", name)
	// create logfile with given name
	logfile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return nil, err
	}

	// create random log color code
	code := genColorCode()

	prefix := loggerPrefix(code, name)

	return &Logging{
		Prefix:   prefix,
		LogColor: code,
		Logfile:  logfile,
	}, nil
}

// Write writes given string into Logfile
func (l *Logging) Write(s string) error {
	if _, err := l.Logfile.WriteString(s + "\n"); err != nil {
		return err
	}
	return nil
}

// Output reads the in then writes into both stdout and logfile
func (l *Logging) Output(in *bufio.Scanner) error {
	for in.Scan() {
		log := l.Prefix + in.Text()
		fmt.Fprintln(os.Stdout, log)
		l.Write(log)
	}

	if err := in.Err(); err != nil {
		return err
	}

	return nil
}

func (l *Logging) Close() error {
	if err := l.Logfile.Close(); err != nil {
		return err
	}

	return nil
}

// LoggerPrefix wraps given string and time with unix color code, as prefix
func loggerPrefix(code int, s string) string {
	t := time.Now().Format("15:04:05 PM")
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Sprintf("\033[38;5;%dm%s %s | \033[0m", code, t, s)
	}
	return fmt.Sprintf("%s %s | ", t, s)
}

func genColorCode() (code int) {
	rand.Seed(int64(time.Now().Nanosecond()))
	code = rand.Intn(231) + 1
	return
}
