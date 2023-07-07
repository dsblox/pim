package main

import (
    "encoding/json"
    "net/http"
    "io"
    "io/ioutil"
    "fmt"
    "time"
    "strings"
    "github.com/gorilla/mux"
)

func enableCors(w *http.ResponseWriter) {
    fmt.Printf("enabling CORS *\n")
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// TEMPORARY way to allow all tasks to be owned by one user while developing user behavior
func UserIfOn(w http.ResponseWriter, r *http.Request) *User {
    // TEMPORARY while we develop to easily turn multiuser functionality on / off

    // pull url parameter to ignore current user
    vars := mux.Vars(r)
    _, ignoreUsers := vars["ignoreusers"]

    // if ignoring users don't get the user
    if ignoreUsers {
        return nil
    }

    // the usual case is this just wraps UserFromRequest
    return UserFromRequest(w, r)
}

// Task: our central type for the whole world here - will become quite large over time
type TaskJSON struct {
    Id string  `json:"id"`        // unique id of the task - TBD make this pass through to mapper!!!
    Name string `json:"name"`     // name of the task
    State TaskState `json:"state"` // state of the task
    TargetStartTime *time.Time `json:"targetStartTime,omitempty"` // targeted start time of the task
    ActualStartTime *time.Time `json:"actualStartTime,omitempty"` // actual start time of the task
    ActualCompletionTime *time.Time `json:"actualCompletionTime,omitempty"` // time task is marked done
    // Estimate time.Duration `json:"estimate"` // estimated duration of the task time.Duration
    Estimate int `json:"estimate"`
    Tags []string `json:"tags"`              // tags to set - for non-updates - make task tags match
    Links []string `json:"links"`             // links to set - for non-updates - make links match
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
        t.SetEstimate(time.Duration(j.Estimate) * time.Minute)
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
    if !update || j.IsDirty("links") { // for now we always take the link(s) we're given
        t.ClearLinks()
        for _, v := range j.Links {
            err := t.AddLink(v, 0, 0)
            if err != nil {
                // TBD: find a way to report an error back to the client of the API
                //      for now we are failing almost silently with just the console error
                fmt.Printf("ToTask(): failed trying to add invalid URL <%v>\n", v) 
            }
        }
    } // in the future we'll support set/reset links like we do for tags perhaps

}

func (j *TaskJSON) FromTask(t *Task) {
    if t != nil {
        j.Id = t.GetId()
        j.Name = t.GetName()
        j.State = t.GetState()
        j.TargetStartTime = t.GetTargetStartTime()
        j.ActualStartTime = t.GetActualStartTime()
        j.ActualCompletionTime = t.GetActualCompletionTime()
        j.Estimate = int(t.GetEstimate())
        j.Tags = t.GetAllTags()
        j.Links = t.GetLinks()
    }
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

// Cmd: this is the outer envelope of all our responses
// NOT YET USED EXCEPT FOR UNDO- but can later be used 
// to standardize all our response envelopes.  For now, 
// each API callreturns a custom JSON blob for its own use.
type CmdJSON struct {
    Command    string   `json:"cmd"`
    TargetName string   `json:"target"`
    Status     int      `json:"status"`
    Error      PimError `json:"error"`
    Task       TaskJSON `json:"task"`
    TaskIds  []TaskJSON `json:"tasks"`
}


func errorResponse(w http.ResponseWriter, e PimError) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(e.Response)
    if err := json.NewEncoder(w).Encode(e); err != nil {
        panic(err)
    }
}

func successResponse(w http.ResponseWriter) {
    payload := pimSuccess()
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(payload.Response)
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        panic(err)
    }    
}

// these are all testing using repo.go instead of our
// real tasks objects.  Next step is to hook these
// handlers into our actual task objects.
// built this from this blog: http://thenewstack.io/make-a-restful-json-api-go/

func ServerStatus(w http.ResponseWriter, r *http.Request) {
    err := master.MapperError()
    if err == nil {
        fmt.Fprintln(w, "OK")
    } else {
        fmt.Fprintln(w, err)
    }
}

func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "PIM Task Manager Server")
}

/*
==============================================================================
 TagIndex()
------------------------------------------------------------------------------
 This gets a list of all the tags that have been set on any task across the
 instance, presumably to allow some user interface to select from all tags
 so they can pick one or more for filtering.

 We've implemented fully in memory, which will work with any of our data
 stores, but would be horribly inefficient for the database mapping, but
 is the required approach for the YAML mapping.  Someday, we should figure
 a way for the DataMapper to provide a method that could take advantage
 of that. 

 Result is a map with tags as the key and the numnber of instances of that
 tag as the value.  This is to let the caller know how "popular" each tag is.
============================================================================*/
func TagIndex(w http.ResponseWriter, r *http.Request) {

    // find my user so I can get tags just for this user
    user := UserIfOn(w, r)

    // fmt.Printf("TagIndex(): entry\n")
    if master.HasChildren() {
        tags := master.Kids(user).GetChildTags()
        if tags != nil {
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusOK)
            if err := json.NewEncoder(w).Encode(tags); err != nil {
                panic(err)
            }
        } else {
            errorResponse(w, pimErr(emptyList))    
        }
    } else {
        errorResponse(w, pimErr(emptyList))         
    }
}

func TaskIndex(w http.ResponseWriter, r *http.Request) {

    // find my user so I can get tasks just for this user
    user := UserIfOn(w, r)

    // pull tag filter
    vars := mux.Vars(r)
    strtags := vars["tags"]
    var tags []string
    if len(strtags) > 0 {
        tags = strings.Split(strtags, ",")
    } else {
        tags = nil
    }
    var kids []TaskJSON = nil
    if master.HasChildren() {

        // if tags filter is here then apply it
        if tags != nil && len(tags) > 0 {
            // the second parm says to automatch today and this week
            // based on dates as well as explicit tag matches
            matching := master.Kids(user).FindTagMatches(tags, true)
            if len(matching) > 0 {
                kids = fromTasks(matching)
            }
        } else {
            // convert kids to JSON-ready tasks            
            kids = fromTasks(master.Kids(user))
        }

        if kids != nil {
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusOK)
            if err := json.NewEncoder(w).Encode(kids); err != nil {
                panic(err)
            }
        } else {
            errorResponse(w, pimErr(emptyList))    
        }
    } else {
        errorResponse(w, pimErr(emptyList))
    }
}

func TaskShow(w http.ResponseWriter, r *http.Request) {

    // find my user so I only return this task id if it is mine
    user := UserIfOn(w, r)

    vars := mux.Vars(r)
    taskId := vars["taskId"]
    t := master.FindChild(taskId, user)
    if t != nil {
        var j TaskJSON
        j.FromTask(t)
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

    // find my user so I can get tasks just for this user
    user := UserIfOn(w, r)

    vars := mux.Vars(r)
    strDate := vars["date"]
    fmt.Printf("strDate=<%v>\n",strDate)
    date, _ := time.Parse("2006-01-02", strDate)
    if !date.IsZero() {
        if master.HasChildren() {
            matching := master.Kids(user).FindByCompletionDate(date)
            if len(matching) > 0 {
                send := fromTasks(matching)
                w.Header().Set("Content-Type", "application/json; charset=UTF-8")
                w.WriteHeader(http.StatusOK)
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

// in development: general find function that can take lots of
// parameters - was going to start with a date range for completion
// dates to fix the problem with timezones only from UTC (change
// the client to ask for local time range and call this).
func TaskGeneralFind(w http.ResponseWriter, r *http.Request) {

    // find my user so I can get tasks just for this user
    // user := UserIfOn(w, r) - don't bother until we actually implement this

    // tbd
    // FindBetweenCompletionDate()
    // extract the search criteria from the request
    vars := mux.Vars(r)
    taskId := vars["fromDate"]
    targetId := vars["toDate"]
    fmt.Printf("TaskGeneralFind() fromDate: %s toDate: %s vars: %v\n", taskId, targetId, vars)
    errorResponse(w, pimErr(emptyList)) // for now until we have a response
}

// TBD: combined with TaskFind
func TaskFindToday(w http.ResponseWriter, r *http.Request) {

    // find my user so I can get tasks just for this user
    user := UserIfOn(w, r)

    if master.HasChildren() {
        matching := master.Kids(user).FindToday()
        if len(matching) > 0 {
            send := fromTasks(matching)
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusOK)
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

    // find my user so I can get tasks just for this user
    user := UserIfOn(w, r)

    if master.HasChildren() {
        matching := master.Kids(user).FindThisWeek()
        if len(matching) > 0 {
            send := fromTasks(matching)
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusOK)
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
    // fmt.Printf("taskRead() payload received: %s\n", body)
    if err := json.Unmarshal(body, &task); err != nil {
        errorResponse(w, pimErr(badRequest))
        fmt.Print("taskRead(): ")
        fmt.Println(err)
        return nil
    }

    return &task
}

func TaskCreate(w http.ResponseWriter, r *http.Request) {

    // find the user creating the task who will own it
    user := UserFromRequest(w, r)
    if user == nil { return }
    // we don't use UserIfOn() - can create new tasks with users even temporarily

    // read the task from the request
    taskJSON := taskRead(w, r)
    if taskJSON == nil {
        return
    }

    // create a persistable task in our world with a unique id
    t := NewTask(taskJSON.Name)
    t.AddUser(user)
    taskJSON.ToTask(t, false)
    master.AddChild(t)
    // err := t.Save(true)
    err := CommandCreateTask(t)
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

    // find my user so I only change the task if it is mine
    user := UserIfOn(w, r)

    // extract the task id from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]

    // make sure the task we wish to replace exists
    // note that we do not allow clients to specify the
    // id of a new task - POST is always used to create tasks
    t := master.FindChild(taskId, user)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    // record the task as it appears before modification
    // to support undo
    cmd := CommandModifyTaskBegin(t)

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
    // log.Println("Task as received from client...")
    // log.Printf("%+v\n", taskJSON)


    // replace all fields of the current task from the request
    taskJSON.ToTask(t, false)

    // run the command for the undo stack
    err := CommandModifyTaskEnd(cmd, t)
    // err := t.Save(false)

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

    // find my user so I only change the task if it is mine
    user := UserIfOn(w, r)

    // extract the task id from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]

    // make sure the task we wish to replace exists
    // note that we do not allow clients to specify the
    // id of a new task - POST is always used to create tasks
    t := master.FindChild(taskId, user)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    cmd := CommandModifyTaskBegin(t)

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
    // t.Save(false)
    CommandModifyTaskEnd(cmd, t)    

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

    // find my user so I only delete the task if it is mine
    user := UserIfOn(w, r)

    // extract the task id from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]

    // make sure the task we wish to replace exists
    t := master.FindChild(taskId, user)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    // call the command system to perform the delete
    err := CommandDeleteTask(t, nil)
    if err != nil {
      errorResponse(w, pimErr(deleteFailed))
      return
    }

    // set the successful response to indicate deletion
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    var j TaskJSON
    j.FromTask(t)    
    if err := json.NewEncoder(w).Encode(j); err != nil {
        panic(err)
    }
}

func TaskFindComplete(w http.ResponseWriter, r *http.Request) {

    // find my user so I only return tasks that are mine
    user := UserIfOn(w, r)

    if master.HasChildren() {
        matching := master.Kids(user).FindCompleted()
        if len(matching) > 0 {
            send := fromTasks(matching)
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusOK)
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

// consider: should this be an PUT or POST on the task itself
// with a new "field" of "relativeImportance"?
func TaskReorder(w http.ResponseWriter, r *http.Request) {

    // find my user so I only change the task if it is mine
    user := UserIfOn(w, r)    

    // extract the task ids from the request
    vars := mux.Vars(r)
    taskId := vars["taskId"]
    targetId := vars["targetId"]
    fmt.Printf("TaskReorder() taskId: %s targetId: %s vars: %v\n", taskId, targetId, vars)

    // make sure the task we wish to move exists
    t := master.FindChild(taskId, user)
    if t == nil {
      errorResponse(w, pimErr(notFound))
      return
    }

    // make sure the task we wish to move before exists
    // but if none specified we're moving to the end
    target := master.FindChild(targetId, user)

    // make the change - assumes flat list for now
    // tbd: a better error return
    err := t.MoveBefore(master.Kids(user), target)
    if (err != nil) {
        errorResponse(w, pimErr(notFound))
        return
    }

    // save the tasks impacted - ordering may have to
    // resave the entire order each time???
    t.Save(false)
    // target.Save(false)

    // set the successful response to indicate it worked
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    var j TaskJSON
    j.FromTask(t)    
    if err := json.NewEncoder(w).Encode(j); err != nil {
        panic(err)
    }
}


/*
==============================================================================
 Undo()
------------------------------------------------------------------------------
 TBD - somehow make this per-user!!!  Undo stack needed per user!
============================================================================*/
func Undo(w http.ResponseWriter, r *http.Request) {

    // fmt.Printf("Undo(): entry\n")
    err := CommandUndo()
    if (err != nil) {
        errorResponse(w, pimErr(undoEmpty))
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)

    var undoResponse CmdJSON
    undoResponse.Command = "UNDO"
    undoResponse.TargetName = "unknown task"
    undoResponse.Status = 0 // ok

    if err := json.NewEncoder(w).Encode(undoResponse); err != nil {
        panic(err)
    }
}


/*
==============================================================================
 UserSignup()
------------------------------------------------------------------------------
 Create a new user and automatically sign them in by returning an
 authentication token.
============================================================================*/
func UserSignup(w http.ResponseWriter, r *http.Request) {

    var creds UserCredentials
    // Get the JSON body and decode into credentials
    err := json.NewDecoder(r.Body).Decode(&creds)
    if err != nil {
        // the structure of the body is wrong
        errorResponse(w, pimErr(badRequest))
        return
    }

    // make sure the email address doesn't already exist
    prev := users.FindByEmail(creds.Email)
    if prev != nil {
        errorResponse(w, pimErr(authTaken))
        return
    }

    // create the new user - may not create if email is invalid
    // or password is not up to standard
    // note that every user should get its own copy of the data mapper
    noob, errCreate := NewUser("", "unspecified", creds.Email, creds.Password, storage.CopyDataMapper())
    if errCreate != success {
        errorResponse(w, pimErr(errCreate))
        return        
    }

    // save the new user
    err = noob.Save()
    if err != nil {
        errorResponse(w, pimErr(errCreate))
        return
    }

    // add the new user to the global list (move inside the object?)
    users = append(users, noob)

    // now we have a valid user - return credentials to the client
    UserSetAuthToken(w, creds.Email)
    successResponse(w)
}

/*
==============================================================================
 UserSignin()
------------------------------------------------------------------------------
 Sign the user into the system by validating credentials that are expected
 in the request and - if the credentials are valid - putting an authentication
 token into the response that the client can reuse on future requests.
============================================================================*/
func UserSignin(w http.ResponseWriter, r *http.Request) {
    fmt.Println("UserSignin() - entry")
    enableCors(&w) 

    // get the JSON body and decode into credentials
    var creds UserCredentials
    err := json.NewDecoder(r.Body).Decode(&creds)
    if err != nil {
        // the structure of the body is wrong
        errorResponse(w, pimErr(badRequest))
        return
    }

    // we do the success case in one place here, when the
    // username and password are good, set the auth token
    // into the response
    user := users.FindByEmail(creds.Email)
    if user != nil {
        if user.CheckPassword(creds.Password) {
            UserSetAuthToken(w, creds.Email)
            successResponse(w)        
            return
        }
    }

    // note that we do not want to leak any username info so it is important
    // to return the identical error when the username doesn't exist and when
    // the password is bad, so we make sure to return the error in one place.
    // if we arrived here then one of those auth errors occurred
    errorResponse(w, pimErr(authFail))
    return
}

/*
==============================================================================
 UserSignReup()
------------------------------------------------------------------------------
 Given an already authenticated user, issue a new set of authentication
 credentials with a new expiration time.  Client is expected to call this
 occassionaly to keep their user from timing out.
============================================================================*/
func UserSignReup(w http.ResponseWriter, r *http.Request) {

    // find my user so I can get creds for next token
    user := UserFromRequest(w, r)
    if user == nil { return }

    // create a new token
    UserSetAuthToken(w, user.GetEmail())
    successResponse(w)
    return
}
