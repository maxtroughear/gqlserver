package gqlserver

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/maxtroughear/gqlserver/auth"
	"github.com/sirupsen/logrus"
)

type Environment string

const (
	Dev     Environment = "dev"
	Test    Environment = "test"
	Staging Environment = "staging"
	Prod    Environment = "prod"
)

type ServerConfig struct {
	//
	IgnoreWebSocketUpgradeCheck bool `env:"IGNORE_WEB_SOCKET_UPGRADE_CHECK"`

	// ComplexityLimit of 0 means no limit
	ComplexityLimit int `env:"COMPLEXITY_LIMIT"`

	// ApqCache for automatic persisted query hashes
	ApqCache graphql.Cache

	// QueryCache
	QueryCache graphql.Cache

	// Endpoint for the GraphQL handler
	GraphqlPath string `env:"GRAPHQL_PATH"`

	// Endpoint for the GraphQL playground handler
	PlaygroundPath string `env:"PLAYGROUND_PATH"`

	// Whether or not the playground is enabled
	PlaygroundEnabled bool `env:"PLAYGROUND_ENABLED"`

	// Whether or not schema introspection is enabled
	// Introspection is enabled when Playground is enabled
	IntrospectionEnabled bool `env:"INTROSPECTION_ENABLED"`

	// Port to bind HTTP server to
	Port int `env:"PORT"`

	// Logger minimum level
	LogLevel logrus.Level `env:"LOG_LEVEL"`

	// Name of this service to report in logs
	ServiceName string `env:"SERVICE_NAME"`

	// Environment of this service to report in logs
	Environment Environment `env:"ENVIRONMENT"`

	// Salt used to generated ID hashes
	IDHashSalt string `env:"ID_HASH_SALT"`

	// Minimum length of ID hashes
	IDHashMinLength int `env:"ID_HASH_MIN_LENGTH"`

	// New Relic Configuration
	NewRelic NewRelicConfig

	// Auth Configuration
	Auth auth.AuthConfig

	// CORS Configuration
	Cors CorsConfig
}

type NewRelicConfig struct {
	// Enable New Relic integration
	Enabled bool `env:"NEW_RELIC_ENABLED"`

	// License key for New Relic integration
	LicenseKey string `env:"NEW_RELIC_LICENSE_KEY"`

	// Whether or not the New Relic account is in the EU region
	EuRegion bool `env:"NEW_RELIC_EU_REGION"`
}

type CorsConfig struct {
	// Enable CORS
	Enabled bool `env:"CORS_ENABLED"`

	AllowOrigins []string `env:"CORS_ALLOW_ORIGINS"`
}

var DefaultConfig = ServerConfig{
	IgnoreWebSocketUpgradeCheck: false,
	ComplexityLimit:             300,
	ApqCache:                    lru.New(100),
	QueryCache:                  lru.New(1000),
	GraphqlPath:                 "/graphql",
	PlaygroundPath:              "/play",
	PlaygroundEnabled:           false,
	IntrospectionEnabled:        false,
	Port:                        3000,
	LogLevel:                    logrus.InfoLevel,
	ServiceName:                 "unnamed",
	Environment:                 Dev,
	IDHashSalt:                  "notasecret",
	IDHashMinLength:             7,
	NewRelic: NewRelicConfig{
		Enabled:  false,
		EuRegion: false,
	},
	Auth: auth.AuthConfig{
		FirebaseEnabled: false,
	},
	Cors: CorsConfig{
		Enabled:      true,
		AllowOrigins: []string{"http://localhost:3000"},
	},
}

func NewConfigFromEnvironment() ServerConfig {
	godotenv.Load()

	config := DefaultConfig

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}
