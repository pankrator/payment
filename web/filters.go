package web

import (
	"net/http"
	"strings"
)

type Filter interface {
	Execute(http.ResponseWriter, *http.Request, http.HandlerFunc)
	Matchers() []Endpoint
}

func MatchFilters(endpoint Endpoint, filters []Filter) []Filter {
	matchingFilters := make([]Filter, 0)
	for _, f := range filters {
		fEndpoints := f.Matchers()
		for _, fEndpoint := range fEndpoints {
			if strings.HasPrefix(endpoint.Path, fEndpoint.Path) && endpoint.Method == fEndpoint.Method {
				matchingFilters = append(matchingFilters, f)
				break
			}
		}
	}
	return matchingFilters
}

func Chain(handler http.HandlerFunc, filters []Filter) http.HandlerFunc {
	handlers := make([]http.HandlerFunc, len(filters)+1)
	handlers[len(filters)] = handler

	for i := len(filters) - 1; i >= 0; i-- {
		i := i
		handlers[i] = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			filters[i].Execute(rw, req, handlers[i+1])
		})
	}

	return handlers[0]
}
