package main

import (
    "encoding/json"
    "net/http"
    "io"
    "io/ioutil"
    "fmt"
    // "errors"

    "github.com/gorilla/mux"
)

// these are all testing using repo.go instead of our
// real tasks objects.  Next step is to hook these
// handlers into our actual task objects.
// built this from this blog: http://thenewstack.io/make-a-restful-json-api-go/


func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "PIM Task Manager Server")
}

func TaskIndex(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    if master.HasChildren() {
        if err := json.NewEncoder(w).Encode(master.Kids()); err != nil {
            panic(err)
        }
    } else {
        myerr := pimErrors[2] // empty list encountered
        if err := json.NewEncoder(w).Encode(myerr); err != nil {
            panic(err)
        }
    }
}

func TaskShow(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    taskId := vars["taskId"]
    t := master.FindChild(taskId)
    if t != nil {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusOK)    
        if err := json.NewEncoder(w).Encode(t); err != nil {
            panic(err)
        }
    } else {
        myerr := pimErrors[1] // requested taskid not found
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(404) // file not found - is this right for not finding the id?
        if err := json.NewEncoder(w).Encode(myerr); err != nil {
            panic(err)
        }
    }
    // fmt.Fprintln(w, "Task show:", taskId)
}

func TaskCreate(w http.ResponseWriter, r *http.Request) {
    var task Task
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        panic(err)
    }
    if err := r.Body.Close(); err != nil {
        panic(err)
    }
    if err := json.Unmarshal(body, &task); err != nil {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(422) // unprocessable entity
        if err := json.NewEncoder(w).Encode(err); err != nil {
            panic(err)
        }
    }

    t := NewTask(task.GetName())
    master.AddChild(t)
    t.Save(true)

    // t := RepoCreateTodo(todo)
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(t); err != nil {
        panic(err)
    }
}