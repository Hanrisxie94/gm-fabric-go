package genproto

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deciphernow/gm-fabric-go/cmd/fabric/config"
	"github.com/rs/zerolog"
)

const (
	streamMethodTemplateName = "stream_method.go"
	streamMethodTemplate     = `package {{.MethodsPackageName}}

import (
	"github.com/pkg/errors"

	{{.ProtobufImport}}
	{{.PBImport}}
)

{{.Comments}}
func (s *serverData) {{.MethodDeclaration}} {
	return errors.New("not implemented")
}
`
	streamMethodPrototype = "HelloStream(*HelloStreamRequest, TestService_HelloStreamServer) error"

	unitaryMethodTemplateName = "unitary_method.go"
	unitaryMethodTemplate     = `package {{.MethodsPackageName}}

import (
	"golang.org/x/net/context"

	"github.com/pkg/errors"
	
	{{.ProtobufImport}}
	{{.PBImport}}
)

{{.Comments}}
func (s *serverData) {{.MethodDeclaration}} {
	return nil, errors.New("not implemented")
}
`
	unitaryMethodPrototype = "HelloProxy(context.Context, *HelloRequest) (*HelloResponse, error)"
)

func TestGenerateMethod(t *testing.T) {
	ownerDir := os.TempDir()
	const serviceName = "service-name"

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	cfg := config.Config{
		ServiceName: serviceName,
		OwnerDir:    ownerDir,
	}
	err := os.MkdirAll(cfg.ProtoPath(), os.ModePerm)
	if err != nil {
		t.Fatalf("os.MkdirAll(%s) failed: %s", cfg.ProtoPath(), err)
	}
	err = os.MkdirAll(cfg.TemplateCachePath(), os.ModePerm)
	if err != nil {
		t.Fatalf("os.MkdirAll(%s) failed: %s", cfg.TemplateCachePath(), err)
	}

	for _, tc := range []struct {
		name   string
		text   string
		pEntry PrototypeEntry
	}{
		{
			name:   streamMethodTemplateName,
			text:   streamMethodTemplate,
			pEntry: PrototypeEntry{nil, streamMethodPrototype},
		},
		{
			name:   unitaryMethodTemplateName,
			text:   unitaryMethodTemplate,
			pEntry: PrototypeEntry{nil, unitaryMethodPrototype},
		},
	} {
		t.Logf("%q", tc)

		templatePath := filepath.Join(cfg.TemplateCachePath(), tc.name)
		err = ioutil.WriteFile(templatePath, []byte(tc.text), 0777)
		if err != nil {
			t.Fatalf("ioutil.WriteFile '%s' failed: %s", templatePath, err)
		}

		testPath := filepath.Join(ownerDir, "method.go")
		err = generateMethod(cfg, logger, tc.pEntry, testPath)
		if err != nil {
			t.Fatalf("generateMethod %s failed: %s", tc.name, err)
		}
	}

}
