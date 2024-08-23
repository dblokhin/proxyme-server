package main

import (
	"errors"
	"fmt"
	"sync"
)

// keyValueDB simple kv mem storage that doesn't allow key duplications.
// Also it is limited size by specifying maxSize.
type keyValueDB struct {
	data    map[string]string
	mu      sync.Mutex
	maxSize int
}

// Add adds k/v to db if k doesn't exist, otherwise throws error
func (d *keyValueDB) Add(key, val string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.data[key]; ok {
		return fmt.Errorf("%q is already exists", key)
	}

	if len(d.data) >= d.maxSize {
		return errors.New("too much entries")
	}

	d.data[key] = val
	return nil
}

// Get returns value by key, return empty string if key is not present
func (d *keyValueDB) Get(key string) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.data[key]
}

// Len returns the number of database entries
func (d *keyValueDB) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return len(d.data)
}
