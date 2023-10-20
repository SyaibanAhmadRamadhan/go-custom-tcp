package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	From    string
	Payload []byte
}

type Server struct {
	ListenAddr string
	Ln         net.Listener
	Quitch     chan struct{}
	Message    chan Message
	Conn       map[string]net.Conn
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Quitch:     make(chan struct{}),
		Message:    make(chan Message, 10),
		Conn:       make(map[string]net.Conn),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.Ln = ln

	go s.acceptLoop()
	go s.writeMessage()

	<-s.Quitch
	close(s.Message)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.Ln.Accept()
		if err != nil {
			log.Printf("accept error : %v", err)
			continue
		}

		s.Conn[conn.RemoteAddr().String()] = conn
		log.Println("new connection to the server : ", conn.RemoteAddr())
		go s.readLoop(conn)

	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("failed close conncetion error : %v", err)
		}
	}()

	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("read error : %v", err)
			continue
		}

		message := Message{
			From:    conn.RemoteAddr().String(),
			Payload: buf[:n],
		}
		s.Message <- message

	}
}

func (s *Server) writeMessage() {
	for {
		msg := <-s.Message
		for i, conn := range s.Conn {
			var formatMsg string
			if i == msg.From {
				formatMsg = fmt.Sprintf("me: %s\nmessage: %s", msg.From, msg.Payload)
			} else {
				formatMsg = fmt.Sprintf("from: %s\nmessage: %s", msg.From, msg.Payload)
			}

			_, err := conn.Write([]byte(formatMsg))
			if err != nil {
				fmt.Printf("write message error : %v", err)
			}

		}
	}
}

func main() {
	server := NewServer(":3000")

	log.Fatal(server.Start())
}
