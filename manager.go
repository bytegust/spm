package spm

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
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

	for _, sock := range job.WaitSockets {
		if err := AwaitReachable(sock.Type, sock.Addr, time.Minute); err != nil {
			log.Println(fmt.Sprintf("cannot start job '%s' because dependency timeout", job.Name))
			return
		}
	}

	c := exec.Command("sh", "-c", job.Command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	job.NotifyEnd = make(chan bool)
	job.Cmd = c

	m.mu.Lock()
	m.Jobs[job.Name] = job
	m.mu.Unlock()

	if err := c.Start(); err != nil {
		log.Println(err)

		m.mu.Lock()
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

func AwaitReachable(typ, addr string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := net.Dial(typ, addr)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("%v unreachable for %v", typ, addr, maxWait)
}
