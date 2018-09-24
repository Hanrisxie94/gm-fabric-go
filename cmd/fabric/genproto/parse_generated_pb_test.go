package genproto

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
)

func TestParseGeneratedPB(t *testing.T) {
	ownerDir := os.TempDir()
	const serviceName = "service-name"

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	os.Args = []string{"xxx", "--init", serviceName, "--dir", ownerDir}
	cfg, err := config.Load(logger)
	if err != nil {
		t.Fatalf("Load failed: %s", err)
	}
	err = os.MkdirAll(cfg.ProtoPath(), os.ModePerm)
	if err != nil {
		t.Fatalf("os.MkdirAll(%s) failed: %s", cfg.ProtoPath(), err)
	}

	// test that we get an error if there is no file
	err = os.Remove(cfg.GeneratedPBFilePath())
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("os.Remove(%s) failed: %s", cfg.GeneratedPBFilePath(), err)
	}

	_, err = parseGeneratedPBFile(cfg, logger)
	if err == nil {
		t.Errorf("expected error for nonexistent generated file: %v", err)
	}

	// now test parsing a real file
	err = loadTestData(cfg.GeneratedPBFilePath())
	if err != nil {
		t.Errorf("loadTestData failed: %v", err)
	}
	_, err = parseGeneratedPBFile(cfg, logger)
	if err != nil {
		t.Errorf("parseGeneratedPBFile failed: %v", err)
	}

}

func loadTestData(destPath string) error {
	const testData = `
type TestServer interface {
	// Hello simply says 'hello' to the server
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	// HelloProxy says 'hello' in a form that is handled by the gateway proxy
	HelloProxy(context.Context, *HelloRequest) (*HelloRequest, error)
	// HelloStream returns multiple replies
	HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error
}`
	return ioutil.WriteFile(destPath, []byte(testData), os.ModePerm)
}
