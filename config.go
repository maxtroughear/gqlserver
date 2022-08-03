package gqlserver

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/lru"
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
	IgnoreWebSocketUpgradeCheck bool

	// ComplexityLimit of 0 means no limit
	ComplexityLimit int

	// ApqCache for automatic persisted query hashes
	ApqCache graphql.Cache

	// QueryCache
	QueryCache graphql.Cache

	// Endpoint for the GraphQL handler
	GraphqlPath string

	// Endpoint for the GraphQL playground handler
	PlaygroundPath string

	// Whether or not the playground is enabled
	PlaygroundEnabled bool

	// Whether or not schema introspection is enabled
	// Introspection is enabled when Playground is enabled
	IntrospectionEnabled bool

	// Port to bind HTTP server to
	Port int

	// Logger minimum level
	LogLevel logrus.Level

	// Name of this service to report in logs
	ServiceName string

	// Environment of this service to report in logs
	Environment Environment

	// Salt used to generated ID hashes
	IDHashSalt string

	// Minimum length of ID hashes
	IDHashMinLength int

	// New Relic Configuration
	NewRelic NewRelicConfig
}

type NewRelicConfig struct {
	// Enable New Relic integration
	Enabled bool

	// License key for New Relic integration
	LicenseKey string
}

var DefaultConfig = ServerConfig{
	IgnoreWebSocketUpgradeCheck: false,
	ComplexityLimit:             300,
	ApqCache:                    lru.New(100),
	QueryCache:                  lru.New(300),
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
		Enabled: false,
	},
}
