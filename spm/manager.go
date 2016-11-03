package spm

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

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
	fileName := fmt.Sprintf("/tmp/spm_%s.log", job.Name)
	logfile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatal(err)
	} else {
		job.Logfile = logfile
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

		job.Logfile.Close()
		go func() { m.jobEnded(job) }()
	}

	// generate random logging color
	rand.Seed(int64(time.Now().Nanosecond()))
	job.LogColor = rand.Intn(250) + 1

	// read command's stdout line by line

	in := bufio.NewScanner(pr)
	go func() {
		for in.Scan() {
			l := m.LoggerPrefix(job) + in.Text()
			// write to stdout (console)
			fmt.Fprintln(os.Stdout, l)
			// write to job specific logfile
			if _, err = logfile.WriteString(l + "\n"); err != nil {
				log.Fatal(err)
			}
		}
		if err := in.Err(); err != nil {
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
	log.Println(fmt.Sprintf("job `%s` ended", job.Name))
	job.NotifyEnd <- true
}

func (m *Manager) Stop(job string) {
	j := m.Jobs[job]
	defer j.Logfile.Close()
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
	file := m.Jobs[job].Logfile
	scanner := reverse.NewScanner(file)
	for i := 0; i < n && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// LoggerPrefix returns given logger values with unix color code and time as prefix
func (m *Manager) LoggerPrefix(job Job) string {
	t := time.Now().Format("15:04:05 PM")
	return fmt.Sprintf("\033[38;5;%dm%s %s | \033[0m", job.LogColor, t, job.Name)
}
