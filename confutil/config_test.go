package confutil

import (
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
usingTLS: {{with $x := "ZK_USING_TLS" | getenv }}{{$x}}{{end}}
metrics:
	announcePort: {{with $x := "ZK_METRICS_ANNOUNCE_PORT" | getenv }}{{$x}}{{end}}
`

func TestCreateConfigFromTemplate(t *testing.T) {

	expectedResult := strings.TrimSpace(`
zookeeper:
addrs: dashboard:2181,zk:2181
announcePath: /services/examples/0.0.1
announceHost: 
announcePort: 8080
usingTLS: false
metrics:
	announcePort: 
	`)

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
		{"ZK_USING_TLS", "false"},
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
