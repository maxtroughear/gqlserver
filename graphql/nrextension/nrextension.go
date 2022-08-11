package nrextension

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type NrExtension struct {
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
} = NrExtension{}

func (n NrExtension) ExtensionName() string {
	return "NrExtension"
}

func (n NrExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (n NrExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	tx := newrelic.FromContext(ctx)
	oc := graphql.GetOperationContext(ctx)

	var opName string
	if oc.OperationName == "" {
		opName = "UNKNOWN"
	} else {
		opName = oc.OperationName
	}

	if tx != nil {
		tx.SetName(opName)
		defer tx.StartSegment(opName).End()
	}

	return next(ctx)
}

func (n NrExtension) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	tx := newrelic.FromContext(ctx)
	fc := graphql.GetFieldContext(ctx)

	if fc.IsResolver && tx != nil {
		defer tx.StartSegment(fc.Field.Name).End()
	}

	// catch any panics and send to NR
	defer func() {
		if r := recover(); r != nil {
			tx.NoticeError(r.(error))
			panic(r)
		}
	}()

	return next(ctx)
}
