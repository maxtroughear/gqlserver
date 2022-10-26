package nrextension

import (
	"context"
	"fmt"

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

	opName := buildOperationName(oc.OperationName)

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
		defer tx.StartSegment(buildResolverName(fc.Field.Name)).End()
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

func buildOperationName(graphqlOperationName string) string {
	if graphqlOperationName == "" {
		graphqlOperationName = "UNKNOWN"
	}
	return fmt.Sprintf("GraphQL/Operation/%s", graphqlOperationName)
}

func buildResolverName(graphqlFieldName string) string {
	if graphqlFieldName == "" {
		graphqlFieldName = "UNKNOWN"
	}
	return fmt.Sprintf("GraphQL/Resolver/%s", graphqlFieldName)
}
