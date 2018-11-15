package types

// JSONPathSelector is the json path to a value in a structured datasource
// see github.com/oliveagle/jsonpath for more details.
type JSONPathSelector string

// OutputFormat are the permissable output formats supported by the projector
// for a given secret
type OutputFormat string

const (
	// FormatDefault is the default output format
	FormatDefault OutputFormat = ""
	// FormatRaw is the raw output format (default for 1 datasource field)
	FormatRaw OutputFormat = "raw"
	// FormatJSON is the json output format (default for json sources with multiple extracted fields)
	FormatJSON OutputFormat = "json"
	// FormatYAML is the YAML output format (default for yaml sources with multiple extracted fields)
	FormatYAML OutputFormat = "yaml"
)
