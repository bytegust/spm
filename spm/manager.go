package spm

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

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
		m.Start(job)
	}
}

func (m *Manager) Start(job Job) {
	_, exists := m.Jobs[job.Name]
	if exists {
		log.Println(fmt.Sprintf("wont start job '%s' because already running", job.Name))
		return
	}

	logging, err := NewLogging(job.Name)
	if err != nil {
		log.Fatal(err)
	} else {
		job.Logging = logging
	}

	c := exec.Command("sh", "-c", job.Command)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	pr, pw, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	c.Stderr = pw
	c.Stdout = pw

	job.NotifyEnd = make(chan bool)
	job.Cmd = c

	m.mu.Lock()
	m.Jobs[job.Name] = job
	m.mu.Unlock()

	log.Println(fmt.Sprintf("job `%s` has been started", job.Name))
	if err := c.Start(); err != nil {
		log.Println(err)
		go m.jobEnded(job)
		return
	}

	// read command's stdout line by line
	in := bufio.NewScanner(pr)
	go func() {
		if err := job.Logging.Output(in); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := c.Wait(); err != nil {
			log.Println(err)
		}
		m.jobEnded(job)
	}()
}

func (m *Manager) jobEnded(job Job) {
	m.mu.Lock()
	delete(m.Jobs, job.Name)
	m.mu.Unlock()
	job.Logging.Close()
	log.Println(fmt.Sprintf("job `%s` ended", job.Name))
	job.NotifyEnd <- true
}

func (m *Manager) Stop(job string) {
	m.mu.Lock()
	j, exists := m.Jobs[job]
	m.mu.Unlock()
	if !exists {
		return
	}
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
	_, exists := m.Jobs[job]
	if !exists {
		lines = append(lines, "job "+job+" is not running")
		return
	}

	file := m.Jobs[job].Logging.Logfile
	scanner := reverse.NewScanner(file)
	for i := 0; i < n && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}

	return lines
}
