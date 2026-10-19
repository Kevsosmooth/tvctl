// Package p2p links tvctl to a TV over a direct WebRTC connection. The two ends
// find each other by posting sealed notes on public Nostr relays, so no server of
// our own is needed.
package p2p

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Code is the pairing code shared by a TV and its helper. Everything posted on
// the public relays is derived from it, so it must never be posted itself.
type Code [16]byte

var codeEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewCode returns a random pairing code.
func NewCode() (Code, error) {
	var c Code
	_, err := rand.Read(c[:])
	return c, err
}

// ParseCode reads a code written by String. Case, spaces and dashes are ignored.
func ParseCode(s string) (Code, error) {
	var c Code
	clean := strings.ToUpper(strings.NewReplacer("-", "", " ", "", "\n", "").Replace(s))
	b, err := codeEncoding.DecodeString(clean)
	if err != nil || len(b) != len(c) {
		return c, fmt.Errorf("pairing code is not valid")
	}
	copy(c[:], b)
	return c, nil
}

// ReadCode loads a code saved by LoadOrCreateCode.
func ReadCode(path string) (Code, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Code{}, err
	}
	return ParseCode(string(b))
}

// LoadOrCreateCode reads the code at path, or saves a new one there readable
// only by its owner.
func LoadOrCreateCode(path string) (Code, error) {
	c, err := ReadCode(path)
	if !errors.Is(err, fs.ErrNotExist) {
		return c, err
	}
	if c, err = NewCode(); err != nil {
		return c, err
	}
	return c, os.WriteFile(path, []byte(c.String()+"\n"), 0o600)
}

func (c Code) String() string {
	return strings.ToLower(codeEncoding.EncodeToString(c[:]))
}

// topic is the public label both ends watch. It reveals nothing about the code.
func (c Code) topic() string {
	sum := sha256.Sum256(append([]byte("tvremote/topic/"), c[:]...))
	return hex.EncodeToString(sum[:16])
}

func (c Code) aead() (cipher.AEAD, error) {
	key, err := hkdf.Key(sha256.New, c[:], nil, "tvremote/seal", 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
