package gqllogrus

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/maxtroughear/gqlserver/middleware"
	"github.com/sirupsen/logrus"
)

type logrusContextKey struct{}

type LogrusExtension struct {
	Logger *logrus.Entry
}

var _ interface {
	graphql.HandlerExtension
	graphql.FieldInterceptor
} = LogrusExtension{}

func (n LogrusExtension) ExtensionName() string {
	return "LogrusExtension"
}

func (n LogrusExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}
func (n LogrusExtension) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	logger := middleware.LogrusFromContext(ctx)
	oc := graphql.GetOperationContext(ctx)
	fc := graphql.GetFieldContext(ctx)

	fieldLogger := logger.WithFields(logrus.Fields{
		"operation": oc.OperationName,
		"resolver":  fc.Field.Name,
	})

	ctx = new(ctx, fieldLogger)
	return next(ctx)
}

func new(ctx context.Context, ctxLogger *logrus.Entry) context.Context {
	return context.WithValue(ctx, logrusContextKey{}, ctxLogger)
}

func From(ctx context.Context) *logrus.Entry {
	logger, ok := ctx.Value(logrusContextKey{}).(*logrus.Entry)
	if !ok {
		return nil
	}
	return logger
}
