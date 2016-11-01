package spm

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/kvz/logstreamer"
	"github.com/rogpeppe/rog-go/reverse"
)

type Manager struct {
	mu   sync.Mutex // protects following
	Jobs map[string]Job
}

func NewManager() *Manager {
	return &Manager{
		Jobs: make(map[string]Job),
	}
}

func (m *Manager) StartAll(jobs []Job) {
	for _, job := range jobs {
		m.start(job)
	}
}

func (m *Manager) start(job Job) {
	_, exists := m.Jobs[job.Name]
	if exists {
		log.Println(fmt.Sprintf("wont start job '%s' because already running", job.Name))
		return
	}

	// create logs file and assign to Job.LogFile
	logFile, err := os.Create(fmt.Sprintf("/tmp/spm_%s.log", job.Name))
	if err != nil {
		log.Fatal(err)
	} else {
		job.LogFile = logFile
	}

	// TODO(u) time color

	// create logger to prefix stream
	logger := log.New(io.MultiWriter(os.Stdout, logFile), "", log.Ltime)
	logStreamerOut := logstreamer.NewLogstreamer(logger, "| ", false)
	logStreamerErr := logstreamer.NewLogstreamer(logger, "| ", true)

	job.Logger = Logger{
		Out: logStreamerOut,
		Err: logStreamerErr,
	}

	c := exec.Command("sh", "-c", job.Command)
	c.Stderr = logStreamerErr
	c.Stdout = logStreamerOut
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	job.Logger.Err.FlushRecord()

	job.NotifyEnd = make(chan bool)
	job.Cmd = c

	m.mu.Lock()
	m.Jobs[job.Name] = job
	m.mu.Unlock()

	if err := c.Start(); err != nil {
		log.Println(err)

		m.mu.Lock()
		job.LogFile.Close()
		defer job.Logger.Out.Close()
		defer job.Logger.Err.Close()
		delete(m.Jobs, job.Name)
		m.mu.Unlock()

		go func() { job.NotifyEnd <- true }()
	}

	go func() {
		if err := c.Wait(); err != nil {
			log.Println(err)
		}

		m.mu.Lock()
		delete(m.Jobs, job.Name)
		m.mu.Unlock()

		job.NotifyEnd <- true
	}()
}

func (m *Manager) Stop(job string) {
	j := m.Jobs[job]
	defer j.LogFile.Close()
	defer j.Logger.Out.Close()
	defer j.Logger.Err.Close()
	pid, _ := syscall.Getpgid(j.Cmd.Process.Pid)
	syscall.Kill(-pid, syscall.SIGTERM)
	<-j.NotifyEnd
}

func (m *Manager) StopAll() {
	var wg sync.WaitGroup
	for job, _ := range m.Jobs {
		wg.Add(1)
		go func(job string) {
			m.Stop(job)
			wg.Done()
		}(job)
	}

	wg.Wait()
}

func (m *Manager) List() (jobs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for job, _ := range m.Jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// ReadLog reads last n lines of the file that corresponds to job.
func (m *Manager) ReadLog(job string, n int) (lines []string) {
	file := m.Jobs[job].LogFile
	scanner := reverse.NewScanner(file)
	for i := 0; i < n && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}
	return lines
}
