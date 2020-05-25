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

        // a little hacky - but for all GETs add a second
        // route with our supported query parameters
        // this may be further restricted to "tasks" in
        // the future.  This is because Gorilla mux does
        // seem to allow for optional query paramters.
        // NOTE: have to register with the parameters FIRST
        // otherwise it will match the one without it first.
        // NOTE: This hack may not work when we want multiple
        // query parameters - may nee to find a way to allow
        // Gorilla to accept optional query parameters OR we'll
        // have to parse the query parameters ourselves in the
        // handlers.
        if route.Method == "GET" {
            router.
                Methods(route.Method).
                Path(route.Pattern).
                Queries("tags", "{tags}").
                Name(route.Name).
                Handler(handler)
        }

        // still be sure to register without in case
        // the caller leaves of the query string
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