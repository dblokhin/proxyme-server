package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/dblokhin/proxyme"
)

// todo: make it configurable?
var keepAliveConfig = net.KeepAliveConfig{
	Enable:   true,
	Idle:     20 * time.Second,
	Interval: 5 * time.Second,
	Count:    5,
}

type server struct {
	protocol *proxyme.SOCKS5
}

// ListenAndServe starts listening incoming connection for SOCKS5 clients.
// Use context for graceful shutdown.
func (s server) ListenAndServe(ctx context.Context, address string) error {
	lc := net.ListenConfig{}
	ls, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	go func() {
		// don't forget to close listener by ctx
		<-ctx.Done()
		_ = ls.Close()
	}()

	var wg sync.WaitGroup
	defer func() {
		log.Println("waiting all connections be closed")
		wg.Wait()
	}()

	for {
		conn, err := ls.Accept()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				time.Sleep(time.Second / 5) // nolint
				continue
			}

			return fmt.Errorf("accept: %w", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.serve(ctx, conn.(*net.TCPConn))
		}()
	}
}

func (s server) serve(ctx context.Context, tcpConn *net.TCPConn) {
	_ = tcpConn.SetLinger(0)
	_ = tcpConn.SetKeepAliveConfig(keepAliveConfig)

	// set up deadline for idle connections
	conn := tcpConnWithTimeout{
		TCPConn: tcpConn,
		timeout: time.Hour,
	}

	done := make(chan any)

	// run socks
	go func() {
		s.protocol.Handle(conn, func(err error) {
			log.Println(err)
		})

		close(done)
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}

	_ = conn.Close()
}

type tcpConnWithTimeout struct {
	*net.TCPConn
	timeout time.Duration
}

func (t tcpConnWithTimeout) ReadFrom(r io.Reader) (n int64, err error) {
	_ = t.TCPConn.SetDeadline(time.Now().Add(t.timeout)) // nolint
	return t.TCPConn.ReadFrom(r)
}

func (t tcpConnWithTimeout) WriteTo(w io.Writer) (n int64, err error) {
	_ = t.TCPConn.SetDeadline(time.Now().Add(t.timeout)) // nolint
	return t.TCPConn.WriteTo(w)
}

func (t tcpConnWithTimeout) Write(p []byte) (n int, err error) {
	_ = t.TCPConn.SetDeadline(time.Now().Add(t.timeout)) // nolint
	return t.TCPConn.Write(p)
}

func (t tcpConnWithTimeout) Read(p []byte) (n int, err error) {
	_ = t.TCPConn.SetDeadline(time.Now().Add(t.timeout)) // nolint
	return t.TCPConn.Read(p)
}
