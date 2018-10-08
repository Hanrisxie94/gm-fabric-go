package genproto

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/rs/zerolog"
)

func TestLoadTemplate(t *testing.T) {
	ownerDir := os.TempDir()
	const serviceName = "service-name"
	const templateName = "testTemplate"

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	cfg := config.Config{
		ServiceName: serviceName,
		OwnerDir:    ownerDir,
	}

	err := os.MkdirAll(cfg.TemplateCachePath(), os.ModePerm)
	if err != nil {
		t.Fatalf("os.MkdirAll(%s) failed: %s", cfg.TemplateCachePath(), err)
	}

	templatePath := filepath.Join(cfg.TemplateCachePath(), templateName)
	err = os.Remove(templatePath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("os.Remove(%s) failed: %s", templatePath, err)
	}

	// we expect an error when there is no file
	_, err = loadTemplateFromCache(
		cfg,
		logger,
		templateName,
	)
	if err == nil {
		t.Fatalf("expecting 'file not found'")
	}

	err = ioutil.WriteFile(templatePath, []byte{'x', 'y', 'z'}, 0777)
	if err != nil {
		t.Fatalf("ioutil.WriteFile '%s' failed: %s", templatePath, err)
	}

	// we expect no error when there is a file
	_, err = loadTemplateFromCache(
		cfg,
		logger,
		templateName,
	)
	if err != nil {
		t.Fatalf("loadTemplateFromCache failed: %v", err)
	}
}
