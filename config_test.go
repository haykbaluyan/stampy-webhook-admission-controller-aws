package main

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSuiteConfigReader(t *testing.T) {

	testCases := []struct {
		name           string
		args           []string
		expectedConfig *Config
		expectedError  string
	}{
		{
			name:           "MissingBucket",
			args:           []string{"x", "-region=test", "-tlsCertdir=", "-tlsPairName="},
			expectedConfig: nil,
			expectedError:  "invalid bucket: empty",
		},
		{
			name:           "MissingRegion",
			args:           []string{"x", "-bucket=test", "-tlsCertdir=", "-tlsPairName="},
			expectedConfig: nil,
			expectedError:  "invalid region: empty",
		},
		{
			name: "All",
			args: []string{"x", "-region=test_region", "-bucket=test_bucket", "-tlsCertdir=", "-tlsPairName="},
			expectedConfig: &Config{
				cert:     ".crt",
				key:      ".key",
				logLevel: logrus.DebugLevel,
				port:     443,
				region:   "test_region",
				bucket:   "test_bucket",
			},
			expectedError: "",
		},
		{
			name: "LogLevel_Info",
			args: []string{"x", "-region=test_region", "-bucket=test_bucket", "-log-level=info", "-tlsCertdir=", "-tlsPairName="},
			expectedConfig: &Config{
				cert:     ".crt",
				key:      ".key",
				port:     443,
				region:   "test_region",
				bucket:   "test_bucket",
				logLevel: logrus.InfoLevel,
			},
			expectedError: "",
		},
		{
			name: "LogLevel_Error",
			args: []string{"x", "--region=test_region", "-bucket=test_bucket", "-log-level=error", "-tlsCertdir=", "-tlsPairName="},
			expectedConfig: &Config{
				cert:     ".crt",
				key:      ".key",
				port:     443,
				region:   "test_region",
				bucket:   "test_bucket",
				logLevel: logrus.ErrorLevel,
			},
			expectedError: "",
		},
		{
			name: "Port",
			args: []string{"x", "-region=test_region", "-bucket=test_bucket", "-log-level=error", "-port=17772", "-tlsCertdir=", "-tlsPairName="},
			expectedConfig: &Config{
				cert:     ".crt",
				key:      ".key",
				port:     17772,
				region:   "test_region",
				bucket:   "test_bucket",
				logLevel: logrus.ErrorLevel,
			},
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = tc.args

			// Act
			config, err := readConfig()

			// Assert
			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			}

			assert.Equal(t, tc.expectedConfig, config)
		})
	}
}
