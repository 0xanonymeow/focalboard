package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupViper() {
	// Reset viper for clean test state
	viper.Reset()
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	// Test basic environment variable override
	t.Run("Basic environment variable override", func(t *testing.T) {
		cleanupViper()
		
		// Set environment variables
		os.Setenv("FOCALBOARD_PORT", "9000")
		os.Setenv("FOCALBOARD_DBTYPE", "postgres")
		os.Setenv("FOCALBOARD_DBCONFIG", "postgres://user:pass@localhost/db")
		defer func() {
			os.Unsetenv("FOCALBOARD_PORT")
			os.Unsetenv("FOCALBOARD_DBTYPE")
			os.Unsetenv("FOCALBOARD_DBCONFIG")
			cleanupViper()
		}()

		config, err := ReadConfigFile("")
		require.NoError(t, err)

		assert.Equal(t, 9000, config.Port)
		assert.Equal(t, "postgres", config.DBType)
		assert.Equal(t, "postgres://user:pass@localhost/db", config.DBConfigString)
	})

	// Test DATABASE_URL override
	t.Run("DATABASE_URL override", func(t *testing.T) {
		cleanupViper()
		
		os.Setenv("DATABASE_URL", "postgres://user:pass@render.com:5432/mydb")
		defer func() {
			os.Unsetenv("DATABASE_URL")
			cleanupViper()
		}()

		config, err := ReadConfigFile("")
		require.NoError(t, err)

		assert.Equal(t, "postgres", config.DBType)
		assert.Equal(t, "postgres://user:pass@render.com:5432/mydb", config.DBConfigString)
	})

	// Test S3 configuration override
	t.Run("S3 configuration override", func(t *testing.T) {
		cleanupViper()
		
		os.Setenv("FOCALBOARD_FILESS3CONFIG_BUCKET", "my-bucket")
		os.Setenv("FOCALBOARD_FILESS3CONFIG_REGION", "us-west-2")
		os.Setenv("FOCALBOARD_FILESS3CONFIG_ACCESSKEYID", "access-key")
		defer func() {
			os.Unsetenv("FOCALBOARD_FILESS3CONFIG_BUCKET")
			os.Unsetenv("FOCALBOARD_FILESS3CONFIG_REGION")
			os.Unsetenv("FOCALBOARD_FILESS3CONFIG_ACCESSKEYID")
			cleanupViper()
		}()

		config, err := ReadConfigFile("")
		require.NoError(t, err)

		assert.Equal(t, "my-bucket", config.FilesS3Config.Bucket)
		assert.Equal(t, "us-west-2", config.FilesS3Config.Region)
		assert.Equal(t, "access-key", config.FilesS3Config.AccessKeyID)
	})

	// Test webhook array parsing
	t.Run("Webhook array parsing", func(t *testing.T) {
		cleanupViper()
		
		os.Setenv("FOCALBOARD_WEBHOOKUPDATE", "http://webhook1.com,http://webhook2.com, http://webhook3.com ")
		defer func() {
			os.Unsetenv("FOCALBOARD_WEBHOOKUPDATE")
			cleanupViper()
		}()

		config, err := ReadConfigFile("")
		require.NoError(t, err)

		expected := []string{"http://webhook1.com", "http://webhook2.com", "http://webhook3.com"}
		assert.Equal(t, expected, config.WebhookUpdate)
	})

	// Test feature flags parsing
	t.Run("Feature flags parsing", func(t *testing.T) {
		cleanupViper()
		
		os.Setenv("FOCALBOARD_FEATUREFLAGS", "feature1:true,feature2:false,feature3:custom-value")
		defer func() {
			os.Unsetenv("FOCALBOARD_FEATUREFLAGS")
			cleanupViper()
		}()

		config, err := ReadConfigFile("")
		require.NoError(t, err)

		expected := map[string]string{
			"feature1": "true",
			"feature2": "false",
			"feature3": "custom-value",
		}
		assert.Equal(t, expected, config.FeatureFlags)
	})
}

func TestParseFeatureFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "Simple key-value pairs",
			input: "flag1:true,flag2:false",
			expected: map[string]string{
				"flag1": "true",
				"flag2": "false",
			},
		},
		{
			name:  "Boolean flags without values",
			input: "flag1,flag2,flag3",
			expected: map[string]string{
				"flag1": "true",
				"flag2": "true",
				"flag3": "true",
			},
		},
		{
			name:  "Mixed format with spaces",
			input: "flag1:value1, flag2 , flag3:value3 ",
			expected: map[string]string{
				"flag1": "value1",
				"flag2": "true",
				"flag3": "value3",
			},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "Single flag",
			input: "singleflag:singlevalue",
			expected: map[string]string{
				"singleflag": "singlevalue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFeatureFlags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaults(t *testing.T) {
	// Test that defaults are properly set when no config file or env vars exist
	cleanupViper()
	defer cleanupViper()
	
	config, err := ReadConfigFile("")
	require.NoError(t, err)

	assert.Equal(t, DefaultPort, config.Port)
	assert.Equal(t, DefaultServerRoot, config.ServerRoot)
	assert.Equal(t, "sqlite3", config.DBType)
	assert.Equal(t, "./focalboard.db", config.DBConfigString)
	assert.Equal(t, "local", config.FilesDriver)
	assert.Equal(t, "./files", config.FilesPath)
	assert.Equal(t, "./pack", config.WebPath)
	assert.Equal(t, false, config.EnablePublicSharedBoards)
	assert.Equal(t, "native", config.AuthMode)
}