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

var template = `Simple Process Manager.

Usage: spm [OPTIONS] COMMAND [arg...]
       spm [ --help ]

Options:

  -f, --file=~/.Procfile    Location of Procfile

  -h, --help                Print usage

Commands:

    start           Start all jobs defined in Procfile

    start <jobs>    Starts jobs if present in Procfile

    stop            Stop all current jobs

    stop <jobs>     Stop jobs if currently running

    list            Lists all running jobs

    logs <job>      Prints last 200 lines of job's logfile

Documentation can be found at https://github.com/bytegust/tools
`

var procfile string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cli.AppHelpTemplate = template

	app := cli.NewApp()
	app.Name = "spm - Simple Process Manager"
	app.Usage = "spm [OPTIONS] COMMAND [args...]"

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

		var j []spm.Job
		if args := c.Args()[1:]; len(args) > 0 {
			for _, arg := range args {
				exist := false
				for _, job := range jobs {
					if job.Name == arg {
						j = append(j, job)
						exist = true
						break
					}
				}
				if !exist {
					fmt.Printf("job %s is not exist in procfile\n", arg)
				}
			}
		} else {
			j = jobs
		}

		if err := sock.Send(spm.Message{
			Command: "start",
			Jobs:    j,
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
			Arguments: c.Args()[1:],
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
			fmt.Printf("\t%s\n", job)
		}
		fmt.Println("") // line break
	case "logs":
		sock := spm.NewSocket()
		if err := sock.Dial(); err != nil {
			log.Fatal(err)
		}

		if job := c.Args().Get(1); job == "" {
			fmt.Println(template)
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
	default:
		fmt.Println(template)
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
		if args := mes.Arguments; len(args) > 0 {
			for _, arg := range args {
				go manager.Stop(arg)
			}
		} else {
			manager.StopAll()
		}
		conn.Close()
	case "logs":
		job := mes.Arguments[0]
		if job == "" {
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
