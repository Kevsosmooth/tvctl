package p2p

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// noteMaxAge rejects replayed notes. Relays may also hand back old notes on subscribe.
const noteMaxAge = 2 * time.Minute

const (
	kindOffer  = "offer"
	kindAnswer = "answer"
)

// note is the sealed message the two ends swap to set up one session.
type note struct {
	Kind    string `json:"k"`
	Session string `json:"s"`
	SDP     string `json:"d"`
	Sent    int64  `json:"t"`
}

func seal(a cipher.AEAD, n note) (string, error) {
	n.Sent = time.Now().Unix()
	plain, err := json.Marshal(n)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, a.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(a.Seal(nonce, nonce, plain, nil)), nil
}

// open returns an error for notes sealed with another code, garbled or stale.
func open(a cipher.AEAD, content string, now time.Time) (note, error) {
	var n note
	raw, err := base64.StdEncoding.DecodeString(content)
	if err != nil || len(raw) < a.NonceSize() {
		return n, errors.New("not a sealed note")
	}
	plain, err := a.Open(nil, raw[:a.NonceSize()], raw[a.NonceSize():], nil)
	if err != nil {
		return n, errors.New("note sealed with another code")
	}
	if err := json.Unmarshal(plain, &n); err != nil {
		return n, err
	}
	age := now.Sub(time.Unix(n.Sent, 0))
	if age > noteMaxAge || age < -noteMaxAge {
		return n, errors.New("stale note")
	}
	return n, nil
}
