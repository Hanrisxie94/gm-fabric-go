package templates

import (
	"github.com/pkg/errors"

	goversion "github.com/hashicorp/go-version"

	"github.com/deciphernow/gm-fabric-go/version"
)

func verifyVersion(minFabricVersionRaw string) error {
	var minFabricVersion *goversion.Version
	var fabricVersion *goversion.Version
	var err error

	minFabricVersion, err = goversion.NewVersion(minFabricVersionRaw)
	if err != nil {
		return errors.Wrapf(err, "version.NewVersion(%s)", minFabricVersionRaw)
	}

	fabricVersion, err = goversion.NewVersion(version.Raw())
	if err != nil {
		return errors.Wrapf(err, "version.NewVersion(%s)", version.Raw())
	}

	if fabricVersion.LessThan(minFabricVersion) {
		return errors.Errorf(
			"fabric generator version %s older than minimum version %s",
			fabricVersion.String(),
			minFabricVersion.String(),
		)
	}

	return nil
}
