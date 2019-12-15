package main

import (
    "encoding/json"
    "net/http"
    "io"
    "io/ioutil"
    "fmt"
    "time"
    "log"
    // "errors"

    "github.com/gorilla/mux"
)

// Task: our central type for the whole world here - will become quite large over time
type TaskJSON struct {
    Id string  `json:"id"`        // unique id of the task - TBD make this pass through to mapper!!!
    Name string `json:"name"`     // name of the task
    State TaskState `json:"state"` // state of the task
    TargetStartTime *time.Time `json:"targetStartTime,omitempty"` // targeted start time of the task
    ActualStartTime *time.Time `json:"actualStartTime,omitempty"` // actual start time of the task
    ActualCompletionTime *time.Time `json:"actualCompletionTime,omitempty"` // time task is marked done
    Estimate time.Duration `json:"estimate"` // estimated duration of the task time.Duration
    Tags []string `json:"tags"`              // tags to set - for non-updates - make task tags match
    Dirty []string `json:"dirty"`            // for updates only - which fields to update
    SetTags []string `json:"setTags"`        // for updates only - which tags to set - set "wins"
    ResetTags []string `json:"resetTags"`    // for updates only - which tags to reset
}


func (j *TaskJSON) IsDirty(field string) bool {
    for _, v := range j.Dirty {
        if v == field {
            return true
        }
    }
    return false
}

// used for Create, Update and Replace - go to an in-memory Task from JSON
func (j *TaskJSON) ToTask(t *Task, update bool) {
    // t.SetId(j.Id) we don't set the ID, that is done by the server
    if !update || j.IsDirty("name") {
        t.SetName(j.Name)        
    }
    if !update || j.IsDirty("state") {
        t.SetState(j.State)
    }
    if !update || j.IsDirty("estimate") {
        t.SetEstimate(j.Estimate * time.Minute)
    }
    if !update || j.IsDirty("targetstarttime") {
        t.SetTargetStartTime(j.TargetStartTime)
    }
    if !update || j.IsDirty("actualstarttime") {
        t.SetActualStartTime(j.ActualStartTime)
    }    
    if !update || j.IsDirty("actualcompletiontime") {
        t.SetActualCompletionTime(j.ActualCompletionTime)
    }
    if !update { // if create or replace just set tags to match
        t.ClearTags()
        for _, v := range j.Tags {
            t.SetTag(v)
        }
    } else { // if it's an update only change what's specified to change
        for _, v := range j.ResetTags {
            t.ResetTag(v)
        }
        for _, v := range j.SetTags { // this means if same tag in both then SET wins
            t.SetTag(v)
        }
    }
}

func (j *TaskJSON) FromTask(t *Task) {
    j.Id = t.GetId()
    j.Name = t.GetName()
    j.State = t.GetState()
    j.TargetStartTime = t.GetTargetStartTime()
    j.ActualStartTime = t.GetActualStartTime()
    j.ActualCompletionTime = t.GetActualCompletionTime()
    j.Estimate = t.GetEstimate()
    j.Tags = t.GetAllTags()
}

// convert a list of tasks to a list of JSON tasks
func fromTasks(ts Tasks) []TaskJSON {
    var js []TaskJSON
    var j TaskJSON
    for _, t := range ts {
        j.FromTask(t)
        js = append(js, j)
    }
    return js
}


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

        // convert kids to JSON-ready tasks
        kids := fromTasks(master.Kids())

        if err := json.NewEncoder(w).Encode(kids); err != nil {
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
    var j TaskJSON
    j.FromTask(t)
    if t != nil {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusOK)    
        if err := json.NewEncoder(w).Encode(j); err != nil {
            panic(err)
        }
    } else {
        errorResponse(w, pimErr(notFound))
    }
    // fmt.Fprintln(w, "Task show:", taskId)
}

// TBD: have this route be a find and make the parameters
// of the URL the meta-data to match on.  For now, only
// support date.
func TaskFind(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    strDate := vars["date"]
    date, _ := time.Parse("2006-01-02", strDate)
    fmt.Println("TaskFind(date=", date, ")")
    if !date.IsZero() {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusOK)
        if master.HasChildren() {
            matching := master.Kids().FindByCompletionDate(date)
            if len(matching) > 0 {
                send := fromTasks(matching)
                if err := json.NewEncoder(w).Encode(send); err != nil {
                    panic(err)
                }
            } else {
                errorResponse(w, pimErr(emptyList))
            }
        } else {
            errorResponse(w, pimErr(emptyList))
        }
    } else {
        e := pimErr(badRequest)
        e.AppendMessage(fmt.Sprintf("date '%s' provided could not be parsed.  YYYY-MM-DD format required.", strDate))
        errorResponse(w, e)        
    }
}

// TBD: combined with TaskFind
func TaskFindToday(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    if master.HasChildren() {
        matching := master.Kids().FindToday()
        fmt.Printf("TaskFindToday() num matching = %d\n", len(matching))
        if len(matching) > 0 {
            send := fromTasks(matching)
            if err := json.NewEncoder(w).Encode(send); err != nil {
                panic(err)
            }
        } else {
            errorResponse(w, pimErr(emptyList))
        }
    } else {
        errorResponse(w, pimErr(emptyList))
    }
}

// TBD: combined with TaskFind
func TaskFindThisWeek(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    if master.HasChildren() {
        matching := master.Kids().FindThisWeek()
        fmt.Printf("TaskFindThisWeek() num matching = %d\n", len(matching))
        if len(matching) > 0 {
            send := fromTasks(matching)
            if err := json.NewEncoder(w).Encode(send); err != nil {
                panic(err)
            }
        } else {
            errorResponse(w, pimErr(emptyList))
        }
    } else {
        errorResponse(w, pimErr(emptyList))
    }
}


func taskRead(w http.ResponseWriter, r *http.Request) *TaskJSON {
    var task TaskJSON
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        panic(err)
    }
    if err := r.Body.Close(); err != nil {
        panic(err)
    }
    fmt.Printf("taskRead() payload received: %s\n", body)
    if err := json.Unmarshal(body, &task); err != nil {
        errorResponse(w, pimErr(badRequest))
        fmt.Println(err)
        return nil
    }

    return &task
}

func TaskCreate(w http.ResponseWriter, r *http.Request) {
    // read the task from the request
    taskJSON := taskRead(w, r)
    if taskJSON == nil {
        return
    }

    // create a persistable task in our world with a unique id
    t := NewTask(taskJSON.Name)
    // var t Task
    taskJSON.ToTask(t, false)
    master.AddChild(t)
    err := t.Save(true)
    if (err != nil) {
        fmt.Printf("TaskCreate: save failed with errror: %s\n", err)
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusInternalServerError)
    } else {

        // set the successful response to include task
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusCreated)
        var j TaskJSON
        j.FromTask(t)
        if err := json.NewEncoder(w).Encode(j); err != nil {
            panic(err)
        }
    }
}

func TaskReplace(w http.ResponseWriter, r *http.Request) {
    log.Println("replace")

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
    taskJSON := taskRead(w, r)
    if taskJSON == nil {
        return
    }

    // make sure ids match - if not its a bad request
    if taskJSON.Id != taskId {
        errorResponse(w, pimErr(badRequest))
        return
    }

    // fmt.Println(task.GetEstimate())
    log.Println("Task as received from client...")
    log.Printf("%+v\n", taskJSON)


    // replace all fields of the current task from the request
    taskJSON.ToTask(t, false)
    err := t.Save(false)

    if (err != nil) {
        fmt.Printf("TaskReplace: save failed with errror: %s\n", err)
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusInternalServerError)
    } else {

        // set the successful response to include replaced task
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusOK)
        var j TaskJSON
        j.FromTask(t)
        if err := json.NewEncoder(w).Encode(j); err != nil {
            panic(err)
        }
    }
}



func TaskUpdate(w http.ResponseWriter, r *http.Request) {
    log.Println("update")

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
    taskJSON := taskRead(w, r)
    if taskJSON == nil {
        return
    }

    // make sure ids match - if not its a bad request
    if taskJSON.Id != taskId {
        errorResponse(w, pimErr(badRequest))
        return
    }

    // replace only the fields of the request that were marked dirty
    taskJSON.ToTask(t, true)

    // log.Printf("update: %+v\n", taskJSON)
    t.Save(false)

    // set the successful response to include replaced task
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    var j TaskJSON
    j.FromTask(t)
    if err := json.NewEncoder(w).Encode(j); err != nil {
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
    var j TaskJSON
    j.FromTask(t)    
    if err := json.NewEncoder(w).Encode(j); err != nil {
        panic(err)
    }
}