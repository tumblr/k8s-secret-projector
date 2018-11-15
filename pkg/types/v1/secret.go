package v1

import (
	"fmt"
)

// Secret ...
type Secret struct {
	Name    string     `json:"name",yaml:"name"`
	Encrypt bool       `json:"encrypt",yaml:"encrypt"`
	Source  DataSource `json:"source",yaml:"source"`
}

func (s *Secret) String() string {
	return fmt.Sprintf("%s:%s", s.Name, s.Source.String())
}

// Project returns the []byte of a projected secret and all its datasources
func (s *Secret) Project(credsPath string) ([]byte, error) {
	return s.Source.Project(credsPath)
}
