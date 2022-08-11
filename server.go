package gqlserver

import (
	"os"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/emvi/hide"
	"github.com/gin-gonic/gin"
	"github.com/maxtroughear/gqlserver/auth"
	"github.com/maxtroughear/gqlserver/graphql/gqllogrus"
	"github.com/maxtroughear/gqlserver/graphql/nrextension"
	"github.com/maxtroughear/gqlserver/middleware"
	"github.com/maxtroughear/logrusnrhook"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/formatter"
)

var (
	parsedSchema string
)

type Server struct {
	router  *gin.Engine
	config  ServerConfig
	handler *handler.Server
	Logger  *logrus.Entry
}

func NewServer(es graphql.ExecutableSchema, cfg ServerConfig) Server {
	gin.SetMode(gin.ReleaseMode)

	hide.UseHash(hide.NewHashID(cfg.IDHashSalt, cfg.IDHashMinLength))

	logger := defaultLogger(cfg)

	router := gin.New()

	// router middleware
	router.Use(gin.Recovery())
	router.Use(middleware.GinContextToContextMiddleware())
	if cfg.NewRelic.Enabled {
		logrus.AddHook(logrusnrhook.NewNrHook(cfg.ServiceName, cfg.NewRelic.LicenseKey, cfg.NewRelic.EuRegion))

		router.Use(middleware.NewRelicMiddleware(newNrApp(cfg)))
	}
	router.Use(middleware.LogrusMiddleware(logger, cfg.NewRelic.Enabled))
	if cfg.Auth.FirebaseEnabled {
		firebaseApp := auth.NewFirebaseAuth(cfg.Auth)
		firebaseApp.FirebaseAuthMiddleware()
	}

	server := Server{
		router:  router,
		config:  cfg,
		handler: handler.New(es),
		Logger:  logger,
	}

	s := new(strings.Builder)
	f := formatter.NewFormatter(s)
	f.FormatSchema(es.Schema())
	parsedSchema = s.String()

	// add logging extensions
	if cfg.NewRelic.Enabled {
		server.RegisterExtension(nrextension.NrExtension{})
	}
	server.RegisterExtension(gqllogrus.LogrusExtension{
		Logger: server.Logger,
	})

	return server
}

func (s *Server) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.router.Use(middleware...)
}

func (s *Server) RegisterExtension(extension graphql.HandlerExtension) {
	s.handler.Use(extension)
}

func (s *Server) Run() {
	registerRoutes(s.handler, &s.router.RouterGroup, s.config)

	s.Logger.Infof("Server listening on %v", s.config.Port)

	s.router.Run(":" + strconv.Itoa(s.config.Port))
}

func ParsedSchema() string {
	return parsedSchema
}

func defaultLogger(cfg ServerConfig) *logrus.Entry {
	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(cfg.LogLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	hostname, _ := os.Hostname()

	logger := logrus.WithFields(logrus.Fields{
		"service":  cfg.ServiceName,
		"env":      cfg.Environment,
		"hostname": hostname,
	})

	return logger
}

func newNrApp(cfg ServerConfig) *newrelic.Application {
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
	return nrApp
}
