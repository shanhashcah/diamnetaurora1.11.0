package aurora

import (
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stellar/throttled"
)

// Config is the configuration for aurora.  It gets populated by the
// app's main function and is provided to NewApp.
type Config struct {
	DatabaseURL        string
	HistoryArchiveURLs []string
	Port               uint
	AdminPort          uint

	EnableCaptiveCoreIngestion bool
	DiamnetCoreBinaryPath      string
	DiamnetCoreConfigPath      string
	DiamnetCoreDatabaseURL     string
	DiamnetCoreURL             string
	RemoteCaptiveCoreURL       string

	// MaxDBConnections has a priority over all 4 values below.
	MaxDBConnections            int
	AuroraDBMaxOpenConnections int
	AuroraDBMaxIdleConnections int

	SSEUpdateFrequency time.Duration
	ConnectionTimeout  time.Duration
	RateQuota          *throttled.RateQuota
	FriendbotURL       *url.URL
	LogLevel           logrus.Level
	LogFile            string
	// MaxPathLength is the maximum length of the path returned by `/paths` endpoint.
	MaxPathLength     uint
	NetworkPassphrase string
	SentryDSN         string
	LogglyToken       string
	LogglyTag         string
	// TLSCert is a path to a certificate file to use for aurora's TLS config
	TLSCert string
	// TLSKey is the path to a private key file to use for aurora's TLS config
	TLSKey string
	// Ingest toggles whether this aurora instance should run the data ingestion subsystem.
	Ingest bool
	// CursorName is the cursor used for ingesting from diamnet-core.
	// Setting multiple cursors in different Aurora instances allows multiple
	// Auroras to ingest from the same diamnet-core instance without cursor
	// collisions.
	CursorName string
	// HistoryRetentionCount represents the minimum number of ledgers worth of
	// history data to retain in the aurora database. For the purposes of
	// determining a "retention duration", each ledger roughly corresponds to 10
	// seconds of real time.
	HistoryRetentionCount uint
	// StaleThreshold represents the number of ledgers a history database may be
	// out-of-date by before aurora begins to respond with an error to history
	// requests.
	StaleThreshold uint
	// SkipCursorUpdate causes the ingestor to skip reporting the "last imported
	// ledger" state to diamnet-core.
	SkipCursorUpdate bool
	// IngestDisableStateVerification disables state verification
	// `System.verifyState()` when set to `true`.
	IngestDisableStateVerification bool
	// ApplyMigrations will apply pending migrations to the aurora database
	// before starting the aurora service
	ApplyMigrations bool
}
