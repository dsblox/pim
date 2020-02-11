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
    Route {
        "TaskFindToday",
        "GET",
        "/tasks/today",
        TaskFindToday,
    },
    Route {
        "TaskFindThisWeek",
        "GET",
        "/tasks/thisweek",
        TaskFindThisWeek,
    },
    Route{
        "TaskFindComplete",
        "GET",
        "/tasks/complete",
        TaskFindComplete,
    },    
    Route{
        "TaskShow",
        "GET",
        "/tasks/{taskId}",
        TaskShow,
    },
    Route {
        "TaskFindByDate",
        "GET",
        "/tasks/date/{date}",
        TaskFind,
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
        TaskUpdate,
    },
    Route{
        "TaskDelete",
        "DELETE",
        "/tasks/{taskId}",
        TaskDelete,
    },
    Route{
        "ServerStatus",
        "GET",
        "/status",
        ServerStatus,
    },}