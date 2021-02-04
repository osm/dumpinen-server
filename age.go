package main

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

// encrypt encrypts the given string.
func (a *app) encrypt(s string) ([]byte, error) {
	out := &bytes.Buffer{}

	w, err := age.Encrypt(out, a.recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypted file: %v", err)
	}

	if _, err := io.WriteString(w, s); err != nil {
		return nil, fmt.Errorf("failed to write to encrypted file: %v", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encrypted file: %v", err)
	}

	return out.Bytes(), nil
}

// decrypt decrypts the given slice of bytes.
func (a *app) decrypt(b []byte) ([]byte, error) {
	r, err := age.Decrypt(bytes.NewReader(b), a.identity)
	if err != nil {
		return nil, fmt.Errorf("failed to open encrypted file: %v", err)
	}

	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		return nil, fmt.Errorf("failed to read encrypted file: %v", err)
	}

	return out.Bytes(), nil
}
