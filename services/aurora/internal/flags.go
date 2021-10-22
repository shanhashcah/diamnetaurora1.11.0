package aurora

import (
	"go/types"
	stdLog "log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/diamnet/go/services/aurora/internal/db2/schema"
	apkg "github.com/diamnet/go/support/app"
	support "github.com/diamnet/go/support/config"
	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/log"
	"github.com/stellar/throttled"
)

const (
	// DiamnetCoreDBURLFlagName is the command line flag for configuring the postgres Diamnet Core URL
	DiamnetCoreDBURLFlagName = "diamnet-core-db-url"
	// DiamnetCoreDBURLFlagName is the command line flag for configuring the URL fore Diamnet Core HTTP endpoint
	DiamnetCoreURLFlagName = "diamnet-core-url"
)

// validateBothOrNeither ensures that both options are provided, if either is provided.
func validateBothOrNeither(option1, option2 string) {
	arg1, arg2 := viper.GetString(option1), viper.GetString(option2)
	if arg1 != "" && arg2 == "" {
		stdLog.Fatalf("Invalid config: %s = %s, but corresponding option %s is not configured", option1, arg1, option2)
	}
	if arg1 == "" && arg2 != "" {
		stdLog.Fatalf("Invalid config: %s = %s, but corresponding option %s is not configured", option2, arg2, option1)
	}
}

func applyMigrations(config Config) {
	dbConn, err := db.Open("postgres", config.DatabaseURL)
	if err != nil {
		stdLog.Fatalf("could not connect to aurora db: %v", err)
	}
	defer dbConn.Close()

	numMigrations, err := schema.Migrate(dbConn.DB.DB, schema.MigrateUp, 0)
	if err != nil {
		stdLog.Fatalf("could not apply migrations: %v", err)
	}
	if numMigrations > 0 {
		stdLog.Printf("successfully applied %v aurora migrations\n", numMigrations)
	}
}

// checkMigrations looks for necessary database migrations and fails with a descriptive error if migrations are needed.
func checkMigrations(config Config) {
	migrationsToApplyUp := schema.GetMigrationsUp(config.DatabaseURL)
	if len(migrationsToApplyUp) > 0 {
		stdLog.Printf(`There are %v migrations to apply in the "up" direction.`, len(migrationsToApplyUp))
		stdLog.Printf("The necessary migrations are: %v", migrationsToApplyUp)
		stdLog.Printf("A database migration is required to run this version (%v) of Aurora. Run \"aurora db migrate up\" to update your DB. Consult the Changelog (https://github.com/diamnet/go/blob/master/services/aurora/CHANGELOG.md) for more information.", apkg.Version())
		os.Exit(1)
	}

	nMigrationsDown := schema.GetNumMigrationsDown(config.DatabaseURL)
	if nMigrationsDown > 0 {
		stdLog.Printf("A database migration DOWN to an earlier version of the schema is required to run this version (%v) of Aurora. Consult the Changelog (https://github.com/diamnet/go/blob/master/services/aurora/CHANGELOG.md) for more information.", apkg.Version())
		stdLog.Printf("In order to migrate the database DOWN, using the HIGHEST version number of Aurora you have installed (not this binary), run \"aurora db migrate down %v\".", nMigrationsDown)
		os.Exit(1)
	}
}

// Flags returns a Config instance and a list of commandline flags which modify the Config instance
func Flags() (*Config, support.ConfigOptions) {
	config := &Config{}
	var dbURLConfigOption = &support.ConfigOption{
		Name:      "db-url",
		EnvVar:    "DATABASE_URL",
		ConfigKey: &config.DatabaseURL,
		OptType:   types.String,
		Required:  true,
		Usage:     "aurora postgres database to connect with",
	}

	// flags defines the complete flag configuration for aurora.
	// Add a new entry here to connect a new field in the aurora.Config struct
	var flags = support.ConfigOptions{
		dbURLConfigOption,
		&support.ConfigOption{
			Name:        "diamnet-core-binary-path",
			OptType:     types.String,
			FlagDefault: "",
			Required:    false,
			Usage:       "path to diamnet core binary (--remote-captive-core-url has higher precedence)",
			ConfigKey:   &config.DiamnetCoreBinaryPath,
		},
		&support.ConfigOption{
			Name:        "remote-captive-core-url",
			OptType:     types.String,
			FlagDefault: "",
			Required:    false,
			Usage:       "url to access the remote captive core server",
			ConfigKey:   &config.RemoteCaptiveCoreURL,
		},
		&support.ConfigOption{
			Name:        "diamnet-core-config-path",
			OptType:     types.String,
			FlagDefault: "",
			Required:    false,
			Usage:       "path to diamnet core config file",
			ConfigKey:   &config.DiamnetCoreConfigPath,
		},
		&support.ConfigOption{
			Name:        "enable-captive-core-ingestion",
			OptType:     types.Bool,
			FlagDefault: false,
			Required:    false,
			Usage:       "[experimental flag!] causes Aurora to ingest from a Diamnet Core subprocess instead of a persistent Diamnet Core database",
			ConfigKey:   &config.EnableCaptiveCoreIngestion,
		},
		&support.ConfigOption{
			Name:      DiamnetCoreDBURLFlagName,
			EnvVar:    "DIAMNET_CORE_DATABASE_URL",
			ConfigKey: &config.DiamnetCoreDatabaseURL,
			OptType:   types.String,
			Required:  false,
			Usage:     "diamnet-core postgres database to connect with",
		},
		&support.ConfigOption{
			Name:      DiamnetCoreURLFlagName,
			ConfigKey: &config.DiamnetCoreURL,
			OptType:   types.String,
			Usage:     "diamnet-core to connect with (for http commands)",
		},
		&support.ConfigOption{
			Name:        "history-archive-urls",
			ConfigKey:   &config.HistoryArchiveURLs,
			OptType:     types.String,
			Required:    false,
			FlagDefault: "",
			CustomSetValue: func(co *support.ConfigOption) {
				stringOfUrls := viper.GetString(co.Name)
				urlStrings := strings.Split(stringOfUrls, ",")

				*(co.ConfigKey.(*[]string)) = urlStrings
			},
			Usage: "comma-separated list of diamnet history archives to connect with",
		},
		&support.ConfigOption{
			Name:        "port",
			ConfigKey:   &config.Port,
			OptType:     types.Uint,
			FlagDefault: uint(8000),
			Usage:       "tcp port to listen on for http requests",
		},
		&support.ConfigOption{
			Name:        "admin-port",
			ConfigKey:   &config.AdminPort,
			OptType:     types.Uint,
			FlagDefault: uint(0),
			Usage:       "WARNING: this should not be accessible from the Internet and does not use TLS, tcp port to listen on for admin http requests, 0 (default) disables the admin server",
		},
		&support.ConfigOption{
			Name:        "max-db-connections",
			ConfigKey:   &config.MaxDBConnections,
			OptType:     types.Int,
			FlagDefault: 0,
			Usage:       "when set has a priority over aurora-db-max-open-connections, aurora-db-max-idle-connections. max aurora database open connections may need to be increased when responses are slow but DB CPU is normal",
		},
		&support.ConfigOption{
			Name:        "aurora-db-max-open-connections",
			ConfigKey:   &config.AuroraDBMaxOpenConnections,
			OptType:     types.Int,
			FlagDefault: 20,
			Usage:       "max aurora database open connections. may need to be increased when responses are slow but DB CPU is normal",
		},
		&support.ConfigOption{
			Name:        "aurora-db-max-idle-connections",
			ConfigKey:   &config.AuroraDBMaxIdleConnections,
			OptType:     types.Int,
			FlagDefault: 20,
			Usage:       "max aurora database idle connections. may need to be set to the same value as aurora-db-max-open-connections when responses are slow and DB CPU is normal, because it may indicate that a lot of time is spent closing/opening idle connections. This can happen in case of high variance in number of requests. must be equal or lower than max open connections",
		},
		&support.ConfigOption{
			Name:           "sse-update-frequency",
			ConfigKey:      &config.SSEUpdateFrequency,
			OptType:        types.Int,
			FlagDefault:    5,
			CustomSetValue: support.SetDuration,
			Usage:          "defines how often streams should check if there's a new ledger (in seconds), may need to increase in case of big number of streams",
		},
		&support.ConfigOption{
			Name:           "connection-timeout",
			ConfigKey:      &config.ConnectionTimeout,
			OptType:        types.Int,
			FlagDefault:    55,
			CustomSetValue: support.SetDuration,
			Usage:          "defines the timeout of connection after which 504 response will be sent or stream will be closed, if Aurora is behind a load balancer with idle connection timeout, this should be set to a few seconds less that idle timeout, does not apply to POST /transactions",
		},
		&support.ConfigOption{
			Name:        "per-hour-rate-limit",
			ConfigKey:   &config.RateQuota,
			OptType:     types.Int,
			FlagDefault: 3600,
			CustomSetValue: func(co *support.ConfigOption) {
				var rateLimit *throttled.RateQuota = nil
				perHourRateLimit := viper.GetInt(co.Name)
				if perHourRateLimit != 0 {
					rateLimit = &throttled.RateQuota{
						MaxRate:  throttled.PerHour(perHourRateLimit),
						MaxBurst: 100,
					}
					*(co.ConfigKey.(**throttled.RateQuota)) = rateLimit
				}
			},
			Usage: "max count of requests allowed in a one hour period, by remote ip address",
		},
		&support.ConfigOption{ // Action needed in release: aurora-v2.0.0
			// remove deprecated flag
			Name:    "rate-limit-redis-key",
			OptType: types.String,
			Usage:   "deprecated, do not use",
		},
		&support.ConfigOption{ // Action needed in release: aurora-v2.0.0
			// remove deprecated flag
			Name:    "redis-url",
			OptType: types.String,
			Usage:   "deprecated, do not use",
		},
		&support.ConfigOption{
			Name:           "friendbot-url",
			ConfigKey:      &config.FriendbotURL,
			OptType:        types.String,
			CustomSetValue: support.SetURL,
			Usage:          "friendbot service to redirect to",
		},
		&support.ConfigOption{
			Name:        "log-level",
			ConfigKey:   &config.LogLevel,
			OptType:     types.String,
			FlagDefault: "info",
			CustomSetValue: func(co *support.ConfigOption) {
				ll, err := logrus.ParseLevel(viper.GetString(co.Name))
				if err != nil {
					stdLog.Fatalf("Could not parse log-level: %v", viper.GetString(co.Name))
				}
				*(co.ConfigKey.(*logrus.Level)) = ll
			},
			Usage: "minimum log severity (debug, info, warn, error) to log",
		},
		&support.ConfigOption{
			Name:      "log-file",
			ConfigKey: &config.LogFile,
			OptType:   types.String,
			Usage:     "name of the file where logs will be saved (leave empty to send logs to stdout)",
		},
		&support.ConfigOption{
			Name:        "max-path-length",
			ConfigKey:   &config.MaxPathLength,
			OptType:     types.Uint,
			FlagDefault: uint(3),
			Usage:       "the maximum number of assets on the path in `/paths` endpoint, warning: increasing this value will increase /paths response time",
		},
		&support.ConfigOption{
			Name:      "network-passphrase",
			ConfigKey: &config.NetworkPassphrase,
			OptType:   types.String,
			Required:  true,
			Usage:     "Override the network passphrase",
		},
		&support.ConfigOption{
			Name:      "sentry-dsn",
			ConfigKey: &config.SentryDSN,
			OptType:   types.String,
			Usage:     "Sentry URL to which panics and errors should be reported",
		},
		&support.ConfigOption{
			Name:      "loggly-token",
			ConfigKey: &config.LogglyToken,
			OptType:   types.String,
			Usage:     "Loggly token, used to configure log forwarding to loggly",
		},
		&support.ConfigOption{
			Name:        "loggly-tag",
			ConfigKey:   &config.LogglyTag,
			OptType:     types.String,
			FlagDefault: "aurora",
			Usage:       "Tag to be added to every loggly log event",
		},
		&support.ConfigOption{
			Name:      "tls-cert",
			ConfigKey: &config.TLSCert,
			OptType:   types.String,
			Usage:     "TLS certificate file to use for securing connections to aurora",
		},
		&support.ConfigOption{
			Name:      "tls-key",
			ConfigKey: &config.TLSKey,
			OptType:   types.String,
			Usage:     "TLS private key file to use for securing connections to aurora",
		},
		&support.ConfigOption{
			Name:        "ingest",
			ConfigKey:   &config.Ingest,
			OptType:     types.Bool,
			FlagDefault: false,
			Usage:       "causes this aurora process to ingest data from diamnet-core into aurora's db",
		},
		&support.ConfigOption{
			Name:        "cursor-name",
			EnvVar:      "CURSOR_NAME",
			ConfigKey:   &config.CursorName,
			OptType:     types.String,
			FlagDefault: "HORIZON",
			Usage:       "ingestor cursor used by aurora to ingest from diamnet core. must be uppercase and unique for each aurora instance ingesting from that core instance.",
		},
		&support.ConfigOption{
			Name:        "history-retention-count",
			ConfigKey:   &config.HistoryRetentionCount,
			OptType:     types.Uint,
			FlagDefault: uint(0),
			Usage:       "the minimum number of ledgers to maintain within aurora's history tables.  0 signifies an unlimited number of ledgers will be retained",
		},
		&support.ConfigOption{
			Name:        "history-stale-threshold",
			ConfigKey:   &config.StaleThreshold,
			OptType:     types.Uint,
			FlagDefault: uint(0),
			Usage:       "the maximum number of ledgers the history db is allowed to be out of date from the connected diamnet-core db before aurora considers history stale",
		},
		&support.ConfigOption{
			Name:        "skip-cursor-update",
			ConfigKey:   &config.SkipCursorUpdate,
			OptType:     types.Bool,
			FlagDefault: false,
			Usage:       "causes the ingester to skip reporting the last imported ledger state to diamnet-core",
		},
		&support.ConfigOption{
			Name:        "ingest-disable-state-verification",
			ConfigKey:   &config.IngestDisableStateVerification,
			OptType:     types.Bool,
			FlagDefault: false,
			Usage:       "ingestion system runs a verification routing to compare state in local database with history buckets, this can be disabled however it's not recommended",
		},
		&support.ConfigOption{
			Name:        "apply-migrations",
			ConfigKey:   &config.ApplyMigrations,
			OptType:     types.Bool,
			FlagDefault: false,
			Required:    false,
			Usage:       "applies pending migrations before starting aurora",
		},
	}

	return config, flags
}

// NewAppFromFlags constructs a new Aurora App from the given command line flags
func NewAppFromFlags(config *Config, flags support.ConfigOptions) *App {
	ApplyFlags(config, flags)
	// Validate app-specific arguments
	if config.DiamnetCoreURL == "" {
		log.Fatalf("flag --%s cannot be empty", DiamnetCoreURLFlagName)
	}
	if config.Ingest && !config.EnableCaptiveCoreIngestion && config.DiamnetCoreDatabaseURL == "" {
		log.Fatalf("flag --%s cannot be empty", DiamnetCoreDBURLFlagName)
	}
	app, err := NewApp(*config)
	if err != nil {
		log.Fatalf("cannot initialize app: %s", err)
	}
	return app
}

// ApplyFlags applies the command line flags on the given Config instance
func ApplyFlags(config *Config, flags support.ConfigOptions) {
	// Verify required options and load the config struct
	flags.Require()
	flags.SetValues()

	if config.ApplyMigrations {
		applyMigrations(*config)
	}

	// Migrations should be checked as early as possible
	checkMigrations(*config)

	// Validate options that should be provided together
	validateBothOrNeither("tls-cert", "tls-key")

	// config.HistoryArchiveURLs contains a single empty value when empty so using
	// viper.GetString is easier.
	if config.Ingest && viper.GetString("history-archive-urls") == "" {
		stdLog.Fatalf("--history-archive-urls must be set when --ingest is set")
	}

	if config.EnableCaptiveCoreIngestion {
		binaryPath := viper.GetString("diamnet-core-binary-path")
		remoteURL := viper.GetString("remote-captive-core-url")
		if binaryPath == "" && remoteURL == "" {
			stdLog.Fatalf("Invalid config: captive core requires that either --diamnet-core-binary-path or --remote-captive-core-url is set")
		}
	}
	if config.Ingest {
		// When running live ingestion a config file is required too
		validateBothOrNeither("diamnet-core-binary-path", "diamnet-core-config-path")
	}

	// Configure log file
	if config.LogFile != "" {
		logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			log.DefaultLogger.Logger.Out = logFile
		} else {
			stdLog.Fatalf("Failed to open file to log: %s", err)
		}
	}

	// Configure log level
	log.DefaultLogger.Logger.SetLevel(config.LogLevel)

	// Configure DB params. When config.MaxDBConnections is set, set other
	// DB params to that value for backward compatibility.
	if config.MaxDBConnections != 0 {
		config.AuroraDBMaxOpenConnections = config.MaxDBConnections
		config.AuroraDBMaxIdleConnections = config.MaxDBConnections
	}
}
