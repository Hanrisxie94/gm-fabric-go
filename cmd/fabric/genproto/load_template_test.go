package genproto

import (
	"os"
	"testing"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/rs/zerolog"
)

func TestLoadTemplate(t *testing.T) {
	ownerDir := os.TempDir()
	const serviceName = "service-name"

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	os.Args = []string{"xxx", "--init", serviceName, "--dir", ownerDir}
	cfg, err := config.Load(logger)
	if err != nil {
		t.Fatalf("Load failed: %s", err)
	}

	t.Logf("cfg.TemplateCachePath() = %v", cfg.TemplateCachePath())
}
