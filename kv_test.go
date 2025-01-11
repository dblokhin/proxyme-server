package main

import (
	"testing"
)

func TestKeyValueDBAdd(t *testing.T) {
	tests := []struct {
		name           string
		initialKeys    map[string]string
		maxSize        int
		keyToAdd       string
		valueToAdd     string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:        "add key to empty DB",
			initialKeys: map[string]string{},
			maxSize:     2,
			keyToAdd:    "foo",
			valueToAdd:  "bar",
			expectErr:   false,
		},
		{
			name:        "add key that already exists",
			initialKeys: map[string]string{"foo": "bar"},
			maxSize:     2,
			keyToAdd:    "foo",
			valueToAdd:  "baz",
			expectErr:   true,
			// Exact wording from the code is `"%q is already exists"`, e.g. "\"foo\" is already exists"
			expectedErrMsg: "\"foo\" is already exists",
		},
		{
			name:           "add key when DB is at max size",
			initialKeys:    map[string]string{"foo": "bar", "abc": "xyz"},
			maxSize:        2,
			keyToAdd:       "new",
			valueToAdd:     "val",
			expectErr:      true,
			expectedErrMsg: "too much entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &keyValueDB{
				data:    tt.initialKeys,
				maxSize: tt.maxSize,
			}

			err := db.Add(tt.keyToAdd, tt.valueToAdd)
			if tt.expectErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("did not expect an error, got: %v", err)
			}
			if tt.expectErr && err != nil {
				if err.Error() != tt.expectedErrMsg {
					t.Fatalf("expected error message %q, got %q", tt.expectedErrMsg, err.Error())
				}
			}

			// If no error, verify the key was added
			if !tt.expectErr {
				got := db.Get(tt.keyToAdd)
				if got != tt.valueToAdd {
					t.Errorf("expected value %q, got %q", tt.valueToAdd, got)
				}
			}
		})
	}
}

func TestKeyValueDBGet(t *testing.T) {
	// Table of test cases
	tests := []struct {
		name        string
		initialKeys map[string]string
		keyToGet    string
		expectedVal string
	}{
		{
			name:        "get existing key",
			initialKeys: map[string]string{"foo": "bar"},
			keyToGet:    "foo",
			expectedVal: "bar",
		},
		{
			name:        "get non-existing key",
			initialKeys: map[string]string{"foo": "bar"},
			keyToGet:    "nonexist",
			expectedVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &keyValueDB{
				data: tt.initialKeys,
				// maxSize not needed for Get test
			}

			got := db.Get(tt.keyToGet)
			if got != tt.expectedVal {
				t.Errorf("expected value %q, got %q", tt.expectedVal, got)
			}
		})
	}
}

func TestKeyValueDBLen(t *testing.T) {
	tests := []struct {
		name        string
		initialKeys map[string]string
		expectedLen int
	}{
		{
			name:        "empty DB",
			initialKeys: map[string]string{},
			expectedLen: 0,
		},
		{
			name:        "db with multiple entries",
			initialKeys: map[string]string{"foo": "bar", "abc": "xyz"},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &keyValueDB{
				data: tt.initialKeys,
			}
			gotLen := db.Len()
			if gotLen != tt.expectedLen {
				t.Errorf("expected length %d, got %d", tt.expectedLen, gotLen)
			}
		})
	}
}
