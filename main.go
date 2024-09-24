package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/Ayobami6/redis_clone/client"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Message struct {
	data []byte
	peer *Peer
}

type Server struct {
	Config    // embed confi struct, inheritance like stuff
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message

	kv *KeyVal
}

// sever constructor

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKeyVal(),
	}

}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		slog.Error("listen err", "err", err)
		return err
	}
	s.ln = ln
	// loop for waiting peer connection
	go s.loop()

	slog.Info("server running", "listenAddr", s.ListenAddr)
	// accepts tcp connection
	return s.acceptLoop()

}

// loop for waiting peer connection
func (s *Server) loop() {
	// start a for select loop to accept ready channel
	for {
		select {
		case rawMsg := <-s.msgCh:

			if err := s.handleRawMsg(rawMsg); err != nil {
				slog.Error("handle raw msg error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true

		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(con net.Conn) {
	peer := NewPeer(con, s.msgCh)
	s.addPeerCh <- peer
	// handle the connection from the peer
	// slog.Info("new peer connected", "remoteAddr", con.RemoteAddr())

	// start a loop for communication
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddr", con.RemoteAddr())
	}

}

func (s *Server) handleRawMsg(msg Message) error {
	cmd, err := parseCommand(string(msg.data))
	if err != nil {
		return err
	}
	switch c := cmd.(type) {
	case SetCommand:
		return s.kv.Set(c.key, c.val)
		// slog.Info("trying to set key into the hashtable", "key", c.key, "val", c.val)
	case GetCommand:
		val, ok := s.kv.Get(c.key)
		if !ok {
			return fmt.Errorf("key not found")

		}
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}

	}
	return nil

}

func main() {
	server := NewServer(Config{ListenAddr: ":5001"})
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second)
	client := client.NewClient("localhost:5001")

	// select {}
	for i := 0; i < 10; i++ {

		if err := client.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second)
		val, err := client.Get(context.TODO(), fmt.Sprintf("foo_%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("got this val back => ", val)
	}
	// select {}

	// test if set works
	fmt.Println(server.kv.data)
}
