package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/oliveagle/jsonpath"
	"github.com/tumblr/k8s-secret-projector/pkg/types"
)

var (
	// ErrUnsupportedStructuredOutputFormat is thrown when the output format requested isnt satisfiable by the projection method
	ErrUnsupportedStructuredOutputFormat = errors.New("output format requested is structured, but the input source type does not support structured output")
	// ErrUnsupportedUnstructuredOutputFormat is thrown when the output format requested isnt satisfiable by the projection method
	ErrUnsupportedUnstructuredOutputFormat = errors.New("output format requested is raw, but the input source type is structured")
	// ErrUnsupportedOutputFormat is thrown when the output format requested isnt satisfiable by the projection method
	ErrUnsupportedOutputFormat = errors.New("unsupported output format")
	// ErrMissingJSONPathSelector is thrown when a structured projection doesnt specify jsonpath or jsonpaths
	ErrMissingJSONPathSelector = errors.New("either JSONPath or JSONPaths need to be defined")
	// ErrMultipleJSONPathSelector is thrown when a structured projection specifies both jsonpath and jsonpaths
	ErrMultipleJSONPathSelector = errors.New("only JSONPath or JSONPaths need to be defined")
)

// DataSource is a source of data that will be projected into a secret
// it specifies its source, output format (optional), and fields to extract
// from the source.
type DataSource struct {
	JSON string `json:"json,omitempty",yaml:"json,omitempty"`
	YAML string `json:"yaml,omitempty",yaml:"yaml,omitempty"`
	Raw  string `json:"raw,omitempty",yaml:"raw,omitempty"`
	// Format is the desired output format for the secret. This defaults to the input format
	// unless overridden. See OutputFormat()
	Format   types.OutputFormat `json:"format,omitempty",yaml:"format,omitempty"`
	JSONPath string             `json:"jsonpath,omitempty",yaml:"jsonpath,omitempty"`
	// JSONPaths is different from the singular path, it represents one (or more) json elements
	// being selected from a datasource and exported into a single file
	// the key is the label it should be defined as, the value is the jsonpath
	JSONPaths map[string]types.JSONPathSelector `json:"jsonpath,omitempty",yaml:"jsonpath,omitempty"`
}

// String returns a string representation of the datasource
func (d *DataSource) String() string {
	switch d.Type() {
	case types.JSONType:
		return fmt.Sprintf("json:%s", d.JSON)
	case types.YAMLType:
		return fmt.Sprintf("yaml:%s", d.YAML)
	case types.RawType:
		return fmt.Sprintf("raw:%s", d.Raw)
	default:
		return "unknown"
	}
}

// OutputFormat returns the inferred/requested output format given the source format
// and extracted fields.
func (d *DataSource) OutputFormat() (types.OutputFormat, error) {
	inferredFormat := types.FormatRaw
	switch d.Format {
	case types.FormatDefault:
		// infer based on config
		// That is to say: (source type -> expected Format)
		// * JSON+JSONPath -> 'raw'
		// * JSON+JSONPaths -> 'json'
		// * Raw -> 'raw'
		// * YAML+JSONPath -> 'raw'
		// * YAML+JSONPaths -> 'yaml'

		if d.JSON != "" && d.JSONPath != "" {
			inferredFormat = types.FormatRaw
		} else if d.JSON != "" && len(d.JSONPaths) > 0 {
			inferredFormat = types.FormatJSON
		} else if d.YAML != "" && d.JSONPath != "" {
			inferredFormat = types.FormatRaw
		} else if d.YAML != "" && len(d.JSONPaths) > 0 {
			inferredFormat = types.FormatYAML
		} else if d.Raw != "" {
			inferredFormat = types.FormatRaw
		} else {
			inferredFormat = types.FormatRaw
		}
	case types.FormatJSON:
		inferredFormat = types.FormatJSON
	case types.FormatYAML:
		inferredFormat = types.FormatYAML
	case types.FormatRaw:
		inferredFormat = types.FormatRaw
	default:
		return types.FormatDefault, fmt.Errorf("unsupported output format for source type")
	}
	if d.Raw != "" && inferredFormat != types.FormatRaw {
		return types.FormatDefault, fmt.Errorf("only raw format is supported for raw sources")
	}
	if len(d.JSONPaths) > 0 && inferredFormat == types.FormatRaw {
		return types.FormatDefault, ErrUnsupportedUnstructuredOutputFormat
	}
	if d.JSONPath != "" && inferredFormat != types.FormatRaw {
		return types.FormatDefault, ErrUnsupportedStructuredOutputFormat
	}
	return inferredFormat, nil
}

// Type returns the type of the datasource
func (d *DataSource) Type() types.DataSourceType {
	if d.JSON != "" {
		return types.JSONType
	}
	if d.YAML != "" {
		return types.YAMLType
	}
	if d.Raw != "" {
		return types.RawType
	}
	return types.UnknownType
}

// Project will resolve the data pointed to by this DataSource, and
// return the data referenced by it as a string
func (d *DataSource) Project(credsPath string) ([]byte, error) {
	switch d.Type() {
	case types.JSONType:
		return d.projectJSON(credsPath)
	case types.YAMLType:
		return d.projectYAML(credsPath)
	case types.RawType:
		return d.projectRaw(credsPath)
	default:
		return nil, fmt.Errorf("unable to project unknown type datasource")
	}
}

func (d *DataSource) projectRaw(credsPath string) ([]byte, error) {
	format, err := d.OutputFormat()
	if err != nil {
		return nil, err
	}
	if format != types.FormatRaw {
		return nil, ErrUnsupportedOutputFormat
	}
	// just read the file, and return it as a []byte
	bytes, err := ioutil.ReadFile(filepath.Join(credsPath, d.Raw))
	return bytes, err
}

func (d *DataSource) projectJSON(credsPath string) ([]byte, error) {
	format, err := d.OutputFormat()
	if err != nil {
		return nil, err
	}
	// if format is raw but there are multiple extracted fields, abort!
	// we cant project multiple fields into a raw type output
	if len(d.JSONPaths) > 0 && format == types.FormatRaw {
		return nil, ErrUnsupportedOutputFormat
	}
	if len(d.JSONPath) == 0 && len(d.JSONPaths) == 0 {
		return nil, ErrMissingJSONPathSelector
	}
	if len(d.JSONPath) > 0 && len(d.JSONPaths) > 0 {
		return nil, ErrMultipleJSONPathSelector
	}

	// read the JSON source file
	var jsonData interface{}
	bytes, err := ioutil.ReadFile(filepath.Join(credsPath, d.JSON))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &jsonData)
	if err != nil {
		return nil, err
	}

	// this is the path for handling the jsonPath entry, it parses the field and returns
	// below that is the code for handling jsonPaths
	// NOTE: this bails out before we get to the JSONPaths projection below
	if len(d.JSONPath) > 0 {
		res, err := jsonpath.JsonPathLookup(jsonData, d.JSONPath)
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		return convertInterfaceValueToBytes(res)
	}

	// this is a map of a subset of labels to json fields (which may or may not be structured)
	resArray := map[string]interface{}{}
	for label, path := range d.JSONPaths {
		res, err := jsonpath.JsonPathLookup(jsonData, string(path))
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		resArray[label] = res
	}

	// return the map as a re-serialized byte array
	// based on the requested structured output format
	switch format {
	case types.FormatJSON:
		return json.Marshal(resArray)
	case types.FormatYAML:
		return yaml.Marshal(resArray)
	default:
		return nil, ErrUnsupportedOutputFormat
	}
}

func (d *DataSource) projectYAML(credsPath string) ([]byte, error) {
	format, err := d.OutputFormat()
	if err != nil {
		return nil, err
	}
	// if format is raw but there are multiple extracted fields, abort!
	// we cant project multiple fields into a raw type output
	if len(d.JSONPaths) > 0 && format == types.FormatRaw {
		return nil, ErrUnsupportedOutputFormat
	}
	if len(d.JSONPath) == 0 && len(d.JSONPaths) == 0 {
		return nil, ErrMissingJSONPathSelector
	}
	if len(d.JSONPath) > 0 && len(d.JSONPaths) > 0 {
		return nil, ErrMultipleJSONPathSelector
	}

	// read the YAML file
	var yamlData interface{}
	bytes, err := ioutil.ReadFile(filepath.Join(credsPath, d.YAML))
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %s", filepath.Join(credsPath, d.YAML), err)
	}
	err = yaml.Unmarshal(bytes, &yamlData)
	if err != nil {
		return nil, err
	}

	if d.JSONPath != "" {
		res, err := jsonpath.JsonPathLookup(yamlData, d.JSONPath)
		if err != nil {
			return nil, err
		}
		return convertInterfaceValueToBytes(res)
	}

	// handle multiple extracted fields
	// this is a map of a subset of labels to json fields (which may or may not be structured)
	resArray := map[string]interface{}{}
	for label, path := range d.JSONPaths {
		res, err := jsonpath.JsonPathLookup(yamlData, string(path))
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		resArray[label] = res
	}

	// return the map as a re-serialized byte array
	// based on the requested structured output format
	switch format {
	case types.FormatJSON:
		return json.Marshal(resArray)
	case types.FormatYAML:
		return yaml.Marshal(resArray)
	default:
		return nil, ErrUnsupportedOutputFormat
	}
}

// takes some interface and returns it converted to a byte buffer
// NOTE: returned []byte is little endian encoded
// this does some reflection to ensure we are rendering a value
// correctly.
func convertInterfaceValueToBytes(data interface{}) ([]byte, error) {
	// try to parse the value out as somethign scalar we can
	// convert into a []byte. Never return the wire representation
	// of a value; always convert into a string first!

	if resv, ok := data.(string); ok {
		return []byte(resv), nil
	} else if resv, ok := data.(int64); ok {
		return []byte(strconv.FormatInt(resv, 10)), nil
	} else if resv, ok := data.(int); ok {
		return []byte(strconv.Itoa(resv)), nil
	} else if resv, ok := data.(float64); ok {
		return []byte(strconv.FormatFloat(resv, 'f', -1, 64)), nil
	} else if resv, ok := data.(bool); ok {
		return []byte(strconv.FormatBool(resv)), nil
	} else if v := reflect.ValueOf(data); v.Kind() == reflect.Slice {
		// make a slice of strings to hold this data
		dataInterfaceSlice := data.([]interface{})
		s := make([]string, len(dataInterfaceSlice))
		for i, x := range dataInterfaceSlice {
			// convert this slice into a []string{...}. for now, we only support scalar extraction of []string, no other types of slice
			if resx, ok := x.(string); ok {
				s[i] = resx
			} else {
				return nil, fmt.Errorf("unable extract scalar value from slice, only []string are supported currently. try extracting a specific element. unsupported datatype %v", x)
			}
		}
		return []byte(strings.Join(s, ",")), nil
	}
	return nil, fmt.Errorf("unable extract scalar value, unsupported datatype %v", data)
}
