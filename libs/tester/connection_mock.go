package tester

import (
	"log"
	"net"
	"time"
)

type ConnectionMock struct {
}

func (o ConnectionMock) Read(b []byte) (n int, err error) {
	return
}

func (o ConnectionMock) Write(b []byte) (n int, err error) {
	log.Printf("ConnectionMock: Write [%s]", string(b))
	return len(b), nil
}

func (o ConnectionMock) Close() error {
	log.Print("ConnectionMock: Close")
	return nil
}

func (o ConnectionMock) LocalAddr() net.Addr {
	return nil
}

func (o ConnectionMock) RemoteAddr() net.Addr {
	return nil
}

func (o ConnectionMock) SetDeadline(t time.Time) error {
	return nil
}

func (o ConnectionMock) SetReadDeadline(t time.Time) error {
	return nil
}

func (o ConnectionMock) SetWriteDeadline(t time.Time) error {
	return nil
}
