package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type logrusContextKey struct{}

func LogrusMiddleware(logger *logrus.Entry) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ctx := ginContext.Request.Context()
		ginLogger := logger.WithContext(ctx).
			WithFields(logrus.Fields{
				"http.host":      ginContext.Request.Host,
				"http.method":    ginContext.Request.Method,
				"http.transport": ginContext.Request.URL.Scheme,
				"http.path":      ginContext.Request.URL.Path,
			})

		tx := newrelic.FromContext(ctx)
		if tx != nil {
			metadata := tx.GetLinkingMetadata()
			ginLogger = logger.WithFields(logrus.Fields{
				"entity.name": metadata.EntityName,
				"entity.guid": metadata.EntityGUID,
				"entity.type": metadata.EntityType,
				"hostname":    metadata.Hostname,
				"trace.id":    metadata.TraceID,
				"span.id":     metadata.SpanID,
			})
		}

		newCtx := context.WithValue(ginContext.Request.Context(), logrusContextKey{}, ginLogger)
		ginContext.Request = ginContext.Request.WithContext(newCtx)
		ginContext.Next()
	}
}

func LogrusFromContext(ctx context.Context) *logrus.Entry {
	logger, ok := ctx.Value(logrusContextKey{}).(*logrus.Entry)
	if !ok {
		return nil
	}
	return logger
}
