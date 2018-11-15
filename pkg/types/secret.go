package types

// Secret is a group of datasources bound to a namespace
type Secret interface {
	String() string
	Project(credsPath string) ([]byte, error)
}
