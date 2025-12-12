package dbnet

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	dbPort  string
	dbHost  string
	appHost string
)

func init() {
	dbPort = os.Getenv("DB_PORT")
	dbHost = os.Getenv("DB_HOST")
	appHost = os.Getenv("APP_HOSTNAME")

	fmt.Println(dbHost, dbPort, appHost)
}

type ConnectionManager struct {
	listener net.Listener
	mutex    sync.RWMutex
	ctx      context.Context
	clients  map[net.Conn]bool
	cancel   context.CancelFunc
}

var connections = make(map[string]*ConnectionManager)

func CreateConnection(subdomain, port string, timeout time.Duration) error {
	if _, exists := connections[subdomain]; exists {
		return fmt.Errorf("Connection already exists")
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	manager := &ConnectionManager{
		ctx:      ctx,
		cancel:   cancel,
		listener: listener,
		clients:  make(map[net.Conn]bool),
	}

	connections[subdomain] = manager
	log.Printf("Listening on port %s for subdomain %s\n", port, subdomain)

	go func() {
		<-ctx.Done()
		log.Printf("Time limit reached! Killing connection for subdomain %s\n", subdomain)

		manager.mutex.Lock()
		for client := range manager.clients {
			client.Close()
		}
		manager.mutex.Unlock()

		listener.Close()
		delete(connections, subdomain)
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Stopping listener for subdomain %s: %v\n", subdomain, err)
			return nil
		}

		select {
		case <-ctx.Done():
			clientConn.Close()
			return nil
		default:
		}

		domain := subdomain + "." + appHost
		clientHost, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())
		log.Printf("New connection from %s to %s\n", clientHost, domain)

		manager.mutex.Lock()
		manager.clients[clientConn] = true
		manager.mutex.Unlock()

		go manager.handleConnection(clientConn, subdomain)
	}
}

func (cm *ConnectionManager) handleConnection(client net.Conn, subdomain string) {
	defer func() {
		cm.mutex.Lock()
		delete(cm.clients, client)
		cm.mutex.Unlock()
		client.Close()
	}()

	targetConn, err := net.Dial("tcp", dbHost+":"+dbPort)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer targetConn.Close()

	targetConn = &loggingConn{
		logger: os.Stdout,
		Conn:   targetConn,
	}

	ctx, cancel := context.WithCancel(cm.ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		log.Println("Closing connections due to timeout")
		client.Close()
		targetConn.Close()
	}()

	done := make(chan error, 2)

	go func() {
		_, err := io.Copy(targetConn, client)
		done <- err
	}()

	go func() {
		_, err := io.Copy(client, targetConn)
		done <- err
	}()

	select {
	case <-done:
		log.Printf("Connection closed for subdomain %s", subdomain)
	case <-ctx.Done():
		log.Printf("Connection terminated due to timeout for subdomain %s", subdomain)
	}
}
