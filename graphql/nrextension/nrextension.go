package nrextension

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/maxtroughear/gqlserver/extension"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type NrExtension struct {
	NrApp *newrelic.Application
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
	graphql.ResponseInterceptor
} = NrExtension{}

func (n NrExtension) ExtensionName() string {
	return "NrExtension"
}

func (n NrExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (n NrExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)

	var opName string
	if oc.OperationName == "" {
		opName = "UNKNOWN"
	} else {
		opName = oc.OperationName
	}
	tx := n.NrApp.StartTransaction(opName)

	ginContext := extension.GinContextFromContext(ctx)

	tx.SetWebRequest(newrelic.WebRequest{
		Header:    ginContext.Request.Header,
		Host:      ginContext.Request.Host,
		Method:    ginContext.Request.Method,
		Transport: newrelic.TransportType(ginContext.Request.URL.Scheme),
		URL:       ginContext.Request.URL,
	})

	ctx = newrelic.NewContext(ctx, tx)
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

func (n NrExtension) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	tx := newrelic.FromContext(ctx)

	defer func() {
		if tx != nil {
			go tx.End()
		}

	}()
	return next(ctx)
}
