package query

import "context"

type ctxKey string

var (
	queryCtxKey = ctxKey("query")
)

type Query struct {
	Type      string
	Key       string
	Operation string
	Value     string
}

func AddQuery(ctx context.Context, q Query) context.Context {
	currentQuery, ok := ctx.Value(queryCtxKey).([]Query)
	if !ok {
		currentQuery = make([]Query, 0)
	}
	currentQuery = append(currentQuery, q)
	return context.WithValue(ctx, queryCtxKey, currentQuery)
}

func QueryFromContext(ctx context.Context) []Query {
	currentQuery, ok := ctx.Value(queryCtxKey).([]Query)
	if !ok {
		return nil
	}
	return currentQuery
}
