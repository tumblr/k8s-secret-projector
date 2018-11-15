package types

import (
	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"k8s.io/api/core/v1"
)

// ProjectionMapping is the interface for a namespace's needs for secrets
// combining a set of declared dependent secrets, with the actual secrets
// on disk
type ProjectionMapping interface {
	GetEncryptionConfig() conf.Encryption
	GetNamespace() string
	GetName() string
	GetRepo() string
	String() string
	//Pluck out secret from json path and repo
	ProjectSecret(credsPath string) (*v1.Secret, error)
	ProjectSecretAsYAMLString(credsPath string) (string, error)
}
