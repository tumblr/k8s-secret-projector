package main // import github.com/tumblr/k8s-secret-projector/cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tumblr/k8s-secret-projector/internal/pkg/version"
	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/projector"
	"github.com/tumblr/k8s-secret-projector/pkg/types"
	k8sv1 "k8s.io/api/core/v1"
)

func main() {
	c, err := conf.LoadConfigFromArgs(os.Args)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	log.Printf("%s version=%s commit=%s branch=%s runtime=%s built=%s", version.Package, version.Version, version.Commit, version.Branch, runtime.Version(), version.BuildDate)
	if c.Debug() {
		for repo, path := range c.CredsRootPaths() {
			log.Printf("creds %s path: %s\n", repo, path)
		}
		log.Printf("projection mappings path: %s\n", c.ProjectionMappingsRootPath())
	}
	app := projector.New(c)

	projectionMappings, err := app.LoadProjectionMappings()
	if err != nil {
		log.Fatalf("Unable to load projection mappings: %s\n", err.Error())
	}
	if len(projectionMappings) == 0 {
		log.Fatal("No projection mappings loaded! Aborting\n")
	}
	log.Printf("Loaded %d projection mappings\n", len(projectionMappings))

	var secrets []k8sv1.Secret
	for _, m := range projectionMappings {
		credsRepoPath := getCredsRepo(c, m)
		if c.Debug() {
			log.Printf("Projecting mapping file: %s\n", m.String())
		}
		k8sSecret, err := m.ProjectSecret(credsRepoPath)
		if err != nil {
			log.Printf("Unable to project %s into a Kubernetes Secret: %s\n", m.String(), err.Error())
			// we will bail out later, dont worry!
			continue
		}
		secrets = append(secrets, *k8sSecret)
		if c.Debug() {
			log.Printf("Generated Secret for %s:\n", k8sSecret.String())
			yamlString, err := m.ProjectSecretAsYAMLString(credsRepoPath)
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Printf(yamlString)
		}
	}

	// fail if we were unable to generate any secret projections
	if len(secrets) != len(projectionMappings) {
		log.Fatalf("Expected we would create %d Secrets, but only successfully created %d\n", len(projectionMappings), len(secrets))
	}

	if c.OutputDir() != "" {
		tUnix := time.Now().Unix()
		dir, err := os.Open(c.OutputDir())
		if err != nil {
			log.Fatalf("Unable to open output dir %s: %s\n", c.OutputDir(), err.Error())
		}
		info, err := dir.Stat()
		if err != nil {
			log.Fatalf("error: output dir %s is jacked up: %s\n", c.OutputDir(), err.Error())
		}
		if !info.Mode().IsDir() {
			log.Fatalf("error: output %s is not a directory\n", c.OutputDir())
		}

		for _, m := range projectionMappings {
			// skip secrets we were told to not write out
			credsRepoPath := getCredsRepo(c, m)
			yamlString, err := m.ProjectSecretAsYAMLString(credsRepoPath)
			if err != nil {
				log.Fatalf("unable to project mapping %s: %s", m.String(), err.Error())
			}
			fname := filepath.Join(c.OutputDir(), fmt.Sprintf("%d-%s-%s.yaml", tUnix, m.GetNamespace(), m.GetName()))

			log.Printf("writing %s/%s Secret to %s...\n", m.GetNamespace(), m.GetName(), fname)
			err = ioutil.WriteFile(fname, []byte(yamlString), 0400)
			if err != nil {
				log.Fatalf("unable to write Secret to %s: %s", fname, err.Error())
			}

		}
	}

	if c.Debug() && c.ShowSecrets() {
		log.Printf("Secrets:\n")
		for _, m := range projectionMappings {
			credsRepoPath := getCredsRepo(c, m)
			yamlString, err := m.ProjectSecretAsYAMLString(credsRepoPath)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Printf("---\n%s", yamlString)
		}
	}
}

func getCredsRepo(c conf.Config, m types.ProjectionMapping) string {
	credsRepoPath, err := c.CredsRootPath(m.GetRepo())
	if err != nil {
		log.Fatalf("Unsupported repo type %s for projection mapping %s/%s (perhaps you missed a --creds-repo=%s=/path/to/repo argument)\n", m.GetRepo(), m.GetNamespace(), m.GetName(), m.GetRepo())
	}

	return credsRepoPath
}
