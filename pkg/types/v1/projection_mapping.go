package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption"
	"github.com/tumblr/k8s-secret-projector/pkg/types"
	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/printers"
)

var (
	// ErrEncryptionRequestedButNoEncryptionConfigSpecified if user asks to encrypt a data item, but
	// didnt specify an encryption config for the projection manifest, we dont know _how_ the user wants
	// this stuff encrypted
	ErrEncryptionRequestedButNoEncryptionConfigSpecified = fmt.Errorf("encryption of a data element was requested, but no encryption_config was found to instantiate an encryptio      n module")
	// DecryptionKeysPrefix is the prefix used for injecting the decryption keys JSON into a Secret when Encryption.IncludeDecryptionKeys is true
	DecryptionKeysPrefix = "keys_"
)

// ProjectionMapping is a v1 implementation of the struct that joins
// a declaration of dependency on a set of secrets, with the secrets
// sourced from a secret repository
type ProjectionMapping struct {
	Name       string          `json:"name",yaml:"name"`
	Namespace  string          `json:"namespace",yaml:"namespace"`
	Repo       string          `json:"repo",yaml:"repo"`
	Data       []Secret        `json:"data",yaml:"data"` //data:
	Encryption conf.Encryption `yaml:"encryption",json:"encryption"`

	crypter encryption.Module
	c       conf.Config
}

// LoadFromYamlBytes parses a ProjectionMapping from a string
func LoadFromYamlBytes(raw []byte, cfg conf.Config) (types.ProjectionMapping, error) {
	/**
		var aux struct {
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
			Data      []struct {
				Name   string `yaml:"name"`
				Source struct {
					JSON     string `yaml:"json"`
					YAML     string `yaml:"yaml"`
					Raw      string `yaml:"raw"`
					JSONPath string `yaml:"jsonpath"`
				} `yaml:"source"`
			} `yaml:"data"`
		}
	  **/
	var m ProjectionMapping
	m.c = cfg
	err := yaml.UnmarshalStrict(raw, &m)
	if err != nil {
		return nil, err
	}

	// setup the crypter. if no module requested, skip setting this up (we will bail if any items asked to be
	// encrypted but didnt specify the module)
	if m.Encryption.Module != "" {
		if m.Encryption.CredsKeysFilePath == "" {
			// override with the value from cfg
			m.Encryption.CredsKeysFilePath = cfg.CredsEncryptionKeyFile()
		}
		if m.Encryption.KeysDecrypterFilePath == "" {
			// override with the value from cfg
			m.Encryption.KeysDecrypterFilePath = cfg.CredsKeyDecryptionKeyFile()
		}
		c, err := encryption.NewModuleFromEncryptionConfig(m.Encryption)
		if err != nil {
			return nil, err
		}
		m.crypter = c
	}
	return &m, nil
}

// ProjectSecretAsYAMLString will take a ProjectionMapping and return the k8s secret resource
// as a YAML representation in string form
func (m *ProjectionMapping) ProjectSecretAsYAMLString(credsPath string) (string, error) {
	sec, err := m.ProjectSecret(credsPath)
	if err != nil {
		return "", err
	}
	// kubernetes has some magic to YAMLify objects, so lets use that
	// cause we cant blindly use yaml.Marshal without proper annotations
	// on the struct fields
	p := printers.YAMLPrinter{}
	buf := bytes.NewBuffer([]byte{})
	err = p.PrintObj(sec, buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ProjectSecret will take a ProjectionMapping and return the k8s secret resource
func (m *ProjectionMapping) ProjectSecret(credsPath string) (*v1.Secret, error) {
	data := map[string][]byte{}
	// the k8s v1.Secret is a combination of all its Secret's datasources
	// so project each one, into the v1.Secret

	for _, s := range m.Data {
		d, err := s.Project(credsPath)
		if err != nil {
			return nil, err
		}
		if s.Encrypt && m.crypter == nil {
			return nil, ErrEncryptionRequestedButNoEncryptionConfigSpecified
		}
		if s.Encrypt {
			ed, err := m.crypter.Encrypt(d)
			if err != nil {
				return nil, err
			}
			data[s.Name] = ed
		} else {
			data[s.Name] = d
		}
	}
	// include decryption keys if requested in the generated Secret
	if m.crypter != nil && m.Encryption.IncludeDecryptionKeys {
		keys, err := m.crypter.DecryptionKeys()
		if err != nil {
			return nil, err
		}
		for i, k := range keys {
			js, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			data[fmt.Sprintf("%s%d.json", DecryptionKeysPrefix, i+1)] = js
		}
	}
	sekrit := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: v1.SecretTypeOpaque,
		Data: data,
	}
	if m.c.AddDeployLabels() {
		sekrit.ObjectMeta.Labels = map[string]string{
			m.c.LabelVersionKey(): m.c.Generation(),
			m.c.LabelManagedKey(): "true",
		}
	}
	return &sekrit, nil
}

func (m *ProjectionMapping) String() string {
	data := make([]string, len(m.Data))
	for i, s := range m.Data {
		data[i] = s.String()
	}
	return fmt.Sprintf("%s/%s:%s{%s}", m.Namespace, m.Name, m.Repo, strings.Join(data, ","))
}

// GetEncryptionConfig is the encryption configuration for this projection
func (m *ProjectionMapping) GetEncryptionConfig() conf.Encryption {
	return m.Encryption
}

// GetNamespace is the namespace the ProjectionMapping is bound to
func (m *ProjectionMapping) GetNamespace() string {
	return m.Namespace
}

// GetName is the name of a ProjectionMapping
func (m *ProjectionMapping) GetName() string {
	return m.Name
}

// GetRepo returns the creds source repository for this ProjectionMapping
func (m *ProjectionMapping) GetRepo() string {
	return m.Repo
}
