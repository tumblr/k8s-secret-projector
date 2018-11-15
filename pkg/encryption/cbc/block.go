package cbc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
	"io"

	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption/key"
)

// Crypter is the module responsible for performing CBC encryption/decryption for secrets
// It uses cipher.Block as its underlying implementation
type Crypter struct {
	c     conf.Encryption
	block cipher.Block
	gcm   cipher.AEAD
	k     *Key
}

// New creates a new Crypter encryption module
func New(c conf.Encryption, credsKeyReader io.Reader) (*Crypter, error) {
	m := Crypter{
		c: c,
	}

	// set defaults if omitted, so we pick appropriate hashers for the cipher
	if c.Params["cipher"] == "" {
		c.Params["cipher"] = "aes"
	}
	if c.Params["hash"] == "" {
		switch c.Params["cipher"] {
		case "aes":
			// aes needs maximally {16,24,32}byte key size
			c.Params["hash"] = "md5"
		}
	}

	var hasher hash.Hash
	switch c.Params["hash"] {
	case "":
		fallthrough
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	case "md5":
		hasher = md5.New()
	default:
		return nil, fmt.Errorf("unsupported hash %s", c.Params["hash"])
	}

	// read in the key, and deserialize it. We only care about the Password
	// field, cause we want to call the constructor with that and our hasher
	tmpKey := Key{}
	err := json.NewDecoder(credsKeyReader).Decode(&tmpKey)
	if err != nil {
		return nil, fmt.Errorf("unable to load key: %s", err.Error())
	}
	key := NewKey(c.Params["cipher"], hasher, tmpKey.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to create key: %s", err.Error())
	}
	m.k = key

	switch c.Params["cipher"] {
	case "":
		fallthrough
	case "aes":
		m.block, err = aes.NewCipher([]byte(m.k.paddedHashedPassword))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported cipher %s", c.Params["cipher"])
	}

	m.gcm, err = cipher.NewGCM(m.block)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Decrypt some bytes
func (c *Crypter) Decrypt(data []byte) (plaintext []byte, err error) {
	nonceSize := c.gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err = c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return plaintext, err
	}
	return plaintext, nil
}

// Encrypt some bytes
func (c *Crypter) Encrypt(data []byte) (ciphertext []byte, err error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return ciphertext, err
	}
	ciphertext = c.gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptionKeys returns the encryption keys
func (c *Crypter) DecryptionKeys() ([]key.Key, error) {
	return []key.Key{c.k}, nil
}
