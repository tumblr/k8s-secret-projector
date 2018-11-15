package encryption

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"plugin"

	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption/cbc"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption/key"
)

var (
	// ErrMissingPluginPath is the error returned when a projection manifest specifies a "module: plugin" but no PluginPath
	ErrMissingPluginPath = fmt.Errorf("plugin-path is required for encryption module 'plugin'")
)

// Module allows a projection to encrypt its data elements. This allows implementation to
// be abstracted from use.
type Module interface {
	// Encrypt takes some []byte and returns them encrypted
	Encrypt([]byte) ([]byte, error)
	// Keys returns a list of key.Key used to decrypt the result of Encrypt()
	DecryptionKeys() ([]key.Key, error)
	// Decrypt takes some []byte and returns them unencrypted
	Decrypt([]byte) ([]byte, error)
}

// NewModuleFromEncryptionConfig returns a new encryption module
// it handles opening up any files referenced in the configs and passing io.Readers to
// the referenced encryption module's setup
func NewModuleFromEncryptionConfig(c conf.Encryption) (Module, error) {
	// clean up the conf.Encryption, in case it was just unmarshalled and lacks some structs
	if c.Params == nil {
		c.Params = map[string]string{}
	}
	fCredsKeysFile, err := os.Open(c.CredsKeysFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open CredsKeysFilePath %s: %s", c.CredsKeysFilePath, err.Error())
	}
	defer fCredsKeysFile.Close()

	var fKeysDecrypterReader io.Reader
	if c.KeysDecrypterFilePath != "" {
		fKeysDecrypterFile, err := os.Open(c.KeysDecrypterFilePath)
		if err != nil {
			return nil, err
		}
		defer fKeysDecrypterFile.Close()
		fKeysDecrypterReader = fKeysDecrypterFile
	} else {
		// allow an empty decrypter string, cause some modules will use symmetric encryption/decryption and only need 1 key
		fKeysDecrypterReader = ioutil.NopCloser(bytes.NewReader(nil))
	}

	switch c.Module {
	case "plugin":
		if c.PluginPath == "" {
			return nil, ErrMissingPluginPath
		}
		p, err := plugin.Open(c.PluginPath)
		if err != nil {
			return nil, err
		}
		// lookup the New function:
		new, err := p.Lookup("New")
		if err != nil {
			return nil, err
		}
		return new.(func(conf.Encryption, io.Reader, io.Reader) (Module, error))(c, fCredsKeysFile, fKeysDecrypterReader)
	case "cbc":
		return cbc.New(c, fCredsKeysFile)
	}
	return nil, fmt.Errorf("unsupported encryption module '%s'", c.Module)
}
