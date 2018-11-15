package main // import github.com/tumblr/k8s-secret-projector

import (
	"C"
	"io"

	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption/cbc"
)

// New creates a new encryption module that we use to return an encryption.Module interface
func New(c conf.Encryption, credsKeyReader io.Reader, _ io.Reader) (encryption.Module, error) {
	return cbc.New(c, credsKeyReader)
}

func main() {}
