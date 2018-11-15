package conf

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/tumblr/k8s-secret-projector/internal/pkg/testing"
)

const (
	testGeneration            = "4206969"
	testFolder                = "test/fixtures/files"
	testCredsEncryptionKey    = "test/fixtures/files/aes_params_test_1.json"
	testCredsKeyDecryptionKey = "test/fixtures/files/aes_params_test_3.json"
)

func cloneTestData(toClone map[string]string) (clone map[string]string) {
	clone = make(map[string]string, len(toClone))
	for k, v := range toClone {
		clone[k] = v
	}
	return clone
}

func mapToArgs(m map[string]string) []string {
	var args []string
	args = append(args, os.Args[0])
	for k, v := range m {
		args = append(args, fmt.Sprintf("-%s=%s", k, v))
	}
	return args
}

func configTester(t *testing.T, testData map[string]string, expectedErr string) {
	args := mapToArgs(testData)
	_, err := LoadConfigFromArgs(args)

	if err != nil && expectedErr == "" || err == nil && expectedErr != "" {
		t.Errorf("expected error '%s' but got '%v'", expectedErr, err)
		t.Fail()
	}
}

func TestLoadConfigFromArgs_Test(t *testing.T) {
	allFlagsHappy := map[string]string{
		"creds-repo":               fmt.Sprintf("%s=%s,%s=%s", "foo", testFolder, "bar", testFolder),
		"manifests":                testFolder,
		"creds-encryption-key":     testCredsEncryptionKey,
		"creds-key-decryption-key": testCredsKeyDecryptionKey,
		"generation":               testGeneration,
	}

	//Test happy path
	t.Log("test happy path")
	configTester(t, allFlagsHappy, "")

	//Test missing arguments
	t.Log("test requires creds-repo argument")
	testMissingArg := cloneTestData(allFlagsHappy)
	delete(testMissingArg, "creds-repo")
	configTester(t, testMissingArg, "creds-repo requires an argument")

	//Test invalid diretory
	t.Log("test invalid directory")
	testInvalidDirectory := cloneTestData(allFlagsHappy)
	testInvalidDirectory["manifests"] = "test/fixtures/files/foobar"
	configTester(t, testInvalidDirectory, "unable to open manifests argument test/fixtures/files/foobar: open test/fixtures/files/foobar: no such file or directory")

	//Test invalid file
	t.Log("test invalid file")
	testInvalidFile := cloneTestData(allFlagsHappy)
	testInvalidFile["creds-encryption-key"] = "test/fixtures/files/"
	configTester(t, testInvalidFile, "ello")

}

func TestConfigGeneration(t *testing.T) {
	// test that the default generation is a timestamp
	stampNow := time.Now().Unix()
	args := mapToArgs(map[string]string{
		"creds-repo":               fmt.Sprintf("%s=%s,%s=%s", "development", testFolder, "production", testFolder),
		"manifests":                testFolder,
		"creds-encryption-key":     testCredsEncryptionKey,
		"creds-key-decryption-key": testCredsKeyDecryptionKey,
	})
	c, err := LoadConfigFromArgs(args)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	i, err := strconv.ParseInt(c.Generation(), 10, 64)
	if err != nil {
		t.Errorf("expected generation to be a parsable 64bit int, but got an error: %v", err.Error())
		t.Fail()
	}
	if i < stampNow {
		t.Errorf("expected generation to be a reasonable timestamp >= %d, but got %d", stampNow, i)
		t.Fail()
	}

}
