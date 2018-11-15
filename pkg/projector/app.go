package projector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/types"
	"github.com/tumblr/k8s-secret-projector/pkg/types/v1"
)

var (
	errUnknown = errors.New("unknown error")
)

type app struct {
	conf.Config
}

// App is the thing that does the needful
type App interface {
	LoadProjectionMappings() ([]types.ProjectionMapping, error)
}

// New returns a new App
func New(c conf.Config) App {
	return &app{
		c,
	}
}

// LoadProjectionMappings returns the list of projection mappings from the projection mappings root path
func (a *app) LoadProjectionMappings() ([]types.ProjectionMapping, error) {
	projectionMappings := []types.ProjectionMapping{}
	errs := []error{}
	filepath.Walk(a.ProjectionMappingsRootPath(), func(path string, info os.FileInfo, err error) error {
		// for each path, test that is is a yaml file, and load it
		if info == nil || info.IsDir() {
			return nil
		}
		// test for *.yaml suffix
		if !strings.HasSuffix(info.Name(), ".yaml") {
			// skip this file
			return nil
		}

		raw, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Error reading projection mapping %s: %s\n", path, err.Error())
			errs = append(errs, err)
			return err
		}
		m, err := v1.LoadFromYamlBytes(raw, a.Config)
		if err != nil {
			log.Printf("Error loading projection mapping %s: %s\n", path, err.Error())
			errs = append(errs, err)
			return err
		}
		if a.Debug() {
			log.Printf("Loaded projection mapping: %s\n", m)
		}
		projectionMappings = append(projectionMappings, m)
		return nil
	})

	if len(errs) > 0 {
		return projectionMappings, fmt.Errorf("unable to load %d projection mappings", len(errs))
	}

	return projectionMappings, nil
}
