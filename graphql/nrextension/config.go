package nrextension

type Config struct {
	// Reports Errors in New Relic if a panic is encountered during resolver execution.
	NoticeErrorOnResolverPanic bool `env:"NEW_RELIC_NOTICE_PANIC"`

	// Reports Errors in New Relic if an error is returned by the resolver
	NoticeErrorOnGraphQLError bool `env:"NEW_RELIC_NOTICE_ERROR"`
}
