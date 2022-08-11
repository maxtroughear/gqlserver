package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ginContextKey struct{}

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ctx := context.WithValue(ginContext.Request.Context(), ginContextKey{}, ginContext)
		ginContext.Request = ginContext.Request.WithContext(ctx)
		ginContext.Next()
	}
}

func GinContextFromContext(ctx context.Context) *gin.Context {
	ginContext, ok := ctx.Value(ginContextKey{}).(*gin.Context)
	if !ok {
		return nil
	}
	return ginContext
}
