package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func NewRouter(files string) *mux.Router {
    router := mux.NewRouter().StrictSlash(true)
    for _, route := range routes {
        var handler http.Handler
        handler = route.HandlerFunc
        handler = Logger(handler, route.Name)

        // we make query parameters optional by registering the route 2x
        // once with the query parameters and once without, and the one
        // with the parameters has to come first. (only for GETs for now?)
        // Gorilla mux doesn't seem to support "optional" query params so
        // we must register 2x, and ALL params specified MUST be on the URL
        // or order to match the route
        if route.Method == "GET" && route.Queries != nil {
            router.
                Methods(route.Method).
                Path(route.Pattern).
                Queries(route.Queries...).
                Name(route.Name).
                Handler(handler)
        }

        // still be sure to register without queiries in case
        // the caller leaves off the query string
        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handler)


    }

    // put this last so routes in routes.go will match first
    router.PathPrefix("/").Handler(http.FileServer(http.Dir(files)))
    
    return router
}