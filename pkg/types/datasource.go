package types

// DataSourceType is the type of DataSource that is represented
type DataSourceType int

const (
	// UnknownType is the type of datasource that is unsupported
	UnknownType DataSourceType = iota
	// JSONType is the type of datasource that is backed by json
	JSONType
	// YAMLType is the type of datasource that is backed by yaml
	YAMLType
	// RawType is the type of datasource that is backed by a raw file
	RawType
)

// DataSource is an interface for a single secret data source
type DataSource interface {
	String() string
	Type() DataSourceType
	OutputFormat() (OutputFormat, error)
	Project(credsPath string) ([]byte, error)
}
