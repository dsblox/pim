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
    Route{
        "TaskReplace",
        "PUT",
        "/tasks/{taskId}",
        TaskReplace,
    },
    Route{
        "TaskUpdate",
        "PATCH",
        "/tasks/{taskId}",
        TaskReplace,
    },
    Route{
        "TaskDelete",
        "DELETE",
        "/tasks/{taskId}",
        TaskDelete,
    },
}