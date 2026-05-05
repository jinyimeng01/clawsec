package constants

const (
	AppName   = "ClawSec"
	AppSlug   = "clawsec"
	Version   = "0.1.0-alpha"
	BuildTime = "dev"
	GitCommit = "dev"
	GitBranch = "dev"
	GoVersion = "1.22"

	// Default paths
	DefaultConfigDir  = ".clawsec"
	DefaultConfigFile = "config.yaml"
	DefaultLogFile    = "clawsec.log"
	DefaultAuditFile  = "audit.log"

	// Default network settings
	DefaultTimeout      = 5
	DefaultRetries      = 1
	DefaultRateLimit    = 150
	DefaultMaxRedirects = 10
	DefaultUserAgent    = "ClawSec/" + Version

	// Scan defaults
	DefaultPortRange   = "top100"
	DefaultThreads     = 50
	DefaultScanTimeout = 3
)

// UserAgentList for rotation
var UserAgentList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
}
