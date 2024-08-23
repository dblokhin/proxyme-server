// Proxyme Developers. All rights reserved.
// License can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

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
	// options
	opts, err := getOpts()
	if err != nil {
		log.Fatal(err)
	}

	srv, err := proxyme.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		log.Println("shutdown proxyme")
		srv.Close()
		os.Exit(0)
	}()

	// start socks5 proxy
	addr := net.JoinHostPort(os.Getenv(envHost), getPort())
	log.Println("starting on", addr)

	if err := srv.ListenAndServe("tcp", addr); err != nil {
		log.Println(err)
	}
}

func getPort() string {
	const defaultPort = "1080"

	if p := os.Getenv(envPort); p != "" {
		return p
	}
	return defaultPort
}

func getOpts() (proxyme.Options, error) {
	// PROXY_BIND_IP=x.x.x.x
	// PROXY_NOAUTH=yes
	// PROXY_USERS=admin:admin,secret:pass
	// env PROXY_BIND_IP enables socks5 BIND operations

	// enable noauth authenticate method if given
	noauth := slices.Contains([]string{"yes", "true", "1"}, strings.ToLower(os.Getenv(envNoAuth)))

	// enable username/password authenticate method if given
	users, err := newUAM(os.Getenv(envUsers))
	if err != nil {
		return proxyme.Options{}, err
	}

	var authenticate func(username, password []byte) error
	if users.len() > 0 {
		authenticate = users.authenticate
	}

	// todo enable gssapi authenticate method if given

	opts := proxyme.Options{
		AllowNoAuth:  noauth,
		Authenticate: authenticate,
		GSSAPI:       nil,
		Connect:      customConnect,
		BindIP:       nil,
		MaxConnIdle:  0,
	}

	return opts, nil
}

// customConnect connects to remote server using dns resolver with lru cache
func customConnect(ctx context.Context, addressType int, addr []byte, port string) (io.ReadWriteCloser, error) {
	const domainType = 3

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
