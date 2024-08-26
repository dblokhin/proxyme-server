// Proxyme Developers. All rights reserved.
// License can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dblokhin/proxyme"
)

const (
	envHost   = "PROXY_HOST"    // proxy host to listen to
	envPort   = "PROXY_PORT"    // port number, 1080 defaults
	envBindIP = "PROXY_BIND_IP" // ipv4/ipv6 address to make BIND socks5 operations
	envNoAuth = "PROXY_NOAUTH"  // yes, true, 1
	envUsers  = "PROXY_USERS"   // user:pass,user2:pass2
)

func main() {
	ctx, _ := signal.NotifyContext(context.TODO(), syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-ctx.Done()
		log.Println("shutdown proxyme")
	}()

	if err := runMain(ctx); err != nil {
		log.Fatal(err)
	}
}

// runMain returns error for os.Exit(1)
func runMain(ctx context.Context) error {
	opts, err := parseOptions()
	if err != nil {
		return fmt.Errorf("parse options: %w", err)
	}

	socks5, err := proxyme.New(opts)
	if err != nil {
		return fmt.Errorf("init socks5 protocol: %w", err)
	}

	srv := server{socks5}
	host := os.Getenv(envHost)
	port := getPort()
	addr := net.JoinHostPort(host, port)

	// start socks5 proxy
	log.Println("starting on", addr)
	if err := srv.ListenAndServe(ctx, addr); err != nil {
		log.Println(err)
	}

	return nil
}

// customConnect connects to remote server using dns resolver with lru cache
func customConnect(addressType int, addr []byte, port string) (net.Conn, error) {
	const (
		maxConnTime = 10 * time.Second
		domainType  = 3
	)

	ctx, cancel := context.WithTimeout(context.TODO(), maxConnTime)
	defer cancel()

	// get the ip addr
	ip := net.IP(addr)
	if addressType == domainType {
		dip, err := defaultDomainResolver(ctx, addr)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", proxyme.ErrHostUnreachable, err)
		}

		ip = dip
	}

	dialAddr := net.JoinHostPort(ip.String(), port)

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", dialAddr)
	if err != nil {
		if errors.Is(err, syscall.EHOSTUNREACH) {
			return conn, fmt.Errorf("%w: %v", proxyme.ErrHostUnreachable, err)
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			return conn, fmt.Errorf("%w: %v", proxyme.ErrConnectionRefused, err)
		}
		if errors.Is(err, syscall.ENETUNREACH) {
			return conn, fmt.Errorf("%w: %v", proxyme.ErrNetworkUnreachable, err)
		}
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return conn, fmt.Errorf("%w: %v", proxyme.ErrTTLExpired, err)
		}
		return conn, err
	}

	_ = conn.(*net.TCPConn).SetLinger(0) // nolint

	return conn, nil
}
