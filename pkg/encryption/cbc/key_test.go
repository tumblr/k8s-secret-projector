package cbc

import (
	"crypto/sha512"
	"testing"
)

func TestHashingPassword(t *testing.T) {
	pw := "thisIsMyPassw0rD6969!"
	expectedHash := "f1ef94d25559d95d19ad0ddcac70f23eadc0387d568f42f224d22d87ee1341048325696ec3c36286ab5e9d3fea3fb8e49029382183687d597ca5f2f942b0784d"
	expectedPaddedHash := "9029382183687d597ca5f2f942b0784d"
	k := NewKey("aes", sha512.New(), pw)
	if k.hashedPassword != expectedHash {
		t.Errorf("expected hashed password of '%s' to be '%s', but got '%s'", pw, expectedHash, k.hashedPassword)
		t.Fail()
	}
	if k.paddedHashedPassword != expectedPaddedHash {
		t.Errorf("expected hashed password of '%s' to be '%s', but got '%s'", pw, expectedPaddedHash, k.paddedHashedPassword)
		t.Fail()
	}
}
