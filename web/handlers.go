package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/pankrator/payment/model"
)

type HTTPError struct {
	StatusCode  int    `json:"status"`
	Description string `json:"description"`
}

func (he *HTTPError) Error() string {
	return he.Description
}

type Request struct {
	Request *http.Request
	Model   model.Object
}

type HandlerFunc func(rw http.ResponseWriter, req *Request)

func HandlerWrapper(handler HandlerFunc, modelBlueprint func() model.Object) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		webRequest := &Request{
			Request: req,
		}
		if req.Method == http.MethodPost {
			data, err := ReadBody(req.Body)
			if err != nil {
				WriteError(rw, err)
				return
			}

			contentType := req.Header.Get("Content-Type")
			object := modelBlueprint()
			object, err = parseModel(contentType, data, object)
			if err != nil {
				WriteError(rw, err)
				return
			}

			if err := object.Validate(); err != nil {
				WriteError(rw, &HTTPError{
					StatusCode:  http.StatusBadRequest,
					Description: fmt.Sprintf("Validation of model failed: %s", err),
				})
				return
			}
			webRequest.Model = object
		}

		handler(rw, webRequest)
	}
}

func parseModel(contentType string, data []byte, object model.Object) (model.Object, error) {
	parser, found := GetParser(contentType)
	if !found {
		return nil, &HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: fmt.Sprintf("No parser found for type %s", contentType),
		}
	}
	err := parser.Unmarshal(data, object)
	if err != nil {
		return nil, &HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: fmt.Sprintf("Could not parse type %s: %s", contentType, err),
		}
	}

	return object, nil
}

func ReadBody(body io.ReadCloser) ([]byte, error) {
	data, err := ioutil.ReadAll(body)
	defer func() {
		if err := body.Close(); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		return nil, &HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: err.Error(),
		}
	}
	return data, nil
}

func BodyToObject(body io.ReadCloser, value interface{}) error {
	bytes, err := ReadBody(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, value)
}

func recoveryMiddleware() mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					WriteError(rw, &HTTPError{
						StatusCode:  http.StatusInternalServerError,
						Description: "Unexpected error occured",
					})
					log.Printf("Unexpected error occured %s", err)
					debug.PrintStack()
				}
			}()
			handler.ServeHTTP(rw, req)
		})
	}
}

func WriteJSON(rw http.ResponseWriter, status int, v interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	bytes, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		panic(marshalErr)
	}
	if _, errWrite := rw.Write(bytes); errWrite != nil {
		panic(errWrite)
	}
}

func WriteError(rw http.ResponseWriter, err error) {
	log.Printf("error occured %s", err)
	rw.Header().Set("Content-Type", "application/json")
	var httpError *HTTPError
	switch v := err.(type) {
	case *HTTPError:
		httpError = v
	default:
		httpError = &HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "Internal Server Error",
		}
	}

	WriteJSON(rw, httpError.StatusCode, httpError)
}
