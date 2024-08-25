package main

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/dblokhin/proxyme"
)

func getPort() string {
	const defaultPort = "1080"

	if p := os.Getenv(envPort); p != "" {
		return p
	}
	return defaultPort
}

func parseOptions() (proxyme.Options, error) {
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

	// enable BIND operation if given
	var customBind func() (net.Listener, error)
	if bind := os.Getenv(envBindIP); bind != "" {
		customBind = func() (net.Listener, error) {
			fmt.Println("bind is called!")
			return net.Listen("tcp", fmt.Sprintf("%s:0", bind))
		}
	}

	// todo enable gssapi authenticate method if given
	opts := proxyme.Options{
		AllowNoAuth:  noauth,
		Authenticate: authenticate,
		GSSAPI:       nil,
		Connect:      customConnect,
		Bind:         customBind,
	}

	return opts, nil
}
