package v1

import (
	"bytes"
	"testing"

	_ "github.com/tumblr/k8s-secret-projector/internal/pkg/testing"
	"github.com/tumblr/k8s-secret-projector/pkg/types"
)

var (
	credsPath          = "test/fixtures/files"
	credsEncryptionKey = "test/fixtures/files/encryption-cbc-key.json"
	jsonTestFile       = "object1.json"
	yamlTestFile       = "object1.yaml"
	rawTestFile        = "raw1.txt"
)

func TestDataSourceTypes(t *testing.T) {
	j := DataSource{JSON: jsonTestFile}
	y := DataSource{YAML: yamlTestFile}
	r := DataSource{Raw: rawTestFile}
	if j.Type() != types.JSONType {
		t.Error("JSON datasource doesnt report JSONType as its Type()!")
	}
	if y.Type() != types.YAMLType {
		t.Error("YAML datasource doesnt report YAMLType as its Type()!")
	}
	if r.Type() != types.RawType {
		t.Error("Raw datasource doesnt report RawType as its Type()!")
	}
}
func TestProjectRaw(t *testing.T) {
	d := DataSource{Raw: rawTestFile}
	x, err := d.Project(credsPath)
	if err != nil {
		t.Error(err)
	}
	// test the contents are what we expect
	expected := bytes.NewBufferString("hello\nthis is a raw file\n")
	if bytes.Compare(x, expected.Bytes()) != 0 {
		t.Errorf("Expected %v, got %v\n", expected.String(), bytes.NewBuffer(x).String())
	}
}

func TestProjectRawMissingFile(t *testing.T) {
	d := DataSource{Raw: "file/doesnt/exist.txt"}
	_, err := d.Project(credsPath)
	if err == nil {
		t.Fatal("expected error for file not existing, but didnt get one")
	}
}

/** JSON datasource type tests **/

var jsonTests = map[string]string{
	"$.secret":              "paSsw0rd!",
	"$.nesting.key1":        "foo",
	"$.nesting.int":         "12345",
	"$.nesting.float":       "1.23",
	"$.nesting.bool":        "true",
	"$.nesting.list[1]":     "def",
	"$.nesting.list":        "abc,def,ghi",
	"$.nesting.list-string": "foo,bar",
}

func TestProjectJSON(t *testing.T) {
	for path, expected := range jsonTests {
		d := DataSource{JSON: jsonTestFile, JSONPath: path}
		x, err := d.Project(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		expected := bytes.NewBufferString(expected)
		if bytes.Compare(x, expected.Bytes()) != 0 {
			t.Errorf("Expected %s would project %v, got %v\n", path, expected.String(), bytes.NewBuffer(x).String())
		}
	}
}

func TestProjectJSONPathsError(t *testing.T) {
	emptyData := DataSource{JSON: jsonTestFile, JSONPaths: map[string]types.JSONPathSelector{}}
	_, err := emptyData.projectJSON(credsPath)

	if err == nil {
		t.Fatal("should expect to fail on an empty map for JSONPaths")
	}

	badKey := DataSource{JSON: jsonTestFile, JSONPaths: map[string]types.JSONPathSelector{"invalidKey": "invalidKey"}}
	_, err = badKey.projectJSON(credsPath)

	if err == nil {
		t.Fatal("should fail on bad key")
	}
}

func TestProjectYAMLWithJSONPaths(t *testing.T) {
	// NOTE: make sure we test both projecting as default format (YAML), and JSON!
	testSources := map[string]DataSource{
		`bool: true
listlabel:
- abc
- def
- ghi
secret: paSsw0rd!
`: DataSource{YAML: yamlTestFile, JSONPaths: map[string]types.JSONPathSelector{
			"secret": "$.secret", "bool": "$.nesting.bool", "listlabel": "$.nesting.list"}},
		`{"bool":true,"listlabel":["abc","def","ghi"],"secret":"paSsw0rd!"}`: DataSource{YAML: yamlTestFile, Format: types.FormatJSON, JSONPaths: map[string]types.JSONPathSelector{
			"secret": "$.secret", "bool": "$.nesting.bool", "listlabel": "$.nesting.list"}},
	}
	for expected, d := range testSources {
		x, err := d.projectYAML(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(x, []byte(expected)) != 0 {
			t.Errorf("Expected would project:\n%s\nGot:\n%s\n", []byte(expected), x)
		}
	}
}

func TestProjectJSONWithJSONPaths(t *testing.T) {
	// NOTE: make sure we test both projecting as default format, and YAML!
	testSources := map[string]DataSource{
		`{"bool":true,"listlabel":["abc","def","ghi"],"secret":"paSsw0rd!"}`: DataSource{JSON: jsonTestFile, JSONPaths: map[string]types.JSONPathSelector{
			"secret": "$.secret", "bool": "$.nesting.bool", "listlabel": "$.nesting.list"}},
		`bool: true
listlabel:
- abc
- def
- ghi
secret: paSsw0rd!
`: DataSource{JSON: jsonTestFile, Format: types.FormatYAML, JSONPaths: map[string]types.JSONPathSelector{
			"secret": "$.secret", "bool": "$.nesting.bool", "listlabel": "$.nesting.list"}},
	}
	for expected, d := range testSources {
		x, err := d.projectJSON(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(x, []byte(expected)) != 0 {
			t.Errorf("Expected would project:\n%s\nGot:\n%s\n", []byte(expected), x)
		}
	}
}

func TestProjectJSONUnknownType(t *testing.T) {
	jsonpaths := []string{"$.nesting.list-int", "$.nesting.map", "$.nesting.list-float"}
	for _, path := range jsonpaths {
		d := DataSource{JSON: jsonTestFile, JSONPath: path}
		_, err := d.Project(credsPath)
		if err == nil {
			t.Fatalf("expected projecting %s would fail, but got no error", path)
		}
	}
}

/** YAML Datasource Tests **/

var yamlTests = map[string]string{
	"$.secret":              "paSsw0rd!",
	"$.nesting.key1":        "foo",
	"$.nesting.integer":     "420",
	"$.nesting.float":       "-69.6969",
	"$.nesting.bool":        "true",
	"$.nesting.list[1]":     "def",
	"$.nesting.list":        "abc,def,ghi",
	"$.nesting.list-string": "foo,bar",
}

func TestProjectYAML(t *testing.T) {
	for path, expected := range yamlTests {
		d := DataSource{YAML: yamlTestFile, JSONPath: path}
		x, err := d.Project(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		expected := bytes.NewBufferString(expected)
		if bytes.Compare(x, expected.Bytes()) != 0 {
			t.Errorf("Expected %v, got %v\n", expected.String(), bytes.NewBuffer(x).String())
		}
	}
}
func TestProjectYAMLUnknownType(t *testing.T) {
	jsonpaths := []string{"$.nesting.list-int", "$.nesting.map", "$.nesting.list-float"}
	for _, path := range jsonpaths {
		d := DataSource{YAML: yamlTestFile, JSONPath: path}
		_, err := d.Project(credsPath)
		if err == nil {
			t.Fatalf("expected projecting %s would fail, but got no error", path)
		}
	}
}

/** structured projection tests **/

func TestStructuredDataSourceOutputFormatInference(t *testing.T) {
	// assert the OutputFormat introspection and inference works as expected
	selectors := map[string]types.JSONPathSelector{
		"key1":  "$.nesting.key1",
		"float": "$.nesting.float",
	}
	path := "$.nesting.key1"
	dses := []DataSource{
		DataSource{YAML: yamlTestFile, JSONPath: path},
		DataSource{JSON: jsonTestFile, JSONPath: path},
		DataSource{Raw: rawTestFile},
		DataSource{YAML: yamlTestFile, JSONPath: path, Format: types.FormatRaw},
		DataSource{JSON: jsonTestFile, JSONPath: path, Format: types.FormatRaw},
		DataSource{Raw: rawTestFile, Format: types.FormatRaw},
		DataSource{YAML: yamlTestFile, JSONPaths: selectors},
		DataSource{JSON: jsonTestFile, JSONPaths: selectors},
	}
	expectedFormats := []types.OutputFormat{
		types.FormatRaw,
		types.FormatRaw,
		types.FormatRaw,
		types.FormatRaw,
		types.FormatRaw,
		types.FormatRaw,
		types.FormatYAML,
		types.FormatJSON,
	}
	for i, ds := range dses {
		f, err := ds.OutputFormat()
		if err != nil {
			t.Fatal(err)
		}
		if f != expectedFormats[i] {
			t.Fatalf("Expected %s but got %s for datasource %v", expectedFormats[i], f, ds)
		}
	}
	errorDataSources := []DataSource{
		DataSource{YAML: yamlTestFile, JSONPath: path, Format: types.FormatJSON},
		DataSource{YAML: yamlTestFile, JSONPath: path, Format: types.FormatYAML},
		DataSource{JSON: jsonTestFile, JSONPath: path, Format: types.FormatJSON},
		DataSource{JSON: jsonTestFile, JSONPath: path, Format: types.FormatYAML},
		DataSource{Raw: rawTestFile, Format: types.FormatJSON},
		DataSource{Raw: rawTestFile, Format: types.FormatYAML},
	}

	for _, ds := range errorDataSources {
		f, err := ds.OutputFormat()
		if err == nil {
			t.Fatalf("expected error due to incorrect output format for %v but got %s", ds.String(), f)
		}
	}
}
