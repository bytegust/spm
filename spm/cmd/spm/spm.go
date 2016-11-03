package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bytegust/tools/spm"
	"github.com/urfave/cli"
)

var procfile string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := cli.NewApp()
	app.Name = "spm - Simple Process Manager"
	app.Usage = "spm [options] [command] [argument]"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       "./",
			Usage:       "procfile location (e.g. ./spm/cmd/Procfile or ./spm/cmd/)",
			Destination: &procfile,
		},
	}

	app.Action = func(c *cli.Context) error {
		command := c.Args().First()
		handleCliCommand(c, command)
		return nil
	}

	app.Run(os.Args)
}

func handleCliCommand(c *cli.Context, command string) {
	switch command {
	// start daemon
	case "":
		startDaemon(c)
	case "start":
		procfile := getProcfilePath(procfile)
		file, err := os.Open(procfile)
		if err != nil {
			log.Fatal(err)
		}

		p := spm.NewParser(file)
		jobs, err := p.Parse()
		if err != nil {
			log.Fatal(err)
		}

		sock := spm.NewSocket()
		if err := sock.Dial(); err != nil {
			log.Fatal(err)
		}

		if err := sock.Send(spm.Message{
			Command: "start",
			Jobs:    jobs,
		}); err != nil {
			log.Fatal(err)
		}

		<-sock.Message
		log.Println("done")
	case "stop":
		sock := spm.NewSocket()
		if err := sock.Dial(); err != nil {
			log.Fatal(err)
		}

		if err := sock.Send(spm.Message{
			Command:   "stop",
			Arguments: []string{c.Args().Get(1)},
		}); err != nil {
			log.Fatal(err)
		}

		<-sock.Message
		log.Println("done")
	case "list":
		sock := spm.NewSocket()
		if err := sock.Dial(); err != nil {
			log.Fatal(err)
		}

		if err := sock.Send(spm.Message{
			Command: "list",
		}); err != nil {
			log.Fatal(err)
		}

		m := <-sock.Message
		fmt.Println("Running jobs:")
		for _, job := range m.JobList {
			fmt.Printf("\t%s", job)
		}
		fmt.Println("") // line break
	case "logs":
		sock := spm.NewSocket()
		if err := sock.Dial(); err != nil {
			log.Fatal(err)
		}

		if err := sock.Send(spm.Message{
			Command:   "logs",
			Arguments: []string{c.Args().Get(1)},
		}); err != nil {
			log.Fatal(err)
		}

		m := <-sock.Message
		for i := len(m.JobLogs) - 1; i >= 0; i-- {
			fmt.Println(m.JobLogs[i])
		}
	}
}

func startDaemon(c *cli.Context) {
	quit := make(chan bool)
	manager := spm.NewManager()
	sock := spm.NewSocket()

	// listen for user termination
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		quit <- true
	}()

	// start listening for cli apps
	go func() {
		if err := sock.Listen(); err != nil {
			log.Fatal(err)
		}
	}()

	// handle incoming cli app connections
	go func() {
		for conn := range sock.Connection {
			go handleMessage(<-conn.Message, conn, manager, quit)
		}
	}()

	log.Println("deamon started")

	<-quit
	sock.Close()
	manager.StopAll()

	log.Println("deamon ended")
}

func handleMessage(mes spm.Message, conn *spm.Socket, manager *spm.Manager, quit chan bool) {
	switch mes.Command {
	case "start":
		go manager.StartAll(mes.Jobs)
		conn.Close()
	case "list":
		if err := conn.Send(spm.Message{
			JobList: manager.List(),
		}); err != nil {
			log.Println(err)
		}
		conn.Close()
	case "stop":
		job := mes.Arguments[0]
		if job == "" {
			manager.StopAll()
		} else {
			manager.Stop(job)
		}
		conn.Close()
	case "logs":
		job := mes.Arguments[0]
		if job == "" {
			// @Todo suggest help command after #4
			conn.Close()
		}
		if err := conn.Send(spm.Message{
			JobLogs: manager.ReadLog(job, 200),
		}); err != nil {
			log.Println(err)
		}
		conn.Close()
	}
}

func getProcfilePath(input string) string {
	re := regexp.MustCompile("(/)$|(/Procfile(\\s+?|$))")
	match := re.FindStringSubmatch(input)

	if len(match) > 0 {
		if match[1] == "/" {
			return input + "Procfile"
		} else if match[2] == "/Procfile" {
			return input
		}
	}

	return input + "/Procfile"
}
