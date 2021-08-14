package internal

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	limiter "github.com/WelcomerTeam/RealRock/limiter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

const VERSION = "0.0.1"

var (
	ErrReadConfigurationFailure        = xerrors.New("Failed to read configuration")
	ErrLoadConfigurationFailure        = xerrors.New("Failed to load configuration")
	ErrConfigurationValidateIdentify   = xerrors.New("Configuration missing valid Identify URI")
	ErrConfigurationValidateRestTunnel = xerrors.New("Configuration missing valid RestTunnel URI")
	ErrConfigurationValidateGRPC       = xerrors.New("Configuration missing valid GRPC Host")
)

type Sandwich struct {
	sync.RWMutex

	Logger    zerolog.Logger
	StartTime time.Time

	ConfigurationLocation atomic.String

	ConfigurationMu sync.RWMutex
	Configuration   *SandwichConfiguration

	// RestTunnel is a third party library that handles the ratelimiting.
	// RestTunnel can accept either a direct URL or path only (when running in reverse mode)
	// https://github.com/WelcomerTeam/RestTunnel
	RestTunnelEnabled  abool.AtomicBool
	RestTunnelOnlyPath abool.AtomicBool

	ProducerClient *MQClient

	// EventPool contains the global event pool limiter defined on startup flags.
	// EventPoolWaiting stores any events that are waiting for a spot.
	EventPool        *limiter.ConcurrencyLimiter
	EventPoolWaiting atomic.Int64
	EventPoolLimit   int

	ManagersMu sync.RWMutex
	Managers   map[string]*Manager

	State *SandwichState
}

// SandwichConfiguration represents the configuration file
type SandwichConfiguration struct {
	Logging struct {
		Level              string
		FileLoggingEnabled bool

		EncodeAsJSON bool

		Directory  string
		Filename   string
		MaxSize    int
		MaxBackups int
		MaxAge     int
		Compress   bool

		MinimalWebhooks bool
	}

	State struct {
		StoreGuildMembers bool
		StoreEmojis       bool

		EnableSmaz bool
	}

	Identify struct {
		// URL allows for variables:
		// {shard}, {shard_count}, {auth}, {manager_name}, {shard_group_id}
		URL string

		Headers map[string]string
	}

	RestTunnel struct {
		Enabled bool
		URL     string
	}

	Producer struct {
		Type          string
		Configuration map[string]interface{}
	}

	GRPC struct {
		Network string
		Host    string
	}

	Webhooks []string

	Managers []*ManagerConfiguration
}

// NewSandwich creates the application state and initializes it
func NewSandwich(logger io.Writer, configurationLocation string, eventPoolLimit int) (sg *Sandwich, err error) {
	sg = &Sandwich{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		ConfigurationMu: sync.RWMutex{},
		Configuration:   &SandwichConfiguration{},

		ManagersMu: sync.RWMutex{},
		Managers:   make(map[string]*Manager),

		EventPool:        limiter.NewConcurrencyLimiter(eventPoolLimit),
		EventPoolWaiting: *atomic.NewInt64(0),
		EventPoolLimit:   eventPoolLimit,

		State: NewSandwichState(),
	}

	sg.Lock()
	defer sg.Unlock()

	configuration, err := sg.LoadConfiguration(configurationLocation)
	if err != nil {
		return nil, err
	}

	sg.ConfigurationMu.Lock()
	defer sg.ConfigurationMu.Unlock()

	sg.Configuration = configuration

	zlLevel, err := zerolog.ParseLevel(sg.Configuration.Logging.Level)
	if err != nil {
		sg.Logger.Warn().Str("level", sg.Configuration.Logging.Level).Msg("Logging level providied is not valid")
	} else {
		sg.Logger.Info().Str("level", sg.Configuration.Logging.Level).Msg("Changed logging level")
		zerolog.SetGlobalLevel(zlLevel)
	}

	// Create file and console logging

	var writers []io.Writer

	writers = append(writers, logger)

	if sg.Configuration.Logging.FileLoggingEnabled {
		if err := os.MkdirAll(sg.Configuration.Logging.Directory, 0o744); err != nil {
			log.Error().Err(err).Str("path", sg.Configuration.Logging.Directory).Msg("Unable to create log directory")
		} else {
			lumber := &lumberjack.Logger{
				Filename:   path.Join(sg.Configuration.Logging.Directory, sg.Configuration.Logging.Filename),
				MaxBackups: sg.Configuration.Logging.MaxBackups,
				MaxSize:    sg.Configuration.Logging.MaxSize,
				MaxAge:     sg.Configuration.Logging.MaxAge,
				Compress:   sg.Configuration.Logging.Compress,
			}

			if sg.Configuration.Logging.EncodeAsJSON {
				writers = append(writers, lumber)
			} else {
				writers = append(writers, zerolog.ConsoleWriter{
					Out:        lumber,
					TimeFormat: time.Stamp,
					NoColor:    true,
				})
			}
		}
	}

	mw := io.MultiWriter(writers...)
	sg.Logger = zerolog.New(mw).With().Timestamp().Logger()
	sg.Logger.Info().Msg("Logging configured")

	return sg, nil
}

// LoadConfiguration handles loading the configuration file
func (sg *Sandwich) LoadConfiguration(path string) (configuration *SandwichConfiguration, err error) {
	sg.Logger.Debug().Msg("Loading configuration")

	defer func() {
		if err == nil {
			sg.Logger.Info().Msg("Configuration loaded")
		}
	}()

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return configuration, ErrReadConfigurationFailure
	}

	configuration = &SandwichConfiguration{}

	err = yaml.Unmarshal(file, &configuration)
	if err != nil {
		return configuration, ErrLoadConfigurationFailure
	}

	err = sg.ValidateConfiguration(configuration)
	if err != nil {
		return configuration, err
	}

	return configuration, nil
}

// SaveConfiguration handles saving the configuration file
func (sg *Sandwich) SaveConfiguration(configuration *SandwichConfiguration, path string) (err error) {
	sg.Logger.Debug().Msg("Saving configuration")

	defer func() {
		if err == nil {
			sg.Logger.Info().Msg("Flushed configuration to disk")
		}
	}()

	// If a manager does not persist, we will have to update
	// the origional configuration instead of override.
	var config *SandwichConfiguration

	oldManagers := make(map[string]*ManagerConfiguration)
	storedManagers := []*ManagerConfiguration{}

	// Iterate over our builtin managers, if any do not persist,
	// we will need to load the old configuration.
	for _, manager := range configuration.Managers {
		if !manager.Persist {
			config, err = sg.LoadConfiguration(path)
			if err != nil {
				return err
			}

			// Add the previous managers to our oldManagers map
			for _, mg := range config.Managers {
				oldManagers[mg.Identifier] = mg
			}

			break
		}
	}

	// If we find a manager that does not persist, try to use the old
	// stored manager. If one does not exist then we can just continue.
	for _, manager := range configuration.Managers {
		if !manager.Persist {
			oldManager, ok := oldManagers[manager.Identifier]
			if !ok {
				continue
			}

			manager = oldManager
		}

		storedManagers = append(storedManagers, manager)
	}

	configuration.Managers = storedManagers

	data, err := yaml.Marshal(configuration)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0o600)
	if err != nil {
		return err
	}

	return nil
}

// ValidateConfiguration ensures certain values in the configuration are passed
func (sg *Sandwich) ValidateConfiguration(configuration *SandwichConfiguration) (err error) {
	if configuration.Identify.URL == "" {
		return ErrConfigurationValidateIdentify
	}

	if configuration.RestTunnel.Enabled && configuration.RestTunnel.URL == "" {
		return ErrConfigurationValidateRestTunnel
	}

	if configuration.GRPC.Host == "" {
		return ErrConfigurationValidateGRPC
	}

	return nil
}
