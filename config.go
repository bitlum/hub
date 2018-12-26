package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"log"

	"github.com/btcsuite/btcutil"
	"github.com/jessevdk/go-flags"
)

const (
	defaultHubPort = "8686"
	defaultHubHost = "localhost"

	defaultGraphQLHost       = "0.0.0.0"
	defaultGraphQLPort       = "3000"
	defaultGraphQLSecurePort = "3443"

	defaultEmulateNetworkPort = "9393"
	defaultEmulateNetworkHost = "localhost"

	defaultLogDirname     = "logs"
	defaultLogLevel       = "info"
	defaultConfigFilename = "hub.conf"
	defaultLogFilename    = "hub.log"

	defaultPrometheusHost     = "0.0.0.0"
	defaultPrometheusPort     = "19999"
	defaultUpdatesLogFileName = "log.protobuf"

	defaultDbPath = "/tmp"
	defaultNet    = "simnet"
)

type graphqlConfig struct {
	ListenHost       string `long:"listenhost" description:"The host on which GraphQL will listen for incoming requests"`
	ListenPort       string `long:"listenport" description:"The port on which GraphQL will listen for incoming requests"`
	SecureListenPort string `long:"securelistenport" description:"The secure port on which GraphQL will listen for incoming requests"`
}

var (
	homeDir           = btcutil.AppDataDir("hubmanager", false)
	defaultConfigFile = filepath.Join(homeDir, defaultConfigFilename)
	defaultLogDir     = filepath.Join(homeDir, defaultLogDirname)
)

// config defines the configuration options for lnd.
//
// See loadConfig for further details regarding the configuration
// loading+parsing process.
type config struct {
	ShowVersion bool `long:"version" description:"Display version information and exit"`

	UpdateLogFile string           `long:"updateslog" description:"Path to log file in which manager will direct lightning node network updates output"`
	LND           *lndClientConfig `group:"Lnd" namespace:"lnd"`
	Bitcoind      *bitcoindConfig  `group:"Bitcoind" namespace:"bitcoind"`

	Prometheus *prometheusConfig `group:"Prometheus" namespace:"prometheus"`
	Hub        *hubConfig        `group:"Hub" namespace:"hub"`
	GraphQL    *graphqlConfig    `group:"GraphQL" namespace:"graphql"`

	ConfigFile string `long:"config" description:"Path to configuration file"`
	LogDir     string `long:"logdir" description:"Directory to log output."`
	DebugLevel string `long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
}

// hubConfig defines the parameters for gRPC endpoint of hub management,
// with this third-party optimisation programs could send lightning
// node channels equilibrium state.
type hubConfig struct {
	Port string `long:"port" description:"Port on which GRPC hub manager is working"`
	Host string `long:"host" description:"Host on which GRPC hub manager is working"`
}

type lndClientConfig struct {
	Network      string            `long:"network" description:"Blockchain network which should be used" choice:"simnet" choice:"testnet" choice:"mainnet"`
	DataDir      string            `long:"dbpath" description:"Path to dir where BoltDB will be stored"`
	TlsCertPath  string            `long:"tlscertpath" description:"Path to the LND certificate"`
	MacaroonPath string            `long:"macaroonpath" description:"Path to the LND macaroon"`
	GRPCHost     string            `long:"grpchost" description:"Host on which we expect to find LND gRPC endpoint"`
	GRPCPort     string            `long:"grpcport" description:"Port on which we expect to find LND gRPC endpoint"`
	NeutrinoHost string            `long:"neutrinohost" description:"Public host where neutrino Bitcoin node resides. Needed only to inform users over API"`
	NeutrinoPort string            `long:"neutrinoport" description:"Public port where neutrino Bitcoin node resides. Needed only to inform users over API"`
	PeerHost     string            `long:"peerhost" description:"Public host where LND node resides. Needed only to inform users over API"`
	PeerPort     string            `long:"peerport" description:"Public port where LND node resides. Needed only to inform users over API"`
	KnownPeers   map[string]string `long:"knownpeer" description:"A map from peer alias to its public key"`
}

type bitcoindConfig struct {
	User string `long:"user" description:"User for accessing rpc endpoint"`
	Pass string `long:"pass" description:"Password for accessing rpc endpoint"`
	Host string `long:"host" description:"Host on which we expect to bitcoind"`
	Port string `long:"port" description:"Port on which we expect to bitcoind"`
}

type prometheusConfig struct {
	ListenHost string `long:"listenhost" description:"The host of the prometheus metrics endpoint, from which metric server is trying to fetch metrics"`
	ListenPort string `long:"listenport" description:"The port of the prometheus metrics endpoint, from which metric server is trying to fetch metrics"`
}

// getDefaultConfig return default version of service config.
func getDefaultConfig() config {
	return config{
		ConfigFile:    defaultConfigFile,
		LogDir:        defaultLogDir,
		DebugLevel:    defaultLogLevel,
		UpdateLogFile: defaultUpdatesLogFileName,

		Hub: &hubConfig{
			Port: defaultHubPort,
			Host: defaultHubHost,
		},
		Prometheus: &prometheusConfig{
			ListenHost: defaultPrometheusHost,
			ListenPort: defaultPrometheusPort,
		},

		LND: &lndClientConfig{
			Network: defaultNet,
			DataDir: defaultDbPath,
		},

		GraphQL: &graphqlConfig{
			ListenHost:       defaultGraphQLHost,
			ListenPort:       defaultGraphQLPort,
			SecureListenPort: defaultGraphQLSecurePort,
		},
	}
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
func (c *config) loadConfig() error {
	// Pre-parse the command line options to pick up an alternative config
	// file.
	preCfg := c
	if _, err := flags.Parse(preCfg); err != nil {
		return err
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", version())
		os.Exit(0)
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	if err := os.MkdirAll(homeDir, 0700); err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = fmt.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// Next, load any additional configuration options from the file.
	var configFileError error
	if err := flags.IniParse(preCfg.ConfigFile, c); err != nil {
		configFileError = err
	}

	// Finally, parse the remaining command line options again to ensure
	// they take precedence.
	if _, err := flags.Parse(c); err != nil {
		return err
	}

	// Ensure that the paths are expanded and cleaned.
	c.LogDir = cleanAndExpandPath(c.LogDir)

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(c.DebugLevel); err != nil {
		err := fmt.Errorf("%s: %v", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return err
	}

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configFileError != nil {
		log.Printf("%v", configFileError)
	}

	return nil
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
// This function is taken from https://github.com/btcsuite/btcd
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(homeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// parseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly. An appropriate error is returned if anything is
// invalid.
func parseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		if !validLogLevel(debugLevel) {
			str := "The specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}

		// Change the logging level for all subsystems.
		setLogLevels(debugLevel)

		return nil
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(debugLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "The specified debug level contains an invalid " +
				"subsystem/level pair [%v]"
			return fmt.Errorf(str, logLevelPair)
		}

		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]

		// Validate subsystem.
		if _, exists := subsystemLoggers[subsysID]; !exists {
			str := "The specified subsystem [%v] is invalid -- " +
				"supported subsytems %v"
			return fmt.Errorf(str, subsysID, supportedSubsystems())
		}

		// Validate log level.
		if !validLogLevel(logLevel) {
			str := "The specified debug level [%v] is invalid"
			return fmt.Errorf(str, logLevel)
		}

		setLogLevel(subsysID, logLevel)
	}

	return nil
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	switch logLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		fallthrough
	case "warn":
		fallthrough
	case "error":
		fallthrough
	case "critical":
		return true
	}
	return false
}

// supportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func supportedSubsystems() []string {
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}

	// Sort the subsystems for stable display.
	sort.Strings(subsystems)
	return subsystems
}
