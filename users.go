package main

import (
	"errors"
	"fmt"
	"strings"
)

const maxUsersTotal = 1024 // limit the number of pairs user/password

var errDenied = errors.New("invalid username or password")

// uam is user authentication mechanism structure, essentially in-memory db of users that provide authenticate method :)
type uam struct {
	users *keyValueDB
}

func (u uam) authenticate(username, password []byte) error {
	if len(username) == 0 || len(password) == 0 {
		return errDenied
	}

	if u.users.Get(string(username)) != string(password) {
		return errDenied
	}

	return nil
}

func (u uam) len() int {
	return u.users.Len()
}

// newUAM initialized structure that is able to authenticate by username/password.
// env is a string username/password pairs in follow format "user1:pass1,user2:pass2".
func newUAM(env string) (uam, error) {
	db := &keyValueDB{
		data:    make(map[string]string),
		maxSize: maxUsersTotal,
	}

	for _, entry := range strings.Split(env, ",") {
		if entry == "" {
			continue
		}

		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return uam{}, fmt.Errorf("invalid entry/pass string %q", entry)
		}

		if parts[0] == "" || parts[1] == "" {
			return uam{}, fmt.Errorf("entry/password must be a non empty string: %q", entry)
		}

		if err := db.Add(parts[0], parts[1]); err != nil {
			return uam{}, fmt.Errorf("user users: %w", err)
		}
	}

	return uam{users: db}, nil
}
