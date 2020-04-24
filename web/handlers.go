package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/pankrator/payment/model"
)

type HTTPError struct {
	StatusCode  int
	Description string
}

type Request struct {
	Request *http.Request
	Model   model.Object
}

type HandlerFunc func(rw http.ResponseWriter, req *Request)

func HandlerWrapper(handler HandlerFunc, modelBlueprint func() model.Object) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			WriteError(rw, &HTTPError{
				StatusCode:  http.StatusInternalServerError,
				Description: err.Error(),
			})
			return
		}

		model := modelBlueprint()

		contentType := req.Header.Get("Content-Type")
		switch contentType {
		case "application/xml":
			err := xml.Unmarshal(data, model)
			if err != nil {
				log.Printf("Could not parse XML: %s", err)
				WriteError(rw, &HTTPError{
					StatusCode:  http.StatusBadRequest,
					Description: "Could not parse XML",
				})
				return
			}
		default:
			err := json.Unmarshal(data, model)
			if err != nil {
				log.Printf("Could not parse JSON: %s", err)
				WriteError(rw, &HTTPError{
					StatusCode:  http.StatusBadRequest,
					Description: "Could not parse JSON: %s",
				})
				return
			}
		}

		if err := model.Validate(); err != nil {
			WriteError(rw, &HTTPError{
				StatusCode:  http.StatusBadRequest,
				Description: fmt.Sprintf("Validation of model failed: %s", err),
			})
			return
		}

		handler(rw, &Request{
			Request: req,
			Model:   model,
		})
	}
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

func WriteJSON(rw http.ResponseWriter, v interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	bytes, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		panic(marshalErr)
	}
	if _, errWrite := rw.Write(bytes); errWrite != nil {
		panic(errWrite)
	}
}

func WriteError(rw http.ResponseWriter, err *HTTPError) {
	log.Printf("error occured %s", err.Description)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(err.StatusCode)
	WriteJSON(rw, err)
}