package v1

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/tumblr/k8s-secret-projector/internal/pkg/testing" // hack to make test fixtures non-relative
	"github.com/tumblr/k8s-secret-projector/pkg/conf"
	"github.com/tumblr/k8s-secret-projector/pkg/encryption"
)

var (
	testEncryptionConfigCbc = conf.Encryption{
		Module:                "cbc",
		IncludeDecryptionKeys: true,
		CredsKeysFilePath:     credsEncryptionKey,
	}
	relManifestsPath = "test/fixtures/manifests"
	relFixtures      = "test/fixtures"
	testManifests    = map[string]string{
		"json-test-1":                    path.Join(relManifestsPath, "manifest_1.yaml"),
		"raw-test-1":                     path.Join(relManifestsPath, "raw.yaml"),
		"deprecated-1":                   path.Join(relManifestsPath, "deprecated_1.yaml"),
		"missing-source-1":               path.Join(relManifestsPath, "missing_source.yaml"),
		"structured-json-1":              path.Join(relManifestsPath, "structured_subset_json_1.yaml"),
		"structured-json-2":              path.Join(relManifestsPath, "structured_subset_json_2_badformat.yaml"),
		"structured-json-3":              path.Join(relManifestsPath, "structured_subset_json_3_as_yaml.yaml"),
		"structured-yaml-1":              path.Join(relManifestsPath, "structured_subset_yaml_1.yaml"),
		"structured-yaml-2":              path.Join(relManifestsPath, "structured_subset_yaml_2_badformat.yaml"),
		"structured-yaml-3":              path.Join(relManifestsPath, "structured_subset_yaml_3_as_json.yaml"),
		"json-slice-extraction-1":        path.Join(relManifestsPath, "json-slice-extraction-1.yaml"),
		"plugin-cbc-enc-nodecryptkeys":   path.Join(relManifestsPath, "plugin-cbc-enc-nodecryptkeys.yaml"),
		"plugin-cbc-enc-withdecryptkeys": path.Join(relManifestsPath, "plugin-cbc-enc-withdecryptkeys.yaml"),
		"missing-pluginpath-1":           path.Join(relManifestsPath, "missing-pluginpath-1.yaml"),
	}

	testManifestStrings = map[string]string{
		"json-test-1":             "json-tests/test1:production{single-json-key:json:object1.json,another-json-key:json:object1.json,array-json:json:object1.json}",
		"raw-test-1":              "raw-test1/test2:production{raw-file:raw:raw1.txt}",
		"deprecated-1":            "unittest/deprecated:test{single-json-key:json:object1.json,another-json-key:json:object1.json}",
		"missing-source-1":        "unittest/missingsource:production{missing-source:raw:test/doesnt-exist.txt}",
		"structured-json-1":       "json-tests/test-json-subset:production{secrets.json:json:object1.json}",
		"structured-json-3":       "json-tests/test-json-subset:production{secrets.yaml:json:object1.json}",
		"structured-yaml-1":       "yaml-tests/test-yaml-subset:production{secrets.yaml:yaml:object1.yaml}",
		"structured-yaml-3":       "yaml-tests/test-yaml-subset:production{secrets.json:yaml:object1.yaml}",
		"json-slice-extraction-1": "json-tests/test-array-extraction:production{array:json:object1.json,array-field-extraction-0:json:object1.json,nesting-array-0:json:object1.json,nesting-array:json:object1.json}",
	}

	expectedSecrets = map[string]string{
		"structured-json-1":       readFixtureSecret("structured-json-1"),
		"structured-json-3":       readFixtureSecret("structured-json-3"),
		"structured-yaml-1":       readFixtureSecret("structured-yaml-1"),
		"structured-yaml-3":       readFixtureSecret("structured-yaml-3"),
		"json-slice-extraction-1": readFixtureSecret("json-slice-extraction-1"),
		"json-test-1":             readFixtureSecret("json-test-1"),
		"raw-test-1":              readFixtureSecret("raw-test-1"),
	}
)

func getTestConfig() TestConfig {
	return TestConfig{
		credsEncryptionKeyFile:    credsEncryptionKey,
		credsKeyDecryptionKeyFile: "",
	}
}

func readFixtureSecret(n string) string {
	s, err := ioutil.ReadFile(path.Join(relFixtures, "secrets", fmt.Sprintf("%s.yaml", n)))
	if err != nil {
		panic(err)
	}
	return string(s)
}

type TestConfig struct {
	credsEncryptionKeyFile    string
	credsKeyDecryptionKeyFile string
}

func (c *TestConfig) CredsKeyDecryptionKeyFile() string {
	return ""
}
func (c *TestConfig) CredsEncryptionKeyFile() string {
	return c.credsEncryptionKeyFile
}

func (c *TestConfig) CredsRootPath(id string) (string, error) {
	return "", nil
}

func (c *TestConfig) CredsRootPaths() map[string]string {
	return map[string]string{}
}

func (c *TestConfig) ProjectionMappingsRootPath() string {
	return ""
}

func (c *TestConfig) Debug() bool {
	return false
}

func (c *TestConfig) ShowSecrets() bool {
	return false
}

func (c *TestConfig) OutputDir() string {
	return ""
}

func (c *TestConfig) Version() string {
	return ""
}

func (c *TestConfig) BuildDate() string {
	return ""
}
func (c *TestConfig) AddDeployLabels() bool {
	return true
}

func (c *TestConfig) LabelVersionKey() string {
	return "test/tumblr-version"
}

func (c *TestConfig) LabelManagedKey() string {
	return "test/managed"
}

func (c *TestConfig) Generation() string {
	return "6969420"
}

func TestSecret_Project_Plugin_Encryption_WithKeys(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		c := getTestConfig()

		var tests = map[string]map[string][]string{
			"plugin-cbc-enc-withdecryptkeys": map[string][]string{
				// data:  [expected, actual]
				"secrets.json.enc": []string{`{"float":1.23,"key1":"foo","list":["abc","def","ghi"]}`, ""},
				"keys_1.json":      []string{`{"password":"ell0_OliV3r!"}`, ""},
			},
		}

		for test := range tests {
			t.Logf("test %s", test)
			p, ok := testManifests[test]
			if !ok {
				t.Fatalf("unable to find %s in testManifests (where is the test fixture?!)", test)
			}
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			m, err := LoadFromYamlBytes(data, &c)
			if err != nil {
				t.Fatal(err)
			}
			secret, err := m.ProjectSecret(credsPath)

			if err != nil {
				t.Fatal(err.Error())
			}
			e, err := encryption.NewModuleFromEncryptionConfig(conf.Encryption{
				Module:                "plugin",
				PluginPath:            "cbc.so",
				IncludeDecryptionKeys: true,
				CredsKeysFilePath:     credsEncryptionKey,
			})

			if err != nil {
				t.Fatal(err)
			}
			if e == nil {
				t.Fatal("failed to instantiate Tumblr creds keychain")
			}
			t.Log("testing decryption key delivery and data encryption/decryption")
			for k, v := range secret.Data {
				if filepath.Ext(k) == ".enc" {
					// first try to decrypt, then
					// store the decrypted value as actual val
					d, err := e.Decrypt(v)
					if err != nil {
						t.Fatalf("failed to decrypt %s: %s", k, err.Error())
					}
					t.Logf("encryption/decryption successful %s %s", k, d)
					tests[test][k][1] = string(d)
				} else {
					// store the raw value as actual
					tests[test][k][1] = string(v)
				}
			}
			for k, v := range tests[test] {
				t.Logf("[%s] test returned data items (unencrypted where possible): %s -> %s", test, k, v[1])
			}
			// there should be exactly 1 keys_*.json entry added
			if tests[test]["keys_1.json"][1] == "" {
				t.Fatalf("[%s] missing all keys_1.json key in data", test)
			}

			// assert the value of the keys are correct
			for k, tst := range tests[test] {
				var actual = tst[1]
				var expected = tst[0]
				if actual != expected {
					t.Fatalf("[%s] expected item %s to be %v but got %v", test, k, expected, actual)
				}
			}
		}
	} else {
		t.Skipf("Skipping because OS=%s is unsupported", runtime.GOOS)
	}
}

func TestSecret_Project_Plugin_Encryption_OmitKeys(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		c := getTestConfig()

		var tests = map[string]map[string][]string{
			"plugin-cbc-enc-nodecryptkeys": map[string][]string{
				// data:  [expected, actual]
				"secrets.json.enc": []string{`{"float":1.23,"key1":"foo","list":["abc","def","ghi"]}`, ""},
			},
		}

		for test := range tests {
			t.Logf("test %s", test)
			p, ok := testManifests[test]
			if !ok {
				t.Fatalf("unable to find %s in testManifests (where is the test fixture?!)", test)
			}
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			m, err := LoadFromYamlBytes(data, &c)
			if err != nil {
				t.Fatal(err)
			}
			secret, err := m.ProjectSecret(credsPath)

			if err != nil {
				t.Fatal(err.Error())
			}
			e, err := encryption.NewModuleFromEncryptionConfig(m.GetEncryptionConfig())

			if err != nil {
				t.Fatal(err)
			}
			if e == nil {
				t.Fatal("failed to instantiate Tumblr creds keychain")
			}
			t.Log("testing decryption key delivery and data encryption/decryption")
			for k, v := range secret.Data {
				t.Logf("[%s] Secret Data item: %s=%s", test, k, v)
				if filepath.Ext(k) == ".enc" {
					// first try to decrypt, then
					// store the decrypted value as actual val
					t.Logf("[%s] decryption item: %s", test, k)
					d, err := e.Decrypt(v)
					if err != nil {
						t.Fatalf("failed to decrypt %s: %s", k, err.Error())
					}
					t.Logf("encryption/decryption successful %s %s", k, d)
					tests[test][k][1] = string(d)
				} else {
					t.Logf("[%s] non-encrypted item: %s=%s", test, k, v)
					// store the raw value as actual
					tests[test][k][1] = string(v)
				}
			}
			for k, v := range tests[test] {
				t.Logf("[%s] test returned data items: %s -> %s", test, k, v[1])
			}
			// there should be NO keys matching keys_*.json
			for _, s := range []string{"keys_1.json"} {
				for k, v := range secret.Data {
					if k == s {
						t.Fatalf("[%s] expected there would be no decryption key %s included with the secret, but found %s", test, k, v)
					}
				}
			}

			// assert the value of the keys are correct
			for k, tst := range tests[test] {
				var actual = tst[1]
				var expected = tst[0]
				if actual != expected {
					t.Fatalf("[%s] expected item %s to be %v but got %v", test, k, expected, actual)
				}
			}
		}
	} else {
		t.Skipf("Skipping because OS=%s is unsupported", runtime.GOOS)
	}
}

func TestSecret_Project_Encryption_CBC(t *testing.T) {
	c := getTestConfig()
	// expected/actual values for the keys_N.json values
	var tests = map[string]map[string][]string{
		"plugin-cbc-enc-withdecryptkeys": map[string][]string{
			// data:  [expected, actual]
			"secrets.json.enc": []string{`{"float":1.23,"key1":"foo","list":["abc","def","ghi"]}`, ""},
			"keys_1.json":      []string{`{"password":"ell0_OliV3r!"}`, ""},
		},
	}

	for test := range tests {
		t.Logf("test %s", test)
		p := testManifests[test]
		data, err := ioutil.ReadFile(p)
		if err != nil {
			t.Fatalf("unable to find %s in testManifests (where is the test fixture?!)", test)
		}
		m, err := LoadFromYamlBytes(data, &c)
		if err != nil {
			t.Fatal(err)
		}
		secret, err := m.ProjectSecret(credsPath)

		if err != nil {
			t.Fatal(err.Error())
		}
		e, err := encryption.NewModuleFromEncryptionConfig(testEncryptionConfigCbc)
		if err != nil {
			t.Fatalf("failed to instantiate encryption module for %s: %s", test, err.Error())
		}
		t.Log("testing decryption key delivery and data encryption/decryption")
		for k, v := range secret.Data {
			if filepath.Ext(k) == ".enc" {
				// first try to decrypt, then
				// store the decrypted value as actual val
				d, err := e.Decrypt(v)
				if err != nil {
					t.Fatalf("failed to decrypt %s: %s", k, err.Error())
				}
				t.Logf("encryption/decryption successful %s %s", k, d)
				tests[test][k][1] = string(d)
			} else {
				// store the raw value as actual
				tests[test][k][1] = string(v)
			}
		}
		for _, expectedKey := range []string{"keys_1.json"} {
			if _, ok := tests[test][expectedKey]; !ok || tests[test][expectedKey][1] == "" {
				t.Fatalf("[%s] missing %s decryption key", test, expectedKey)
			}
		}

		// assert the value of the keys are correct
		for k, tst := range tests[test] {
			var actual = tst[1]
			var expected = tst[0]
			if actual != expected {
				t.Fatalf("[%s] expected item %s to be %v but got %v", test, k, expected, actual)
			}
		}
	}
}

func TestJSONManifest(t *testing.T) {
	for _, test := range []string{"json-test-1", "json-slice-extraction-1"} {
		config := getTestConfig()
		path := testManifests[test]
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m, err := LoadFromYamlBytes(data, &config)
		if err != nil {
			t.Fatal(err)
		}
		if m.String() != testManifestStrings[test] {
			t.Fatalf("Expected %s.String() to be %s, got %s", test, m.String(), testManifestStrings[test])
		}
		_, err = m.ProjectSecret(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		yamlString, err := m.ProjectSecretAsYAMLString(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		if yamlString != expectedSecrets[test] {
			t.Fatalf("Expected %s Secret to be:\n%s\nBut got:\n%s\n", test, expectedSecrets[test], yamlString)
		}
	}
}

func TestRawManifest(t *testing.T) {
	test := "raw-test-1"
	path := testManifests[test]
	config := getTestConfig()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := LoadFromYamlBytes(data, &config)
	if err != nil {
		t.Fatal(err)
	}

	if m.String() != testManifestStrings[test] {
		t.Fatalf("Expected %s.String() to be %s, got %s", test, m.String(), testManifestStrings[test])
	}

	_, err = m.ProjectSecret(credsPath)
	if err != nil {
		t.Fatal(err)
	}

	yamlString, err := m.ProjectSecretAsYAMLString(credsPath)
	if err != nil {
		t.Fatal(err)
	}
	if yamlString != expectedSecrets[test] {
		t.Fatalf("Expected %s Secret to be:\n%s\nBut got:\n%s\n", test, expectedSecrets[test], yamlString)
	}
}

func TestMissingEncryptionPluginPath(t *testing.T) {
	// we shouldnt be able to load a manifest with a missing plugin
	test := "missing-pluginpath-1"
	expectedErr := `plugin.Open("plugin-missing.so"): realpath failed`
	config := getTestConfig()
	path := testManifests[test]
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = LoadFromYamlBytes(data, &config)
	if err == nil {
		t.Fatal("expected error, but got none")
	}
	if err.Error() != expectedErr {
		t.Fatalf("expected error %s but got %s", expectedErr, err.Error())
	}
}

func TestMissingSourceManifest(t *testing.T) {
	test := "missing-source-1"
	path := testManifests[test]
	config := getTestConfig()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := LoadFromYamlBytes(data, &config)
	if err != nil {
		t.Fatal(err)
	}
	if m.String() != testManifestStrings[test] {
		t.Fatalf("Expected %s.String() to be %s, got %s", test, m.String(), testManifestStrings[test])
	}
	_, err = m.ProjectSecret(credsPath)
	if err.Error() != "open test/fixtures/files/test/doesnt-exist.txt: no such file or directory" {
		t.Fatalf("expected unable to open up non-existent file, but didnt error correctly. Got: %s", err.Error())
	}
}

func TestStructuredProjection(t *testing.T) {
	tests := []string{
		"structured-json-1",
		"structured-json-3",
		"structured-yaml-1",
		"structured-yaml-3",
	}
	for _, test := range tests {
		config := getTestConfig()
		path := testManifests[test]
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m, err := LoadFromYamlBytes(data, &config)
		if err != nil {
			t.Fatal(err)
		}

		if m.String() != testManifestStrings[test] {
			t.Fatalf("Expected %s.String() to be %s, got %s", test, m.String(), testManifestStrings[test])
		}

		d, err := m.ProjectSecret(credsPath)
		if err != nil {
			t.Fatal(err)
		}

		for filename, secrets := range d.Data {
			t.Logf("%s:%s", filename, string(secrets))
		}

		yamlString, err := m.ProjectSecretAsYAMLString(credsPath)
		if err != nil {
			t.Fatal(err)
		}
		if yamlString != expectedSecrets[test] {
			t.Fatalf("Expected %s Secret to be:\n%s\nBut got:\n%s\n", test, expectedSecrets[test], yamlString)
		}
	}
	explodyTests := []string{
		"structured-json-2",
		"structured-yaml-2",
	}
	for _, test := range explodyTests {
		config := getTestConfig()
		path := testManifests[test]
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m, err := LoadFromYamlBytes(data, &config)
		if err != nil {
			t.Fatal(err)
		}

		_, err = m.ProjectSecret(credsPath)
		if err == nil {
			t.Fatalf("expected projecting %s (%s) would result in error, but didnt get an error", test, path)
		}
	}
}
