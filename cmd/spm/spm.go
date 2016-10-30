package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytegust/spm"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "spm - Simple Process Manager"
	app.Usage = "spm [command] [argument]"

	wait := make(chan bool, 0)
	manager := spm.NewManager()

	go func() {
		wait <- <-manager.NotifyEnd
	}()

	app.Action = func(c *cli.Context) error {
		sock := spm.NewSocket()

		switch c.Args().First() {
		// start deamon
		case "":
			go func() {
				ch := make(chan os.Signal)
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
				<-ch
				wait <- true
			}()

			go func() {
				if err := sock.Listen(); err != nil {
					log.Fatal(err)
				}
			}()
			go func() {
				for conn := range sock.Connection {
					go func() {
						for m := range conn.Message {
							switch m.Command {
							case "start":
								go manager.StartAll(m.Jobs)
								conn.Close()
								return
							case "list":
								if err := conn.Send(spm.Message{
									JobList: manager.List(),
								}); err != nil {
									log.Println(err)
								}
								conn.Close()
								return
							case "stop":
								job := m.Arguments[0]
								if job == "" {
									manager.StopAll()
								} else {
									manager.Stop(job)
								}
								conn.Close()
								return
							}
						}
					}()
				}
			}()

			<-wait

			sock.Close()
			manager.StopAll()
		case "start":
			file, err := os.Open("Procfile")
			if err != nil {
				log.Fatal(err)
			}
			p := spm.NewParser(file)
			jobs, err := p.Parse()
			if err != nil {
				log.Fatal(err)
			}

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
			if err := sock.Dial(); err != nil {
				log.Fatal(err)
			}
			if err := sock.Send(spm.Message{
				Command: "list",
			}); err != nil {
				log.Fatal(err)
			}
			m := <-sock.Message

			log.Println("running jobs:")
			for _, job := range m.JobList {
				log.Println(job)
			}
		}

		return nil
	}

	app.Run(os.Args)
}
