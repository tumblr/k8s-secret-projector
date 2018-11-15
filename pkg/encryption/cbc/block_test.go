package cbc

import (
	//	"crypto/sha512"
	"io"
	"strings"
	"testing"

	"github.com/tumblr/k8s-secret-projector/pkg/conf"
)

var (
	someData1 = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	jsonKey1  = `{"password":"elloOliv3R!420"}`

	aesMd5Config = conf.Encryption{
		Module: "cbc",
		Params: map[string]string{
			"cipher": "aes",
			"hash":   "md5",
		},
	}
)

func newReaderFromString(s string) io.Reader {
	return strings.NewReader(s)
}

func TestBadCipher(t *testing.T) {
	expectedErrString := "unsupported cipher xxx420"
	_, err := New(conf.Encryption{
		Module: "cbc",
		Params: map[string]string{
			"cipher": "xxx420",
			"hash":   "md5",
		},
	}, newReaderFromString(jsonKey1))
	if err == nil || err.Error() != expectedErrString {
		t.Errorf("creating CBC module with cipher 'xxx420' should have failed with '%s' but got '%v'", expectedErrString, err)
		t.Fail()
	}
}

func TestAES(t *testing.T) {
	c, err := New(aesMd5Config, newReaderFromString(jsonKey1))
	if err != nil {
		t.Errorf("error creating new CBC encryption module: %s", err.Error())
		t.Fail()
	}

	keys, err := c.DecryptionKeys()
	if err != nil {
		t.Errorf("error reading DecryptionKeys: %s", err.Error())
		t.Fail()
	}
	if len(keys) != 1 {
		t.Errorf("expected 1 key, but got %d", len(keys))
	}
	expectedPlaintextPw := "elloOliv3R!420"
	if keys[0].Plaintext() != expectedPlaintextPw {
		t.Errorf("expected key password '%s' but got '%s'", expectedPlaintextPw, keys[0].Plaintext())
	}

	expectedHashedPw := "8980a9ab9b3e58458fca53bc7af70743"
	if c.k.hashedPassword != expectedHashedPw {
		t.Errorf("expected hashed key password '%s' but got '%s'", expectedHashedPw, c.k.hashedPassword)
	}
	if c.k.paddedHashedPassword != expectedHashedPw {
		t.Errorf("expected padded hashed key password '%s' but got '%s'", expectedHashedPw, c.k.paddedHashedPassword)
	}

	enc, err := c.Encrypt([]byte(someData1))
	if err != nil {
		t.Errorf("error encrypting data: %s", err.Error())
	}

	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Errorf("error decrypting data: %s", err.Error())
	}
	if string(dec) != someData1 {
		t.Errorf("encrypted data '%s' should have been '%s'", dec, someData1)
	}

}

func TestAESHashes(t *testing.T) {
	allowedHashes := []string{"md5", "sha1", "sha256", "sha512"}
	failHashes := []string{"fuck"}
	for _, h := range allowedHashes {
		cfg := aesMd5Config
		cfg.Params["hash"] = h
		_, err := New(cfg, newReaderFromString(jsonKey1))
		if err != nil {
			t.Errorf("error creating new CBC encryption module with hash %s: %s", h, err.Error())
			t.Fail()
		}
	}

	for _, h := range failHashes {
		cfg := aesMd5Config
		cfg.Params["hash"] = h
		_, err := New(cfg, newReaderFromString(jsonKey1))
		if err == nil {
			t.Errorf("creating new CBC encryption module with hash %s should have failed", h)
			t.Fail()
		}
	}

}
