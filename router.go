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