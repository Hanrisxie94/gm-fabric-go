package confutil

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testGMConfigTemplate = `
zookeeper:
addrs: {{with $x := "ZK_ADDRS" | getenv }}{{$x}}{{end}}
announcePath: {{with $x := "ZK_ANNOUNCE_PATH" | getenv }}{{$x}}{{end}}
announceHost: {{with $x := "ZK_ANNOUNCE_HOST" | getenv }}{{$x}}{{end}}
announcePort: {{with $x := "ZK_ANNOUNCE_PORT" | getenv }}{{$x}}{{end}}
usingTLS: {{with $x := "ZK_USING_TLS" | getbool }}{{$x}}{{else}}false{{end}}
metrics:
	announcePort: {{with $x := "ZK_METRICS_ANNOUNCE_PORT" | getenv }}{{$x}}{{end}}
`

const validGMConfig = `
zookeeper:
addrs: dashboard:2181,zk:2181
announcePath: /services/examples/0.0.1
announceHost: 
announcePort: 8080
usingTLS: true
metrics:
	announcePort: 
`

func TestCreateConfigFromTemplate(t *testing.T) {

	testPath := filepath.Join(os.TempDir(), "TestCreateConfigFromTemplate")
	err := os.Remove(testPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("os.Remove(%s) failed: %s", testPath, err)
	}

	for _, e := range []struct {
		key   string
		value string
	}{
		{"ZK_ADDRS", "dashboard:2181,zk:2181"},
		{"ZK_ANNOUNCE_PATH", "/services/examples/0.0.1"},
		{"ZK_ANNOUNCE_HOST", ""},
		{"ZK_ANNOUNCE_PORT", "8080"},
		{"ZK_USING_TLS", "true"},
		{"ZK_METRICS_ANNOUNCE_PORT", ""},
	} {
		os.Setenv(e.key, e.value)
	}

	err = CreateConfigFromTemplate(testGMConfigTemplate, testPath)
	if err != nil {
		t.Fatalf("CreateConfigFromTemplate failed: %s", err)
	}

	testResultBytes, err := ioutil.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ioutil.ReadFile(%s) failed: %s", testPath, err)
	}
	testResult := strings.TrimSpace(string(testResultBytes))

	expectedResult := strings.TrimSpace(validGMConfig)
	if testResult != expectedResult {
		for i, r := range testResult {
			t.Logf("('%c', '%c'\n", r, rune(expectedResult[i]))
			if r != rune(expectedResult[i]) {
				t.Fatalf("mismatch at %d '%c' != '%c'", i, r, rune(expectedResult[i]))
			}
		}
		t.Fatalf("result (%d):\n%s\n!= expected result (%d):\n%s\n",
			len(testResult), testResult,
			len(expectedResult), expectedResult)
	}
}

func TestCreateConfigFromBase64(t *testing.T) {
	const testEnvVar = "TEST_CREATE_CONFIG_FROM_BASE64"

	testPath := filepath.Join(os.TempDir(), "TestCreateConfigFromBase64")
	err := os.Remove(testPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("os.Remove(%s) failed: %s", testPath, err)
	}

	encodedData := base64.StdEncoding.EncodeToString([]byte(validGMConfig))

	// we expect no error if the environment variable isn't set
	os.Setenv(testEnvVar, "")
	err = CreateConfigFromBase64(testEnvVar, testPath)
	if err != nil {
		t.Fatalf("CreateConfigFromBase64 failed: %s", err)
	}

	os.Setenv(testEnvVar, encodedData)
	err = CreateConfigFromBase64(testEnvVar, testPath)
	if err != nil {
		t.Fatalf("CreateConfigFromBase64 failed: %s", err)
	}

	testResultBytes, err := ioutil.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ioutil.ReadFile(%s) failed: %s", testPath, err)
	}
	testResult := strings.TrimSpace(string(testResultBytes))

	expectedResult := strings.TrimSpace(validGMConfig)
	if testResult != expectedResult {
		for i, r := range testResult {
			t.Logf("('%c', '%c'\n", r, rune(expectedResult[i]))
			if r != rune(expectedResult[i]) {
				t.Fatalf("mismatch at %d '%c' != '%c'", i, r, rune(expectedResult[i]))
			}
		}
		t.Fatalf("result (%d):\n%s\n!= expected result (%d):\n%s\n",
			len(testResult), testResult,
			len(expectedResult), expectedResult)
	}
}
