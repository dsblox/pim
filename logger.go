package main

import (
    "log"
    "net/http"
    "time"
)

/*
==========================================================
 LoggingResponseWriter
----------------------------------------------------------
 This class exists merely so we can log the http response
 code and the only way to do this is to wrap the response
 object, catch / override the WriteHeader() function and
 save the status code when it is provided.  We then
 modify our Logger() to use our wrapped version instead
 of the raw ResponseWriter.
========================================================*/
type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
    return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) StatusCode() int {
    return lrw.statusCode;
}

/*
==========================================================
 Logger
----------------------------------------------------------
 Using the logger as the handler allows us to log all
 HTTP as it moves through the system.
========================================================*/
func Logger(inner http.Handler, name string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        lrw := NewLoggingResponseWriter(w)
        inner.ServeHTTP(lrw, r)

        log.Printf(
            "%s\t%s\t%s\t%d\t%s",
            r.Method,
            r.RequestURI,
            name,
            lrw.StatusCode(),
            time.Since(start))

    })
}

func CORS(inner http.Handler, name string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        // log.Fatal(http.ListenAndServe(":5000", handlers.CORS(credentials, methods, origins)(router)))
        
        credentials := handlers.AllowCredentials()
        methods := handlers.AllowedMethods([]string{"POST"})
        ttl := handlers.MaxAge(3600)
        origins := handlers.AllowedOrigins([]string{"*"})

        inner.ServeHTTP(w, r)
    })    
}

