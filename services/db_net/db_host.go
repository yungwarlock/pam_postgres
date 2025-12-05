package dbnet

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"crypto/rand"
	"math/big"

	"github.com/docker/docker/pkg/namesgenerator"
)

var hostname = "example.com"
var connections = make(map[string]net.Listener)

func CreateConnection(subdomain, port string) error {
	if _, exists := connections[subdomain]; exists {
		return fmt.Errorf("Connection already exists")
	}

	// start listening on the specified port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	connections[subdomain] = listener
	log.Printf("Listening on port %s for subdomain %s\n", port, subdomain)

	// var hasListener bool = false
	var closeChan = make(chan bool)

	go func() {
		<-closeChan
		listener.Close()
		delete(connections, subdomain)
		log.Printf("Closed connection for subdomain %s\n", subdomain)
	}()

	for {
		clientConn, _ := listener.Accept()

		domain := subdomain + "." + hostname
		clientHost, _, _ := net.SplitHostPort(clientConn.LocalAddr().String())
		// if hasListener {
		// 	log.Printf("Rejected connection from %s to %s: already has active connection\n", clientHost, domain)
		// 	clientConn.Close()
		// 	continue
		// }

		// hasListener = true
		log.Printf("New connection from %s to %s\n", clientHost, domain)
		go handleConnection(clientConn, 5*time.Minute, closeChan)
	}
}

type loggingConn struct {
	net.Conn
	r io.Writer
	w io.Writer
}

func (c *loggingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if n > 0 && c.r != nil {
		_, _ = c.r.Write(b[:n])
	}
	return n, err
}

func (c *loggingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if n > 0 && c.w != nil {
		_, _ = c.w.Write(b[:n])
	}
	return n, err
}

func handleConnection(client net.Conn, duration time.Duration, closeChan chan bool) {
	targetConn, err := net.Dial("tcp", "localhost:5432")
	if err != nil {
		client.Close()
		return
	}

	// Wrap targetConn so it's still a net.Conn but logs traffic to stdout
	targetConn = &loggingConn{Conn: targetConn, r: os.Stdout, w: os.Stdout}

	defer client.Close()

	// Set the "Kill Switch"
	timer := time.AfterFunc(duration, func() {
		log.Println("Time limit reached! Killing connection.")
		client.Close()
		targetConn.Close()
		closeChan <- true
	})
	defer timer.Stop()

	// Pipe data back and forth
	go io.Copy(targetConn, client)
	io.Copy(client, targetConn)
}

func generateCryptoRandInt(min, max int) (int, error) {
	// The range is inclusive, so max - min + 1 possible values
	rangeSize := big.NewInt(int64(max - min + 1))

	// rand.Int generates a value [0, rangeSize)
	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return 0, err
	}

	// Add the minimum value to the result
	return int(n.Int64()) + min, nil
}

func GenerateSubdomainAndPort() (string, string, string) {
	name := namesgenerator.GetRandomName(10)
	port, err := generateCryptoRandInt(20000, 40000)
	if err != nil {
		return "", "0", ""
	}

	fullName := name + "." + hostname

	return name, fmt.Sprintf("%d", port), fullName
}
