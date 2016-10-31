package spm

import (
	"encoding/json"
	"net"
	"sync"
)

// unix socket for communicating between cli apps and running deamon.

var sockName = "/tmp/s4s4pm.sock"

type Socket struct {
	// Message emits imcoming messages from dialer or listener.
	Message chan Message

	// conn is a dialer when Socket is a dialer
	conn net.Conn

	// ln is a listener when Socket is a listener.
	ln net.Listener

	// Connection emits connected dialers when Socket is a listener.
	Connection chan *Socket

	mu sync.Mutex // protects following
	// Connections holds connected dialers when Socket is a listener.
	Connections []*Socket
}

type Message struct {
	// Command can be "empty", start, stop and etc.
	Command   string
	Arguments []string
	Jobs      []Job
	JobList   []string
}

func (s *Socket) Send(m Message) error {
	enc := json.NewEncoder(s.conn)
	return enc.Encode(m)
}

func NewSocket() *Socket {
	return &Socket{
		Message:    make(chan Message, 0),
		Connection: make(chan *Socket, 0),
	}
}

func (s *Socket) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}

	if s.ln != nil {
		if err := s.ln.Close(); err != nil {
			return err
		}
		s.mu.Lock()
		for _, sock := range s.Connections {
			sock.Close()
		}
		s.mu.Unlock()
	}

	return nil
}

func (s *Socket) readLoop(parent *Socket) {
	if parent != nil {
		parent.mu.Lock()
		parent.Connections = append(parent.Connections, s)
		parent.mu.Unlock()
	}

	dec := json.NewDecoder(s.conn)
	for {
		var mes Message
		if err := dec.Decode(&mes); err != nil {
			if parent != nil {
				parent.mu.Lock()
				for i, conn := range parent.Connections {
					if conn == s {
						close(conn.Message)
						parent.Connections = append(parent.Connections[:i], parent.Connections[i+1:]...)
						break
					}
				}
				parent.mu.Unlock()
			} else {
				close(s.Message)
				close(s.Connection)
			}
			return
		}
		s.Message <- mes
	}
}

func (s *Socket) Listen() error {
	ln, err := net.Listen("unix", sockName)
	if err != nil {
		return err
	}
	s.ln = ln

	for {
		c, err := ln.Accept()
		if err != nil {
			return nil
		}

		sock := NewSocket()
		sock.conn = c
		go sock.readLoop(s)

		s.Connection <- sock
	}
}

func (s *Socket) Dial() error {
	conn, err := net.Dial("unix", sockName)
	if err != nil {
		return err
	}
	s.conn = conn
	go s.readLoop(nil)
	return nil
}
