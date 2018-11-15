package cbc

import (
	"encoding/hex"
	"fmt"
	"hash"
)

var (
	cipherKeyByteLength = map[string]int{
		"aes": 32,
	}
)

func padOrTrim(d []byte, cipherName string) (padded []byte) {
	size, ok := cipherKeyByteLength[cipherName]
	if !ok {
		// unsupported cipher %s, unsure what padding to use for this hashed key
		// just leave it as is.
		return d
	}
	l := len(d)
	if l == size {
		return d
	}
	if l > size {
		return d[l-size:]
	}
	tmp := make([]byte, size)
	copy(tmp[size-l:], d)
	return tmp
}

// Key stores the key used to decrypt a CBC encrypted jawn
type Key struct {
	Password             string `json:"password",yaml:"password"`
	hasher               hash.Hash
	hashedPassword       string
	paddedHashedPassword string
}

// NewKey takes a hasher and a password, and return a Key struct
func NewKey(cipherName string, hasher hash.Hash, password string) *Key {
	k := Key{
		Password: password,
		hasher:   hasher,
	}
	hasher.Write([]byte(k.Password))
	k.hashedPassword = hex.EncodeToString(k.hasher.Sum(nil))
	k.paddedHashedPassword = string(padOrTrim([]byte(k.hashedPassword), cipherName))
	return &k
}

// Plaintext returns the password for this key
func (k *Key) Plaintext() string {
	return k.Password
}

// String returns a string representation
func (k *Key) String() string {
	return fmt.Sprintf("{k:%s}", k.Password)
}
