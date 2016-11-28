package main

import (
    "net/http"
)

type Route struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

type Routes []Route

// tbd - DELETE
var routes = Routes{
    Route{
        "TaskIndex",
        "GET",
        "/tasks",
        TaskIndex,
    },
    Route{
        "TaskShow",
        "GET",
        "/tasks/{taskId}",
        TaskShow,
    },
    Route{
        "TaskCreate",
        "POST",
        "/tasks",
        TaskCreate,
    },
}