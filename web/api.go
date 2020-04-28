package web

import "github.com/pankrator/payment/model"

// Controller is the interface that should be implemented by structures to be controllers
type Controller interface {
	Routes() []Route
}

// Route is a mapping between an Endpoint and a REST API Handler
type Route struct {
	// Endpoint is the combination of Path and HTTP Method for the specified route
	Endpoint Endpoint

	// Handler is the function that should handle incoming requests for this endpoint
	Handler HandlerFunc

	ModelBlueprint func() model.Object
}

// Endpoint is the pair of the http method and path
type Endpoint struct {
	Method string
	Path   string
}

// Api contains all the controllers for the instantiated Api and can be attached to the http server
type Api struct {
	Filters     []Filter
	Controllers []Controller
}
