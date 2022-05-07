package main

import (
    "net/http"
)

// Queries format = array of strings key/value pairs where
//    key is the query parameter and
//    value is the format for it:
//      {name} matches anything until the next slash.
//      {name:pattern} matches the given regexp pattern.
//    note Gorilla mux seems to requires ALL parameters be
//    present in order to match a URL coming in

type Route struct {
    Name        string
    Method      string
    Pattern     string
    Queries     []string
    HandlerFunc http.HandlerFunc
    NoAuth      bool
}

type Routes []Route

// this is for temporary use while i develop user functionality
// to allow me to see tasks for all users at once
var queryIgnoreUsers = []string{"ignoreusers", "{ignoreusers}"}

// i reuse this alot so put it here (temporarily added ignore users for convenience while I support it)
var queryTags = []string{"tags", "{tags}", "ignoreusers", "{ignoreusers}"}

var routes = Routes{
    Route{
        Name: "Signin",
        Method: "POST",
        Pattern: "/signin",
        Queries: []string{"email", "{email}", "password", "{password}"},
        HandlerFunc: UserSignin,
        NoAuth: true,
    },
    Route{
        Name: "Signup",
        Method: "POST",
        Pattern: "/signup",
        Queries: []string{"email", "{email}", "password", "{password}"},
        HandlerFunc: UserSignup,
        NoAuth: true,
    },
    Route{
        Name: "SignReup",
        Method: "POST",
        Pattern: "/signreup",
        HandlerFunc: UserSignReup,
    },
    Route{
        Name: "TaskIndex",
        Method: "GET",
        Pattern: "/tasks",
        Queries: queryTags,
        HandlerFunc: TaskIndex,
    },
    Route {
        Name: "TaskFindToday",
        Method: "GET",
        Pattern: "/tasks/today",
        Queries: queryTags,
        HandlerFunc: TaskFindToday,
    },
    Route {
        Name: "TaskFindThisWeek",
        Method: "GET",
        Pattern: "/tasks/thisweek",
        Queries: queryTags,
        HandlerFunc: TaskFindThisWeek,
    },
    Route{
        Name: "TaskFindComplete",
        Method: "GET",
        Pattern: "/tasks/complete",
        Queries: queryTags,
        HandlerFunc: TaskFindComplete,
    },    
    Route{
        Name: "TaskGeneralFind",
        Method: "GET",
        Pattern: "/tasks/find",
        Queries: append([]string{"fromDate", "{date}", "toDate", "{date}"}, queryIgnoreUsers...),
        HandlerFunc: TaskGeneralFind,
    }, 
    Route{
        Name: "TaskShow",
        Method: "GET",
        Pattern: "/tasks/{taskId}",
        Queries: queryTags,
        HandlerFunc: TaskShow,
    },
    Route {
        Name: "TaskFindByDate",
        Method: "GET",
        Pattern: "/tasks/date/{date}",
        Queries: queryTags,
        HandlerFunc: TaskFind,
    },
    Route{
        Name: "TaskCreate",
        Method: "POST",
        Pattern: "/tasks",
        HandlerFunc: TaskCreate,
    },
    Route{
        Name: "TaskReplace",
        Method: "PUT",
        Pattern: "/tasks/{taskId}",
        Queries: queryIgnoreUsers,
        HandlerFunc: TaskReplace,
    },
    Route{
        Name: "TaskUpdate",
        Method: "PATCH",
        Pattern: "/tasks/{taskId}",
        Queries: queryIgnoreUsers,
        HandlerFunc: TaskUpdate,
    },
    Route{
        Name: "TaskDelete",
        Method: "DELETE",
        Pattern: "/tasks/{taskId}",
        Queries: queryIgnoreUsers,
        HandlerFunc: TaskDelete,
    },
    Route{
        Name: "TagIndex",
        Method: "GET",
        Pattern: "/tags",
        Queries: queryIgnoreUsers,
        HandlerFunc: TagIndex,
    },
    Route{
        Name: "TaskReorder",
        Method: "GET",
        Pattern: "/reorder/{taskId}",
        Queries: []string{"targetId", "{targetId}"},
        HandlerFunc: TaskReorder,
    }, 
    Route{
        Name: "Undo",
        Method: "GET",
        Pattern: "/undo",
        HandlerFunc: Undo,
    }, 
    Route{
        Name: "ServerStatus",
        Method: "GET",
        Pattern: "/status",
        HandlerFunc: ServerStatus,
        NoAuth: true,
    },}