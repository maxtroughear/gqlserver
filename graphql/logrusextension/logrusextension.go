package logrusextension

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type logrusContextKey struct{}

type LogrusExtension struct {
	Logger      *logrus.Entry
	UseNewRelic bool
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
	oc := graphql.GetOperationContext(ctx)
	fc := graphql.GetFieldContext(ctx)

	// TODO: get request ID from request context (gin middleware?)
	logger := n.Logger.WithContext(ctx).
		WithFields(logrus.Fields{
			"operation": oc.OperationName,
			"field":     fc.Field.Name,
		})

	if n.UseNewRelic {
		nr := newrelic.FromContext(ctx)
		metadata := nr.GetLinkingMetadata()
		logger = logger.WithFields(logrus.Fields{
			"entity.name": metadata.EntityName,
			"entity.guid": metadata.EntityGUID,
			"entity.type": metadata.EntityType,
			"hostname":    metadata.Hostname,
			"trace.id":    metadata.TraceID,
			"span.id":     metadata.SpanID,
		})
	}

	ctx = new(ctx, logger)
	return next(ctx)
}

func new(ctx context.Context, ctxLogger *logrus.Entry) context.Context {
	return context.WithValue(ctx, logrusContextKey{}, ctxLogger)
}

func From(ctx context.Context) *logrus.Entry {
	l, _ := ctx.Value(logrusContextKey{}).(*logrus.Entry)
	return l
}
