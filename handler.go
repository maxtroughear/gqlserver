package gqlserver

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func graphqlHandler(handler *handler.Server, cfg ServerConfig) gin.HandlerFunc {
	var webSocketUpgradeCheckOrigin func(r *http.Request) bool

	if cfg.IgnoreWebSocketUpgradeCheck {
		webSocketUpgradeCheckOrigin = func(r *http.Request) bool { return true }
	} else {
		webSocketUpgradeCheckOrigin = nil
	}

	handler.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: webSocketUpgradeCheckOrigin,
		},
	})
	handler.AddTransport(transport.Options{})
	handler.AddTransport(transport.GET{})
	handler.AddTransport(transport.POST{})
	handler.AddTransport(transport.MultipartForm{})

	if cfg.ApqCache != nil {
		handler.Use(extension.AutomaticPersistedQuery{
			Cache: cfg.ApqCache,
		})
	}

	if cfg.QueryCache != nil {
		handler.SetQueryCache(cfg.QueryCache)
	}

	if cfg.PlaygroundEnabled || cfg.IntrospectionEnabled {
		handler.Use(extension.Introspection{})
	}

	if cfg.ComplexityLimit > 0 {
		handler.Use(&extension.ComplexityLimit{
			Func: func(ctx context.Context, rc *graphql.OperationContext) int {
				return cfg.ComplexityLimit
			},
		})
	}

	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler(cfg ServerConfig) gin.HandlerFunc {
	playground := playground.Handler("GraphQL Playground", cfg.GraphqlPath)
	return func(c *gin.Context) {
		playground.ServeHTTP(c.Writer, c.Request)
	}
}

func healthHandler() gin.HandlerFunc {
	// TODO(Max): Implement real health checks
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

func readyHandler() gin.HandlerFunc {
	// TODO(Max): Implement real ready checks
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

func registerRoutes(handler *handler.Server, router *gin.RouterGroup, cfg ServerConfig) {
	router.GET("/health", healthHandler())
	router.GET("/ready", readyHandler())

	if cfg.ApqCache != nil {
		router.GET(cfg.GraphqlPath, graphqlHandler(handler, cfg))
	}

	router.POST(cfg.GraphqlPath, graphqlHandler(handler, cfg))
	router.OPTIONS(cfg.GraphqlPath, graphqlHandler(handler, cfg))

	if cfg.PlaygroundEnabled {
		router.GET(cfg.PlaygroundPath, playgroundHandler(cfg))
	}
}
