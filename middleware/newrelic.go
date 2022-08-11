package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewRelicMiddleware(newRelicApp *newrelic.Application) gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ctx := ginContext.Request.Context()

		tx := newRelicApp.StartTransaction(ginContext.Request.URL.Path)
		defer tx.End()

		tx.SetWebRequest(newrelic.WebRequest{
			Header:    ginContext.Request.Header,
			Host:      ginContext.Request.Host,
			Method:    ginContext.Request.Method,
			Transport: newrelic.TransportType(ginContext.Request.URL.Scheme),
			URL:       ginContext.Request.URL,
		})

		ginContext.Request = ginContext.Request.WithContext(newrelic.NewContext(ctx, tx))
		ginContext.Next()
	}
}
