package spm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Manager struct {
	mu       sync.Mutex
	Commands map[string]*exec.Cmd

	NotifyEnd chan bool
	wg        sync.WaitGroup
}

func NewManager() *Manager {
	return &Manager{
		Commands:  make(map[string]*exec.Cmd),
		NotifyEnd: make(chan bool),
	}
}

func (m *Manager) Stop(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.destroy(job)
}

func (m *Manager) destroy(job string) {
	m.endProcess(m.Commands[job], syscall.SIGTERM)
	delete(m.Commands, job)
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for job, _ := range m.Commands {
		m.destroy(job)
	}
}

func (m *Manager) StartAll(jobs []Job) {
	for _, job := range jobs {
		m.wg.Add(1)
		go m.start(job)
	}

	m.wg.Wait()
	if len(m.Commands) == 0 {
		m.NotifyEnd <- true
	}
}

func (m *Manager) start(job Job) {
	_, exists := m.Commands[job.Name]
	if exists {
		log.Println(fmt.Sprintf("wont start job '%s' because already running", job.Name))
		return
	}
	for _, cmd := range job.Commands {
		c := exec.Command(cmd.Cmd[0], cmd.Cmd[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		c.Env = os.Environ()
		for key, val := range cmd.Env {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", key, val))
		}

		m.mu.Lock()
		m.Commands[job.Name] = c
		m.mu.Unlock()

		if err := c.Run(); err != nil {
			log.Println(err)
			m.mu.Lock()
			delete(m.Commands, job.Name)
			m.mu.Unlock()
			m.wg.Done()
		}
	}
}

func (m *Manager) endProcess(cmd *exec.Cmd, signal syscall.Signal) {
	pid, _ := syscall.Getpgid(cmd.Process.Pid)
	syscall.Kill(-pid, signal)
}

func (m *Manager) List() (jobs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for job, _ := range m.Commands {
		jobs = append(jobs, job)
	}
	return jobs
}
