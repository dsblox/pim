package main

import (
    "encoding/json"
    "net/http"
    "io"
    "io/ioutil"
    "fmt"
    "time"
    // "errors"

    "github.com/gorilla/mux"
)

func errorResponse(w http.ResponseWriter, e PimError) {
  w.Header().Set("Content-Type", "application/json; charset=UTF-8")
  w.WriteHeader(e.Response) // file not found - is this right for not finding the id?
  if err := json.NewEncoder(w).Encode(e); err != nil {
    panic(err)
  }
}


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
        errorResponse(w, pimErr(emptyList))
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
        errorResponse(w, pimErr(notFound))
    }
    // fmt.Fprintln(w, "Task show:", taskId)
}

func taskRead(w http.ResponseWriter, r *http.Request) *Task {
    var task Task
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        panic(err)
    }
    if err := r.Body.Close(); err != nil {
        panic(err)
    }
    if err := json.Unmarshal(body, &task); err != nil {
        errorResponse(w, pimErr(badRequest))
        fmt.Println(err)
        return nil
    }

    return &task
}

func TaskCreate(w http.ResponseWriter, r *http.Request) {
    // read the task from the request
    task := taskRead(w, r)
    if task == nil {
        return
    }

    // create a persistable task in our world
    t := NewTask(task.GetName())
    t.SetState(task.GetState())
    t.SetTargetStartTime(task.GetTargetStartTime())
    t.SetEstimate(task.GetEstimate() * time.Minute)
    master.AddChild(t)
    t.Save(true)

    // set the successful response to include task
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(t); err != nil {
        panic(err)
    }
}

func TaskReplace(w http.ResponseWriter, r *http.Request) {
    fmt.Println("replace")

    // extract the task id from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]

    // make sure the task we wish to replace exists
    // note that we do not allow clients to specify the
    // id of a new task - POST is always used to create tasks
    t := master.FindChild(taskId)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    // read the task from the request
    task := taskRead(w, r)
    if task == nil {
        return
    }

    // make sure ids match - if not its a bad request
    if task.GetId() != taskId {
        errorResponse(w, pimErr(badRequest))
        return
    }

    // fmt.Println(task.GetEstimate())
    fmt.Println("Task as received from client...")
    fmt.Println(task)


    // replace all fields of the current task from the request
    t.SetName(task.GetName())
    t.SetState(task.GetState())
    t.SetTargetStartTime(task.GetTargetStartTime())
    t.SetEstimate(task.GetEstimate() * time.Minute)
    t.Save(false)

    // set the successful response to include replaced task
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(t); err != nil {
        panic(err)
    }
}

func TaskUpdate(w http.ResponseWriter, r *http.Request) {
    fmt.Println("update")

    // extract the task id from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]

    // make sure the task we wish to replace exists
    // note that we do not allow clients to specify the
    // id of a new task - POST is always used to create tasks
    t := master.FindChild(taskId)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    // read the task from the request
    task := taskRead(w, r)
    if task == nil {
        return
    }

    // make sure ids match - if not its a bad request
    if task.GetId() != taskId {
        errorResponse(w, pimErr(badRequest))
        return
    }

    // replace only the fields of the request that have a value
    // in the request - THIS HAS NOT BEEN TESTED
    // TBD: more fields - for now just name
    // if (t.GetName() != nil) {
    //     t.SetName(task.GetName())
    // }
    t.Save(false)

    // set the successful response to include replaced task
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(t); err != nil {
        panic(err)
    }
}

func TaskDelete(w http.ResponseWriter, r *http.Request) {

    // extract the task id from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]

    // make sure the task we wish to replace exists
    t := master.FindChild(taskId)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    // delete the task which will immediately delete in storage
    // the parameter is for a new parent of any of this tasks
    // children - which we may wish to support later
    t.Remove(nil)

    // set the successful response to indicate deletion
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(t); err != nil {
        panic(err)
    }
}