package main

import (
	"errors"
	"testing"
)

func TestNewUAM(t *testing.T) {
	tests := []struct {
		name          string
		env           string
		wantErr       bool
		expectedErr   string
		expectedUsers int
	}{
		{
			name:          "empty env string",
			env:           "",
			wantErr:       false,
			expectedUsers: 0,
		},
		{
			name:          "single valid user",
			env:           "user1:pass1",
			wantErr:       false,
			expectedUsers: 1,
		},
		{
			name:          "multiple valid users",
			env:           "user1:pass1,user2:pass2,user3:pass3",
			wantErr:       false,
			expectedUsers: 3,
		},
		{
			name:        "invalid format (missing colon)",
			env:         "user1pass1", // no ':'
			wantErr:     true,
			expectedErr: `invalid entry/pass string "user1pass1"`,
		},
		{
			name:        "empty user part",
			env:         ":pass1",
			wantErr:     true,
			expectedErr: `entry/password must be a non empty string ":pass1"`,
		},
		{
			name:        "empty password part",
			env:         "user1:",
			wantErr:     true,
			expectedErr: `entry/password must be a non empty string "user1:"`,
		},
		{
			name:        "duplicate user",
			env:         "user1:pass1,user1:anotherpass",
			wantErr:     true,
			expectedErr: `user users: "user1" is already exists`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := newUAM(tt.env)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("did not expect an error, got: %v", err)
			}

			// If no error, check the user count
			if !tt.wantErr {
				if got := u.len(); got != tt.expectedUsers {
					t.Errorf("expected %d users, got %d", tt.expectedUsers, got)
				}
			}
		})
	}
}

func TestUAM_Authenticate(t *testing.T) {
	// We will initialize a valid uam with known users,
	// then run table-driven tests on authenticate.
	env := "alice:1234,bob:secret,charlie:abc"
	u, err := newUAM(env)
	if err != nil {
		t.Fatalf("failed to init uam: %v", err)
	}

	tests := []struct {
		name        string
		username    []byte
		password    []byte
		expectedErr error
	}{
		{
			name:        "valid user alice",
			username:    []byte("alice"),
			password:    []byte("1234"),
			expectedErr: nil,
		},
		{
			name:        "valid user bob",
			username:    []byte("bob"),
			password:    []byte("secret"),
			expectedErr: nil,
		},
		{
			name:        "invalid password",
			username:    []byte("alice"),
			password:    []byte("wrongpass"),
			expectedErr: errDenied,
		},
		{
			name:        "non-existent user",
			username:    []byte("nobody"),
			password:    []byte("nopass"),
			expectedErr: errDenied,
		},
		{
			name:        "empty username",
			username:    []byte(""),
			password:    []byte("1234"),
			expectedErr: errDenied,
		},
		{
			name:        "empty password",
			username:    []byte("alice"),
			password:    []byte(""),
			expectedErr: errDenied,
		},
		{
			name:        "both user and password empty",
			username:    []byte(""),
			password:    []byte(""),
			expectedErr: errDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.authenticate(tt.username, tt.password)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestUAM_Len(t *testing.T) {
	env := "u1:p1,u2:p2,u3:p3"
	u, err := newUAM(env)
	if err != nil {
		t.Fatalf("failed to init uam: %v", err)
	}

	want := 3
	got := u.len()
	if got != want {
		t.Errorf("expected length %d, got %d", want, got)
	}
}
