package cloudobserver

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChooseSessionType(t *testing.T) {
	testCases := []struct {
		configFile         string
		expectedSess       string
		awsRegion          string
		awsProfile         string
		awsAccessKeyId     string
		awsSecretAccessKey string
		awsSessionToken    string
		description        string
	}{
		/*
			{
				configFile:         "testdata/testconfig",
				expectedSess:       "default",
				awsRegion:          "",
				awsProfile:         "",
				awsAccessKeyId:     "",
				awsSecretAccessKey: "",
				awsSessionToken:    "",
				description:        "[0] expecting default sess: empty",
			},
		*/
		{
			configFile:         "testdata/testconfig",
			expectedSess:       "profile",
			awsRegion:          "",
			awsProfile:         "profile",
			awsAccessKeyId:     "access_key_id",
			awsSecretAccessKey: "secret_access_key",
			awsSessionToken:    "session_token",
			description:        "[1] expecting profile sess: no region so it's not static",
		},
		/*
			{
				configFile:         "testdata/testconfig",
				expectedSess:       "default",
				awsRegion:          "region",
				awsProfile:         "",
				awsAccessKeyId:     "access_key_id",
				awsSecretAccessKey: "",
				awsSessionToken:    "session_token",
				description:        "[2] expecting default sess: an AccessKeyId but no accompanying SecretAccessKey, no profile",
			},
			{
				configFile:         "testdata/testconfig",
				expectedSess:       "default",
				awsRegion:          "region",
				awsProfile:         "",
				awsAccessKeyId:     "",
				awsSecretAccessKey: "secret_access_key",
				awsSessionToken:    "session_token",
				description:        "[3] expecting default sess: no AccessKeyId but non-empty SecretAccessKey",
			},
			{
				configFile:         "testdata/testconfig",
				expectedSess:       "profile",
				awsRegion:          "region",
				awsProfile:         "profile",
				awsAccessKeyId:     "",
				awsSecretAccessKey: "",
				awsSessionToken:    "",
				description:        "[4] expecting profile sess: no static creds",
			},
			{
				configFile:         "testdata/testconfig",
				expectedSess:       "profile",
				awsRegion:          "region",
				awsProfile:         "profile",
				awsAccessKeyId:     "",
				awsSecretAccessKey: "secret_access_key",
				awsSessionToken:    "",
				description:        "[5] expecting profile sess: combination of static creds is invalid",
			},
		*/
		{
			configFile:         "testdata/testconfig",
			expectedSess:       "static",
			awsRegion:          "region",
			awsProfile:         "profile",
			awsAccessKeyId:     "access_key_id",
			awsSecretAccessKey: "secret_access_key",
			awsSessionToken:    "session_token",
			description:        "[6] expecting static sess: all options non empty",
		},
		{
			configFile:         "testdata/testconfig",
			expectedSess:       "static",
			awsRegion:          "region",
			awsProfile:         "profile",
			awsAccessKeyId:     "access_key_id",
			awsSecretAccessKey: "secret_access_key",
			awsSessionToken:    "",
			description:        "[7] expecting static sess: session token optional",
		},
		{
			configFile:         "",
			expectedSess:       "",
			awsRegion:          "",
			awsProfile:         "profile",
			awsAccessKeyId:     "",
			awsSecretAccessKey: "",
			awsSessionToken:    "",
			description:        "[8] expecting error: cannot start a profile session without a config file",
		},
		/*
			{
				configFile:         "",
				expectedSess:       "default",
				awsRegion:          "",
				awsProfile:         "",
				awsAccessKeyId:     "",
				awsSecretAccessKey: "",
				awsSessionToken:    "",
				description:        "[9] attempt to start session based on default locations/profiles/etc",
			},
		*/
	}

	for index, tc := range testCases {
		t.Run(fmt.Sprintf("%d--%s", index, tc.description), func(t *testing.T) {

			setAndTestEnv(t, tc.configFile)

			availRegions := []string{"region", "another-region"}

			_, tag, err := ChooseSessionType(tc.awsRegion, tc.awsProfile, CreateStaticCreds(tc.awsAccessKeyId, tc.awsSecretAccessKey, tc.awsSessionToken), availRegions)
			if len(tc.configFile) != 0 {
				assert.NoError(t, err)
			} else {
				if len(tc.awsProfile) != 0 {
					assert.Errorf(t, err, "expected error because there is no config file")
				}
			}
			assert.Equal(t, tc.expectedSess, tag)

		})
	}
}

func setAndTestEnv(t *testing.T, configFile string) {
	confErr := os.Setenv("AWS_CONFIG_FILE", filepath.Join(configFile))
	if confErr != nil {
		t.Fatalf("os.Setenv config file error: %s", confErr)
	}
	credErr := os.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(configFile))
	if credErr != nil {
		t.Fatalf("os.Setenv cred file error: %s", credErr)
	}
	assert.Equal(t, os.Getenv("AWS_CONFIG_FILE"), configFile)
	assert.Equal(t, os.Getenv("AWS_SHARED_CREDENTIALS_FILE"), configFile)

}
