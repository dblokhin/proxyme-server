package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/dblokhin/proxyme"
)

type server struct {
	protocol *proxyme.SOCKS5
}

// ListenAndServe starts listening incoming connection for SOCKS5 clients.
// Use context for graceful shutdown.
func (s server) ListenAndServe(ctx context.Context, address string) error {
	lc := net.ListenConfig{}
	ls, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return err
	}

	go func() {
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

			s.protocol.Handle(conn.(*net.TCPConn), func(err error) {
				log.Println(err)
			})
		}()
	}
}
