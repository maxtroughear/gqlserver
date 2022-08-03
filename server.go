package gqlserver

import (
	"os"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/emvi/hide"
	"github.com/gin-gonic/gin"
	"github.com/maxtroughear/nrextension"
	"github.com/maxtroughear/zapgqlgen"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vektah/gqlparser/v2/formatter"
	"go.uber.org/zap"
)

var (
	parsedSchema string
)

type Server struct {
	router  *gin.Engine
	config  ServerConfig
	handler *handler.Server
	logger  *zap.Logger
}

func NewServer(es graphql.ExecutableSchema, cfg ServerConfig) Server {
	gin.SetMode(gin.ReleaseMode)

	hide.UseHash(hide.NewHashID(cfg.IDHashSalt, cfg.IDHashMinLength))

	router := gin.New()
	router.Use(gin.Recovery())

	server := Server{
		router:  router,
		config:  cfg,
		handler: handler.New(es),
		logger:  defaultLogger(cfg),
	}

	s := new(strings.Builder)
	f := formatter.NewFormatter(s)
	f.FormatSchema(es.Schema())
	parsedSchema = s.String()

	// add logging extensions
	if cfg.NewRelic.Enabled {
		nrApp, err := newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.ServiceName),
			newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
			newrelic.ConfigDistributedTracerEnabled(true),
			func(cfg *newrelic.Config) {
				cfg.ErrorCollector.RecordPanics = true
			},
		)
		if err != nil {
			panic(err)
		}

		server.handler.Use(nrextension.NrExtension{
			NrApp: nrApp,
		})
	}

	server.handler.Use(zapgqlgen.ZapExtension{
		Logger:      server.logger,
		UseNewRelic: cfg.NewRelic.Enabled,
	})

	return server
}

func (s *Server) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.router.RouterGroup.Use(middleware...)
}

func (s *Server) RegisterExtension(extension graphql.HandlerExtension) {
	s.handler.Use(extension)
}

func (s *Server) Run() {
	registerRoutes(s.handler, &s.router.RouterGroup, s.config)

	s.logger.Info("Server starting", []zap.Field{
		zap.Int("port", s.config.Port),
	}...)

	s.router.Run(":" + strconv.Itoa(s.config.Port))
}

func ParsedSchema() string {
	return parsedSchema
}

func defaultLogger(cfg ServerConfig) *zap.Logger {

	logger, err := createLogger(cfg)
	if err != nil {
		panic(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("failed to retrieve hostname", zap.Error(err))
		hostname = "unknown"
	}

	defaultFields := []zap.Field{
		zap.String("service", cfg.ServiceName),
		zap.String("env", string(cfg.Environment)),
		zap.String("hostname", hostname),
	}

	logger = logger.With(defaultFields...)

	return logger
}

func createLogger(cfg ServerConfig) (*zap.Logger, error) {
	if cfg.Environment == Dev {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
