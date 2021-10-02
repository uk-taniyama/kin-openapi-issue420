package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

//go:embed petstore.yaml
var petstoreYaml []byte

func main() {
	serverURL := os.Getenv("SERVER_URL")
	serverAddr := os.Getenv("SERVER_ADDR")

	doc, err := openapi3.NewLoader().LoadFromData(petstoreYaml)
	if err != nil {
		panic(err)
	}

	// replace serverUrl
	doc.Servers = openapi3.Servers{
		&openapi3.Server{
			URL: serverURL,
		},
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(serverAddr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api-docs" {
			b, _ := json.Marshal(doc)
			w.Write(b)
			return
		}
		route, pathParams, err := router.FindRoute(r)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				MultiError: true,
			},
		})
		switch err := err.(type) {
		case nil:
		case openapi3.MultiError:
			issues := convertError(err)
			names := make([]string, 0, len(issues))
			for k := range issues {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, k := range names {
				msgs := issues[k]
				fmt.Println("===== Start New Error =====")
				fmt.Println(k + ":")
				for _, msg := range msgs {
					fmt.Printf("\t%s\n", msg)
				}
			}
			w.WriteHeader(http.StatusBadRequest)
		default:
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	if err != nil {
		panic(err)
	}
}

const (
	prefixBody = "@body"
	unknown    = "@unknown"
)

func convertError(me openapi3.MultiError) map[string][]string {
	issues := make(map[string][]string)
	for _, err := range me {
		switch err := err.(type) {
		case *openapi3.SchemaError:
			// Can inspect schema validation errors here, e.g. err.Value
			field := prefixBody
			if path := err.JSONPointer(); len(path) > 0 {
				field = fmt.Sprintf("%s.%s", field, strings.Join(path, "."))
			}
			if _, ok := issues[field]; !ok {
				issues[field] = make([]string, 0, 3)
			}
			issues[field] = append(issues[field], err.Error())
		case *openapi3filter.RequestError: // possible there were multiple issues that failed validation
			if err, ok := err.Err.(openapi3.MultiError); ok {
				for k, v := range convertError(err) {
					if _, ok := issues[k]; !ok {
						issues[k] = make([]string, 0, 3)
					}
					issues[k] = append(issues[k], v...)
				}
				continue
			}

			// check if invalid HTTP parameter
			if err.Parameter != nil {
				prefix := err.Parameter.In
				name := fmt.Sprintf("%s.%s", prefix, err.Parameter.Name)
				if _, ok := issues[name]; !ok {
					issues[name] = make([]string, 0, 3)
				}
				issues[name] = append(issues[name], err.Error())
				continue
			}

			// check if requestBody
			if err.RequestBody != nil {
				if _, ok := issues[prefixBody]; !ok {
					issues[prefixBody] = make([]string, 0, 3)
				}
				issues[prefixBody] = append(issues[prefixBody], err.Error())
				continue
			}
		default:
			reasons, ok := issues[unknown]
			if !ok {
				reasons = make([]string, 0, 3)
			}
			reasons = append(reasons, err.Error())
			issues[unknown] = reasons
		}
	}
	return issues
}
