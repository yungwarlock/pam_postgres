package dbmanager

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/jackc/pgx/pgproto3"
)

type loggingConn struct {
	net.Conn
	logger io.Writer
}

func (c *loggingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.handleManualDecode(b)
	return n, err
}

func (c *loggingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.handleManualDecode(b)
	return n, err
}

func (c *loggingConn) handleManualDecode(b []byte) {
	var t [1]byte
	var lenBuf [4]byte
	copy(t[:], b[:1])
	copy(lenBuf[:], b[1:5])

	msgLen := int(binary.BigEndian.Uint32(lenBuf[:]))
	if msgLen < 4 {
		log.Printf("invalid message length: %d", msgLen)
		return
	}

	bodyLen := msgLen - 4
	body := make([]byte, bodyLen)
	copy(body, b[5:5+bodyLen])

	msgType := t[0]

	switch msgType {
	case 'Q':
		var q pgproto3.Query
		if err := q.Decode(body); err != nil {
			fmt.Fprintf(c.logger, "decode Query: %v", err)
		} else {
			fmt.Fprintf(c.logger, "Decoded Query: %#v\n", q.String)
		}
	}
}
