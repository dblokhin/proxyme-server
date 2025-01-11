package main

import (
	"os"
	"testing"
)

func Test_getPort(t *testing.T) {
	tests := []struct {
		name    string
		envPort string
		want    string
	}{
		{
			name:    "empty",
			envPort: "",
			want:    "1080",
		},
		{
			name:    "common case",
			envPort: "9999",
			want:    "9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save old value (if any)
			oldValue, hadValue := os.LookupEnv(envPort)
			// Set or unset PORT env
			if tt.envPort == "" {
				os.Unsetenv(envPort)
			} else {
				os.Setenv(envPort, tt.envPort)
			}

			got := getPort()
			if got != tt.want {
				t.Fatalf("getPort() = %q, want %q", got, tt.want)
			}

			// Restore old value
			if hadValue {
				os.Setenv(envPort, oldValue)
			} else {
				os.Unsetenv(envPort)
			}
		})
	}
}

func Test_parseOptions(t *testing.T) {
	tests := []struct {
		name             string
		noAuthEnv        string
		usersEnv         string
		bindIPEnv        string
		wantAllowNoAuth  bool
		wantAuthenticate bool
		wantListen       bool
		wantErr          bool
	}{
		{
			name:             "no environments",
			noAuthEnv:        "",
			usersEnv:         "",
			bindIPEnv:        "",
			wantAllowNoAuth:  false,
			wantAuthenticate: false,
			wantListen:       false,
			wantErr:          false,
		},
		{
			name:             "noauth enabled",
			noAuthEnv:        "yes", // could also test "true" or "1"
			usersEnv:         "",
			bindIPEnv:        "",
			wantAllowNoAuth:  true,
			wantAuthenticate: false,
			wantListen:       false,
			wantErr:          false,
		},
		{
			name:             "valid users case",
			noAuthEnv:        "",
			usersEnv:         "bob:secret,alice:12345",
			bindIPEnv:        "",
			wantAllowNoAuth:  false,
			wantAuthenticate: true,
			wantListen:       false,
			wantErr:          false,
		},
		{
			name:             "invalid users format",
			noAuthEnv:        "",
			usersEnv:         "bob-secret", // missing ':'
			bindIPEnv:        "",
			wantAllowNoAuth:  false,
			wantAuthenticate: false,
			wantListen:       false,
			wantErr:          true,
		},
		{
			name:             "bind ip",
			noAuthEnv:        "",
			usersEnv:         "",
			bindIPEnv:        "127.0.0.1",
			wantAllowNoAuth:  false,
			wantAuthenticate: false,
			wantListen:       true,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save old env values
			oldNoAuth, hadNoAuth := os.LookupEnv(envNoAuth)
			oldUsers, hadUsers := os.LookupEnv(envUsers)
			oldBindIP, hadBindIP := os.LookupEnv(envBindIP)

			// Set or unset NO_AUTH
			if tt.noAuthEnv == "" {
				os.Unsetenv(envNoAuth)
			} else {
				os.Setenv(envNoAuth, tt.noAuthEnv)
			}
			// Set or unset USERS
			if tt.usersEnv == "" {
				os.Unsetenv(envUsers)
			} else {
				os.Setenv(envUsers, tt.usersEnv)
			}
			// Set or unset BIND_IP
			if tt.bindIPEnv == "" {
				os.Unsetenv(envBindIP)
			} else {
				os.Setenv(envBindIP, tt.bindIPEnv)
			}

			opts, err := parseOptions()

			if tt.wantErr && err == nil {
				t.Fatalf("parseOptions() error = nil; wantErr = true")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("parseOptions() error = %v; wantErr = false", err)
			}

			// Check AllowNoAuth
			if opts.AllowNoAuth != tt.wantAllowNoAuth {
				t.Errorf("AllowNoAuth = %v; want %v", opts.AllowNoAuth, tt.wantAllowNoAuth)
			}
			// Check Authenticate is set/unset
			gotAuthenticate := (opts.Authenticate != nil)
			if gotAuthenticate != tt.wantAuthenticate {
				t.Errorf("Authenticate != expected. Got %v, want %v", gotAuthenticate, tt.wantAuthenticate)
			}
			// Check Listen is set/unset
			gotListen := (opts.Listen != nil)
			if gotListen != tt.wantListen {
				t.Errorf("Listen != expected. Got %v, want %v", gotListen, tt.wantListen)
			}

			// Restore old env values
			if hadNoAuth {
				os.Setenv(envNoAuth, oldNoAuth)
			} else {
				os.Unsetenv(envNoAuth)
			}
			if hadUsers {
				os.Setenv(envUsers, oldUsers)
			} else {
				os.Unsetenv(envUsers)
			}
			if hadBindIP {
				os.Setenv(envBindIP, oldBindIP)
			} else {
				os.Unsetenv(envBindIP)
			}
		})
	}
}
