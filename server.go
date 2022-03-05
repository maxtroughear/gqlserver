package gqlserver

import (
	"os"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/emvi/hide"
	"github.com/gin-gonic/gin"
	"github.com/maxtroughear/gqlserver/auth"
	"github.com/maxtroughear/logrusextension"
	"github.com/sirupsen/logrus"
)

type Server struct {
	router  *gin.Engine
	config  ServerConfig
	handler *handler.Server
	logger  *logrus.Entry
}

func NewServer(es graphql.ExecutableSchema, cfg ServerConfig) Server {
	gin.SetMode(gin.ReleaseMode)

	hide.UseHash(hide.NewHashID(cfg.IDHashSalt, cfg.IDHashMinLength))

	server := Server{
		router:  gin.Default(),
		config:  cfg,
		handler: handler.New(es),
		logger:  defaultLogger(cfg),
	}

	// add logging extension
	server.handler.Use(logrusextension.LogrusExtension{
		Logger: server.logger,
	})

	firebaseAuth := auth.NewFirebaseAuth("")

	server.RegisterMiddleware(firebaseAuth.FirebaseAuthMiddleware())

	return server
}

func (s *Server) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.router.RouterGroup.Use(middleware...)
}

func (s *Server) Run() {
	registerRoutes(s.handler, &s.router.RouterGroup, s.config)

	s.logger.Infof("Server listening on %v", s.config.Port)

	s.router.Run(":" + strconv.Itoa(s.config.Port))
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
