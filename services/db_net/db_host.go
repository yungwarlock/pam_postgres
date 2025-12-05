package dbnet

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	dbPort  = os.Getenv("DB_PORT")
	dbHost  = os.Getenv("DB_HOSTNAME")
	appHost = os.Getenv("APP_HOSTNAME")
)

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

	var closeChan = make(chan bool)

	go func() {
		<-closeChan
		listener.Close()
		delete(connections, subdomain)
		log.Printf("Closed connection for subdomain %s\n", subdomain)
	}()

	for {
		clientConn, _ := listener.Accept()

		domain := subdomain + "." + appHost
		clientHost, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())
		log.Printf("New connection from %s to %s\n", clientHost, domain)
		go handleConnection(clientConn, 5*time.Minute, closeChan)
	}
}

func handleConnection(client net.Conn, duration time.Duration, closeChan chan bool) {
	targetConn, err := net.Dial("tcp", dbHost+":"+dbPort)
	if err != nil {
		client.Close()
		return
	}

	targetConn = &loggingConn{
		logger: os.Stdout,
		Conn:   targetConn,
	}

	defer client.Close()

	timer := time.AfterFunc(duration, func() {
		log.Println("Time limit reached! Killing connection.")
		client.Close()
		targetConn.Close()
		closeChan <- true
	})
	defer timer.Stop()

	go io.Copy(targetConn, client)
	io.Copy(client, targetConn)
}
