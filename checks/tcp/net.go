package tcp

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

// Server defines the minimum contract our
// TCP and UDP server implementations must satisfy.
type Server interface {
	Run() error
	Close() error
}

// NewServer creates a new Server using given protocol
// and addr.
func NewServer(protocol, addr string) (Server, error) {
	switch strings.ToLower(protocol) {
	case "tcp":
		return &TCPServer{
			addr: addr,
		}, nil
	case "udp":
		return &UDPServer{
			addr: addr,
		}, nil
	}
	return nil, errors.New("Invalid protocol given")
}

// TCPServer holds the structure of our TCP
// implementation.
type TCPServer struct {
	addr   string
	server net.Listener
}

// Run starts the TCP Server.
func (t *TCPServer) Run() (err error) {
	t.server, err = net.Listen("tcp", t.addr)
	if err != nil {
		return err
	}
	defer t.Close()

	for {
		conn, err := t.server.Accept()
		if err != nil {
			err = errors.New("could not accept connection")
			break
		}
		if conn == nil {
			err = errors.New("could not create connection")
			break
		}
		return t.handleConnections()
	}
	return
}

// Close shuts down the TCP Server
func (t *TCPServer) Close() (err error) {
	return t.server.Close()
}

// handleConnections is used to accept connections on
// the TCPServer and handle each of them in separate
// goroutines.
func (t *TCPServer) handleConnections() (err error) {
	for {
		conn, err := t.server.Accept()
		if err != nil || conn == nil {
			err = errors.New("could not accept connection")
			break
		}

		go t.handleConnection(conn)
	}
	return
}

// handleConnections deals with the business logic of
// each connection and their requests.
func (t *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	for {
		req, err := rw.ReadString('\n')
		if err != nil {
			rw.WriteString("failed to read input")
			rw.Flush()
			return
		}

		rw.WriteString(fmt.Sprintf("Request received: %s", req))
		rw.Flush()
	}
}

// UDPServer holds the necessary structure for our
// UDP server.
type UDPServer struct {
	addr   string
	server *net.UDPConn
}

// Run starts the UDP server.
func (u *UDPServer) Run() (err error) {
	laddr, err := net.ResolveUDPAddr("udp", u.addr)
	if err != nil {
		return errors.New("could not resolve UDP addr")
	}

	u.server, err = net.ListenUDP("udp", laddr)
	if err != nil {
		return errors.New("could not listen on UDP")
	}

	return u.handleConnections()
}

func (u *UDPServer) handleConnections() error {
	var err error
	for {
		buf := make([]byte, 2048)
		n, conn, err := u.server.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			break
		}
		if conn == nil {
			continue
		}

		go u.handleConnection(conn, buf[:n])
	}
	return err
}

func (u *UDPServer) handleConnection(addr *net.UDPAddr, cmd []byte) {
	u.server.WriteToUDP([]byte(fmt.Sprintf("Request recieved: %s", cmd)), addr)
}

// Close ensures that the UDPServer is shut down gracefully.
func (u *UDPServer) Close() error {
	return u.server.Close()
}
