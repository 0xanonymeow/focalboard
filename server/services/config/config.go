package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	DefaultServerRoot = "http://localhost:8000"
	DefaultPort       = 8000
	DBPingAttempts    = 5
)

type AmazonS3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PathPrefix      string
	Region          string
	Endpoint        string
	SSL             bool
	SignV2          bool
	SSE             bool
	Trace           bool
	Timeout         int64
}

// Configuration is the app configuration stored in a json file.
type Configuration struct {
	ServerRoot               string            `json:"serverRoot" mapstructure:"serverRoot"`
	Port                     int               `json:"port" mapstructure:"port"`
	DBType                   string            `json:"dbtype" mapstructure:"dbtype"`
	DBConfigString           string            `json:"dbconfig" mapstructure:"dbconfig"`
	DBPingAttempts           int               `json:"dbpingattempts" mapstructure:"dbpingattempts"`
	DBTablePrefix            string            `json:"dbtableprefix" mapstructure:"dbtableprefix"`
	UseSSL                   bool              `json:"useSSL" mapstructure:"useSSL"`
	SecureCookie             bool              `json:"secureCookie" mapstructure:"secureCookie"`
	WebPath                  string            `json:"webpath" mapstructure:"webpath"`
	FilesDriver              string            `json:"filesdriver" mapstructure:"filesdriver"`
	FilesS3Config            AmazonS3Config    `json:"filess3config" mapstructure:"filess3config"`
	FilesPath                string            `json:"filespath" mapstructure:"filespath"`
	MaxFileSize              int64             `json:"maxfilesize" mapstructure:"maxfilesize"`
	Telemetry                bool              `json:"telemetry" mapstructure:"telemetry"`
	TelemetryID              string            `json:"telemetryid" mapstructure:"telemetryid"`
	PrometheusAddress        string            `json:"prometheusaddress" mapstructure:"prometheusaddress"`
	WebhookUpdate            []string          `json:"webhook_update" mapstructure:"webhook_update"`
	Secret                   string            `json:"secret" mapstructure:"secret"`
	SessionExpireTime        int64             `json:"session_expire_time" mapstructure:"session_expire_time"`
	SessionRefreshTime       int64             `json:"session_refresh_time" mapstructure:"session_refresh_time"`
	LocalOnly                bool              `json:"localonly" mapstructure:"localonly"`
	EnableLocalMode          bool              `json:"enableLocalMode" mapstructure:"enableLocalMode"`
	LocalModeSocketLocation  string            `json:"localModeSocketLocation" mapstructure:"localModeSocketLocation"`
	EnablePublicSharedBoards bool              `json:"enablePublicSharedBoards" mapstructure:"enablePublicSharedBoards"`
	FeatureFlags             map[string]string `json:"featureFlags" mapstructure:"featureFlags"`
	EnableDataRetention      bool              `json:"enable_data_retention" mapstructure:"enable_data_retention"`
	DataRetentionDays        int               `json:"data_retention_days" mapstructure:"data_retention_days"`
	TeammateNameDisplay      string            `json:"teammate_name_display" mapstructure:"teammateNameDisplay"`
	ShowEmailAddress         bool              `json:"show_email_address" mapstructure:"showEmailAddress"`
	ShowFullName             bool              `json:"show_full_name" mapstructure:"showFullName"`

	AuthMode string `json:"authMode" mapstructure:"authMode"`

	LoggingCfgFile string `json:"logging_cfg_file" mapstructure:"logging_cfg_file"`
	LoggingCfgJSON string `json:"logging_cfg_json" mapstructure:"logging_cfg_json"`

	AuditCfgFile string `json:"audit_cfg_file" mapstructure:"audit_cfg_file"`
	AuditCfgJSON string `json:"audit_cfg_json" mapstructure:"audit_cfg_json"`

	NotifyFreqCardSeconds  int `json:"notify_freq_card_seconds" mapstructure:"notify_freq_card_seconds"`
	NotifyFreqBoardSeconds int `json:"notify_freq_board_seconds" mapstructure:"notify_freq_board_seconds"`
}

// ReadConfigFile read the configuration from the filesystem.
func ReadConfigFile(configFilePath string) (*Configuration, error) {
	// Step 0: Load .env file if it exists (for development convenience)
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't fail if it doesn't exist
		log.Printf(".env file not found or invalid: %v. Using system environment variables.", err)
	}

	// Step 1: Set up defaults
	setDefaults()

	// Step 2: Read config file FIRST (if it exists)
	if configFilePath == "" {
		viper.SetConfigFile("./config.json")
	} else {
		viper.SetConfigFile(configFilePath)
	}

	err := viper.ReadInConfig()
	if err != nil {
		// Config file is optional if environment variables are provided
		log.Printf("Config file not found or invalid: %v. Using defaults and environment variables.", err)
	}

	// Step 3: Set up environment variable support AFTER config file
	// This ensures environment variables override config file values
	viper.SetEnvPrefix("focalboard")
	viper.AutomaticEnv()
	bindEnvironmentVariables()

	// Step 4: Apply manual environment variable overrides for complex structures before unmarshaling
	applyEnvironmentOverridesPre()

	// Step 5: Unmarshal configuration
	configuration := Configuration{}
	err = viper.Unmarshal(&configuration)
	if err != nil {
		return nil, err
	}

	// Step 6: Apply post-unmarshal environment variable overrides
	applyEnvironmentOverridesPost(&configuration)

	log.Println("readConfigFile")
	log.Printf("%+v", removeSecurityData(configuration))

	return &configuration, nil
}

// setDefaults sets all default configuration values using mapstructure keys
func setDefaults() {
	viper.SetDefault("serverRoot", DefaultServerRoot)
	viper.SetDefault("port", DefaultPort)
	viper.SetDefault("dbtype", "sqlite3")
	viper.SetDefault("dbconfig", "./focalboard.db")
	viper.SetDefault("dbpingattempts", DBPingAttempts)
	viper.SetDefault("dbtableprefix", "")
	viper.SetDefault("useSSL", false)
	viper.SetDefault("secureCookie", false)
	viper.SetDefault("webpath", "./pack")
	viper.SetDefault("filesdriver", "local")
	viper.SetDefault("filespath", "./files")
	viper.SetDefault("maxfilesize", int64(0))
	viper.SetDefault("telemetry", true)
	viper.SetDefault("telemetryid", "")
	viper.SetDefault("prometheusaddress", "")
	viper.SetDefault("webhook_update", []string{})
	viper.SetDefault("secret", "")
	viper.SetDefault("session_expire_time", int64(60*60*24*30)) // 30 days session lifetime
	viper.SetDefault("session_refresh_time", int64(60*60*5))    // 5 minutes session refresh
	viper.SetDefault("localonly", false)
	viper.SetDefault("enableLocalMode", false)
	viper.SetDefault("localModeSocketLocation", "/var/tmp/focalboard_local.socket")
	viper.SetDefault("enablePublicSharedBoards", false)
	viper.SetDefault("featureFlags", map[string]string{})
	viper.SetDefault("enable_data_retention", false)
	viper.SetDefault("data_retention_days", 365) // 1 year is default
	viper.SetDefault("teammateNameDisplay", "username")
	viper.SetDefault("showEmailAddress", false)
	viper.SetDefault("showFullName", false)
	viper.SetDefault("authMode", "native")
	viper.SetDefault("logging_cfg_file", "")
	viper.SetDefault("logging_cfg_json", "")
	viper.SetDefault("audit_cfg_file", "")
	viper.SetDefault("audit_cfg_json", "")
	viper.SetDefault("notify_freq_card_seconds", 120)    // 2 minutes after last card edit
	viper.SetDefault("notify_freq_board_seconds", 86400) // 1 day after last card edit
}

// bindEnvironmentVariables binds all configuration keys to environment variables using mapstructure keys
func bindEnvironmentVariables() {
	// Main configuration fields (using mapstructure keys)
	viper.BindEnv("serverRoot", "FOCALBOARD_SERVERROOT")
	viper.BindEnv("port", "FOCALBOARD_PORT")
	viper.BindEnv("dbtype", "FOCALBOARD_DBTYPE")
	viper.BindEnv("dbconfig", "FOCALBOARD_DBCONFIG")
	viper.BindEnv("dbpingattempts", "FOCALBOARD_DBPINGATTEMPTS")
	viper.BindEnv("dbtableprefix", "FOCALBOARD_DBTABLEPREFIX")
	viper.BindEnv("useSSL", "FOCALBOARD_USESSL")
	viper.BindEnv("secureCookie", "FOCALBOARD_SECURECOOKIE")
	viper.BindEnv("webpath", "FOCALBOARD_WEBPATH")
	viper.BindEnv("filesdriver", "FOCALBOARD_FILESDRIVER")
	viper.BindEnv("filespath", "FOCALBOARD_FILESPATH")
	viper.BindEnv("maxfilesize", "FOCALBOARD_MAXFILESIZE")
	viper.BindEnv("telemetry", "FOCALBOARD_TELEMETRY")
	viper.BindEnv("telemetryid", "FOCALBOARD_TELEMETRYID")
	viper.BindEnv("prometheusaddress", "FOCALBOARD_PROMETHEUSADDRESS")
	viper.BindEnv("secret", "FOCALBOARD_SECRET")
	viper.BindEnv("session_expire_time", "FOCALBOARD_SESSIONEXPIRETIME")
	viper.BindEnv("session_refresh_time", "FOCALBOARD_SESSIONREFRESHTIME")
	viper.BindEnv("localonly", "FOCALBOARD_LOCALONLY")
	viper.BindEnv("enableLocalMode", "FOCALBOARD_ENABLELOCALMODE")
	viper.BindEnv("localModeSocketLocation", "FOCALBOARD_LOCALMODESOCKETLOCATION")
	viper.BindEnv("enablePublicSharedBoards", "FOCALBOARD_ENABLEPUBLICSHAREDBOARDS")
	viper.BindEnv("enable_data_retention", "FOCALBOARD_ENABLEDATARETENTION")
	viper.BindEnv("data_retention_days", "FOCALBOARD_DATARETENTIONDAYS")
	viper.BindEnv("teammateNameDisplay", "FOCALBOARD_TEAMMATENAMEDISPLAY")
	viper.BindEnv("showEmailAddress", "FOCALBOARD_SHOWEMAILADDRESS")
	viper.BindEnv("showFullName", "FOCALBOARD_SHOWFULLNAME")
	viper.BindEnv("authMode", "FOCALBOARD_AUTHMODE")
	viper.BindEnv("logging_cfg_file", "FOCALBOARD_LOGGINGCFGFILE")
	viper.BindEnv("logging_cfg_json", "FOCALBOARD_LOGGINGCFGJSON")
	viper.BindEnv("audit_cfg_file", "FOCALBOARD_AUDITCFGFILE")
	viper.BindEnv("audit_cfg_json", "FOCALBOARD_AUDITCFGJSON")
	viper.BindEnv("notify_freq_card_seconds", "FOCALBOARD_NOTIFYFREQCARDSECONDS")
	viper.BindEnv("notify_freq_board_seconds", "FOCALBOARD_NOTIFYFREQBOARDSECONDS")

	// S3 Configuration fields (using mapstructure keys)
	viper.BindEnv("filess3config.accesskeyid", "FOCALBOARD_FILESS3CONFIG_ACCESSKEYID")
	viper.BindEnv("filess3config.secretaccesskey", "FOCALBOARD_FILESS3CONFIG_SECRETACCESSKEY")
	viper.BindEnv("filess3config.bucket", "FOCALBOARD_FILESS3CONFIG_BUCKET")
	viper.BindEnv("filess3config.pathprefix", "FOCALBOARD_FILESS3CONFIG_PATHPREFIX")
	viper.BindEnv("filess3config.region", "FOCALBOARD_FILESS3CONFIG_REGION")
	viper.BindEnv("filess3config.endpoint", "FOCALBOARD_FILESS3CONFIG_ENDPOINT")
	viper.BindEnv("filess3config.ssl", "FOCALBOARD_FILESS3CONFIG_SSL")
	viper.BindEnv("filess3config.signv2", "FOCALBOARD_FILESS3CONFIG_SIGNV2")
	viper.BindEnv("filess3config.sse", "FOCALBOARD_FILESS3CONFIG_SSE")
	viper.BindEnv("filess3config.trace", "FOCALBOARD_FILESS3CONFIG_TRACE")
	viper.BindEnv("filess3config.timeout", "FOCALBOARD_FILESS3CONFIG_TIMEOUT")
}

// applyEnvironmentOverridesPre applies environment variable overrides before viper unmarshaling
func applyEnvironmentOverridesPre() {
	// Handle FeatureFlags map - set in viper before unmarshaling
	if featureFlags := os.Getenv("FOCALBOARD_FEATUREFLAGS"); featureFlags != "" {
		if featureFlags == "null" || featureFlags == "" {
			viper.Set("featureFlags", map[string]string{})
		} else {
			viper.Set("featureFlags", parseFeatureFlags(featureFlags))
		}
	}

	// Handle WebhookUpdate array - set in viper before unmarshaling
	if webhooks := os.Getenv("FOCALBOARD_WEBHOOKUPDATE"); webhooks != "" {
		if webhooks == "null" || webhooks == "" {
			viper.Set("webhook_update", []string{})
		} else {
			webhookList := strings.Split(webhooks, ",")
			// Trim whitespace from each webhook URL
			for i, webhook := range webhookList {
				webhookList[i] = strings.TrimSpace(webhook)
			}
			viper.Set("webhook_update", webhookList)
		}
	}
}

// applyEnvironmentOverridesPost applies environment variable overrides after viper unmarshaling
func applyEnvironmentOverridesPost(config *Configuration) {
	// Handle DATABASE_URL (common in cloud deployments like Render)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		config.DBConfigString = databaseURL
		// Auto-detect database type from URL
		if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
			config.DBType = "postgres"
		} else if strings.HasPrefix(databaseURL, "mysql://") {
			config.DBType = "mysql"
		}
	}

	// Note: Other FOCALBOARD_ environment variables are now handled automatically by Viper
	// since we set up environment variable support AFTER reading the config file
}

// parseFeatureFlags parses feature flags from environment variable format
// Expected format: "flag1:value1,flag2:value2"
func parseFeatureFlags(featureFlagsStr string) map[string]string {
	flags := make(map[string]string)
	pairs := strings.Split(featureFlagsStr, ",")
	
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			flags[key] = value
		} else {
			// If no colon, treat as boolean flag set to "true"
			flags[pair] = "true"
		}
	}
	
	return flags
}

func removeSecurityData(config Configuration) Configuration {
	clean := config
	return clean
}
