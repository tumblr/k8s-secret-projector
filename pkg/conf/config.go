package conf

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/tumblr/k8s-secret-projector/internal/pkg/version"
)

type resourceType int

const (
	directory resourceType = iota
	file
)

// config is the config loaded for a running instance; flags are stuffed in here!
type config struct {
	showSecrets     bool
	debug           bool
	addDeployLabels bool
	credsRootPaths  map[string]string
	// mappingsRootPath is the root where projection mappings are loaded from
	mappingsRootPath string
	outputDir        string
	// credsEncryptionKeys path to credential encryption key
	credsEncryptionKeyFile string
	// credsKeyEncryptionKeys path to credential keys encryption key
	credsKeyDecryptionKeyFile string

	// Label all generated ConfigMaps with this key, using the value of --generation
	labelVersionKey string
	// Label all generated ConfigMaps with this key=true
	labelManagedKey string
	// Generation label used when annotating secrets
	labelSecretGeneration string
}

// Config is the interface for loading flag settings for the CLI app
type Config interface {
	CredsRootPaths() map[string]string
	CredsRootPath(string) (string, error)
	CredsEncryptionKeyFile() string
	CredsKeyDecryptionKeyFile() string
	ProjectionMappingsRootPath() string
	OutputDir() string
	Debug() bool
	ShowSecrets() bool
	Version() string
	BuildDate() string
	Generation() string
	LabelVersionKey() string
	LabelManagedKey() string
	AddDeployLabels() bool
}

// LoadConfigFromArgs returns a new config given some CLI args
func LoadConfigFromArgs(args []string) (Config, error) {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	c := config{}
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: (version=%s commit=%s branch=%s runtime=%s built=%s)\n", args[0], version.Version, version.Commit, version.Branch, runtime.Version(), version.BuildDate)
		fs.PrintDefaults()
	}
	credsRepoFlags := NewMapStringStringFlag()

	fs.BoolVar(&c.showSecrets, "debug-show-secrets", true, "Show generated secrets YAML contents (only if -debug)")
	fs.BoolVar(&c.debug, "debug", false, "Debug")
	fs.StringVar(&c.outputDir, "output", "", "Output generated secrets here")

	fs.Var(&credsRepoFlags, "creds-repo", "label=<path> pair identifying a source credentials repository (i.e. production=/path/to/repo/production) (required)")

	fs.StringVar(&c.credsEncryptionKeyFile, "creds-encryption-key", "", "path to load creds_keys.json from creds_internal (optional, depends on your encryption modules in use)")
	fs.StringVar(&c.credsKeyDecryptionKeyFile, "creds-key-decryption-key", "", "path to load decryption keys from (optional, depends on your encryption modules in use)")
	fs.StringVar(&c.mappingsRootPath, "manifests", "", "Path to projection mapping yamls (required)")
	fs.BoolVar(&c.addDeployLabels, "label-secrets", true, "Label secrets generated with --label-version-key and --label-managed-key")
	fs.StringVar(&c.labelSecretGeneration, "generation", strconv.FormatInt(time.Now().Unix(), 10), "Generation label used when annotating Secrets. See --label-version-key")
	fs.StringVar(&c.labelManagedKey, "label-managed-key", "tumblr.com/managed-secret", "Label all generated Secrets with this key=true")
	fs.StringVar(&c.labelVersionKey, "label-version-key", "tumblr.com/secret-version", "Label all generated Secrets with this key, using the value of --generation")

	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	c.credsRootPaths = credsRepoFlags.ToMapStringString()

	err = c.Validate()
	return &c, err
}

func (c *config) Validate() (err error) {
	requiredDirs := map[string]string{
		"manifests": c.mappingsRootPath,
	}
	requiredFiles := map[string]string{}
	optionalFiles := map[string]string{
		"creds-encryption-key":     c.credsEncryptionKeyFile,
		"creds-key-decryption-key": c.credsKeyDecryptionKeyFile,
	}

	if len(c.credsRootPaths) == 0 {
		return fmt.Errorf("at least 1 --creds-repo argument is required")
	}
	for identifier, path := range c.credsRootPaths {
		if err = validateKeyedResource("creds-repo", identifier, path, directory); err != nil {
			return err
		}
	}
	for flag, value := range optionalFiles {
		if value != "" {
			if err := validateResource(flag, value, file); err != nil {
				return err
			}
		}
	}

	if err = validateResources(requiredFiles, file); err != nil {
		return err
	}
	if err = validateResources(requiredDirs, directory); err != nil {
		return err
	}
	return nil
}

func validateResources(resources map[string]string, resourceType resourceType) error {
	for k, v := range resources {
		if err := validateResource(k, v, resourceType); err != nil {
			return err
		}
	}
	return nil
}

func validateKeyedResource(resourceName string, id string, resourcePath string, t resourceType) error {
	if id == "" {
		return fmt.Errorf("%s requires an identifier=path argument, but no identifier found", resourceName)
	}
	if resourcePath == "" {
		return fmt.Errorf("%s identifier %s requires a value", resourceName, id)
	}
	f, err := os.Open(resourcePath)
	if err != nil {
		return fmt.Errorf("unable to open %s %s argument %s: %s", resourceName, id, resourcePath, err.Error())
	}

	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat %s: %s", resourcePath, err.Error())
	}

	switch t {
	case directory:
		if !s.IsDir() {
			err = fmt.Errorf("%s %s argument %s is not a directory", resourceName, id, resourcePath)
		}
	case file:
		if s.IsDir() {
			err = fmt.Errorf("%s %s argument %s is not a file", resourceName, id, resourcePath)
		}
	default:
		err = fmt.Errorf("unsupported resourceType %d", t)
	}
	return err
}

func validateResource(resourceName string, resourcePath string, t resourceType) error {
	if resourcePath == "" {
		return fmt.Errorf("%s requires an argument", resourceName)
	}
	f, err := os.Open(resourcePath)
	if err != nil {
		return fmt.Errorf("unable to open %s argument %s: %s", resourceName, resourcePath, err.Error())
	}

	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat %s: %s", resourcePath, err.Error())
	}

	switch t {
	case directory:
		if !s.IsDir() {
			err = fmt.Errorf("%s argument %s is not a directory", resourceName, resourcePath)
		}
	case file:
		if s.IsDir() {
			err = fmt.Errorf("%s argument %s is not a file", resourceName, resourcePath)
		}
	default:
		err = fmt.Errorf("unsupported resourceType %d", t)
	}
	return err
}

func (c *config) CredsRootPaths() map[string]string {
	return c.credsRootPaths
}

func (c *config) CredsRootPath(id string) (string, error) {
	v, ok := c.credsRootPaths[id]
	if !ok {
		return "", fmt.Errorf("no creds repo configured for '%s'", id)
	}
	return v, nil
}

func (c *config) CredsEncryptionKeyFile() string {
	return c.credsEncryptionKeyFile
}

func (c *config) CredsKeyDecryptionKeyFile() string {
	return c.credsKeyDecryptionKeyFile
}

func (c *config) ProjectionMappingsRootPath() string {
	return c.mappingsRootPath
}

func (c *config) Debug() bool {
	return c.debug
}

func (c *config) ShowSecrets() bool {
	return c.showSecrets
}

func (c *config) OutputDir() string {
	return c.outputDir
}

func (c *config) Version() string {
	return version.Version
}

func (c *config) BuildDate() string {
	return version.BuildDate
}

func (c *config) LabelVersionKey() string {
	return c.labelVersionKey
}

func (c *config) LabelManagedKey() string {
	return c.labelManagedKey
}

func (c *config) Generation() string {
	return c.labelSecretGeneration
}

func (c *config) AddDeployLabels() bool {
	return c.addDeployLabels
}
