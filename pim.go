package main

import (
  "fmt"
  "bufio"
  "os"
  "strings"
  "log"
  "flag"
  "net/http"
  "time"
  "errors"
)

import (
     // "reflect"
//     _ "github.com/lib/pq"
 )

const (
    DB_NAME = "pim"
)

// interface to represent a linkage to a storage mechanism
type PimPersist interface {
  NewDataMapper(storageName string) PimPersist // return an empty mapper of your implementation type
  CopyDataMapper() PimPersist  // create a new mapper from an existing one
  Error() error // returns nil if the mapper is in a non-error state, or an error if in an error state
}

// this will replace initMasterTask() but it's return can then be used to initialize
// a task data mapper and a user data mapper.  THIS FUNCTION IS NOT YET WRITTEN FOR YAML.
func initPIMPersistence(dbName string) (*PimPersist, error) {
    if dbName == "YAML" || dbName == "yaml" {
      /*
      tdmyaml := NewTaskDataMapperYAML("yaml/tasks.yaml") // TBD: separate DB initialization for object mapper
      if tdmyaml == nil {
      log.Printf("PIM was unable to use create the YAML Data Mapper.  Exiting...\n")
      return nil, errors.New("Unable to create the data mapper")
      }
      masterTask.SetDataMapper(tdmyaml)
      */
      return nil, nil

    } else { // postgres is the default
      return nil, nil
      // return NewPimPersistPostgreSQL(dbName)
    }
}


type PimCmd int
const (
  help = iota 
  quit 
  printList
  addTask
  deleteTask
  upCurrent
  downCurrent
  completeTask
  startTask
  resetTask
  holdTask
  renameTask
  setStartTime
  setEstimate
  debugTask
)
var cmdChars = []rune{ '?', 'q', 'p', 'a', 'x', 'u', 'd', 'c', 's', 'r', 'h', 'n', 't', 'e', '~'}

// findCommand: given an entered rune will look up the command to execute
func findCommand(entered rune) int {
  for i, c := range cmdChars {
    if entered == c {
      return i
    }
  }
  return help
}

// printHelp: outputs the help on commands to the console
func printHelp() {
  fmt.Println("PIM Console Help")
  fmt.Println("  p = print task list")
  fmt.Println("  a = add task as child of current task")
  fmt.Println("  x = delete current task")
  fmt.Println("  r = rename current task")
  fmt.Println("  u = move current task up")
  fmt.Println("  d = move current task down")
  fmt.Println("  c = complete current task")
  fmt.Println("  s = start current task")
  fmt.Println("  r = reset current task to not started")
  fmt.Println("  h = put current task on hold")
  fmt.Println("  t = set start time for current task")
  fmt.Println("  e = set estimated duration of current task")
  fmt.Println("  ~ = debug by dumping all info on current task")
  fmt.Println("  ? = help")
  fmt.Println("  q = quit")
}

// moveUp: move the current task up one
func moveUp(oldCurrentTask *Task) *Task {
  var newCurrentTask = oldCurrentTask.PrevSibling(nil)
  // TBD - go to the deepest children of my previous sibling
  if newCurrentTask == nil {
    newCurrentTask = oldCurrentTask.FirstParent()
  }
  if newCurrentTask != nil {
    oldCurrentTask.current = false
    newCurrentTask.current = true
  } else {
    newCurrentTask = oldCurrentTask
  }
  return newCurrentTask // may be unchanged if at top of list
}

// moveDown: move the current task down one - note that we maintain
// the "current" flag on the entire path down the hierarhcy so we
// can know which parent we care to traverse
func moveDown(oldCurrentTask *Task) *Task {

  // first try to get the first child of the current task
  var newCurrentTask = oldCurrentTask.FirstChild()

  // if current task doesn't have any kids
  if (newCurrentTask == nil) {

    // then I need the parent of the current task to find
    // my next sibling - nill means use "current" parent
    newCurrentTask = oldCurrentTask.NextSibling(nil)

    // if still nil we need to run the (current) parent chain
    // and get the next subling of the first parent we find
    if newCurrentTask == nil && oldCurrentTask.HasParents() {
      for p := oldCurrentTask.Parent(); p != nil && newCurrentTask == nil; p = p.Parent() {
        newCurrentTask = p.NextSibling(nil)
      }
    }
  }

  // if we are changing the current task lets reset current flags
  if newCurrentTask != nil {
    oldCurrentTask.current = false
    newCurrentTask.current = true
  } else {
    newCurrentTask = oldCurrentTask
  }
  return newCurrentTask
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func initStorage(dbName string) (TaskDataMapper, error) {
  var tdm TaskDataMapper
    if dbName == "YAML" || dbName == "yaml" {

      tdmyaml := NewTaskDataMapperYAML("yaml/tasks.yaml")
      if tdmyaml == nil {
      log.Printf("PIM was unable to create the YAML Data Mapper.  Exiting...\n")
      return nil, errors.New("PIM was unable to create the YAML Data Mapper")
      }
      tdm = tdmyaml
    } else {

      // initialize the persistence layer - use PostgreSQL
      // and assign to masterTask - creating the first data
      // mapper initializes the DB - I chose to test a data-mapper
      // pattern to fully abstract the persistence layer from the
      // task functionality.  This should be the only place the
      // Tasks know how they are stored.
      tdmpg := NewTaskDataMapperPostgreSQL(false, dbName) // this is a pointer to the concrete object
      if tdmpg == nil {
        fmt.Print("PIM requires a local PostgreSQL database to running.  Exiting...\n")
        return nil, errors.New("PIM requires a local PostgreSQL database to running")
      }
      tdm = tdmpg
  }
  return tdm, nil
}

func initMasterTask(tdm TaskDataMapper) (*Task, error) {

  // dummy master task as parent to all other tasks
  var masterTask *Task = NewTaskMemoryOnly("Your Task List")

  // set the storage provider on the top-level task
  // which will automatically be inherited by all child tasks
  masterTask.SetDataMapper(tdm)

  // now load the task list recursively
  err := masterTask.Load(true)
  if err != nil {
    fmt.Printf("Error loading master task: %s\n", err)
    return nil, err
  }

  return masterTask, nil
}

func runConsoleApp(dbName string) {
  fmt.Printf("*** Welcome to PIM - The Perfect Task Manager for Your Life ***\n")

  // make sure we have somewhere to load / save things
  tdm, err := initStorage(dbName)
  if err != nil {
    log.Fatal(err)
  }     

  // now load up the master task
  var masterTask *Task
  masterTask, err = initMasterTask(tdm)
  if err != nil {
    log.Fatal(err)
  }     

  // keep track of a "cursor" task on which to work
  var currentTask *Task = masterTask
  currentTask.current = true
  printHelp()
  fmt.Println(masterTask)

  // start the "event loop" based on user input
  reader := bufio.NewReader(os.Stdin)
  var printListAfterCmd bool = true
  for cmd := printList; cmd != quit;  {

    printListAfterCmd = true
    fmt.Print("Command: ")
    text, _ := reader.ReadString('\n')
    cmd = findCommand(([]rune(text))[0])
    switch cmd {
      case help: 
        printHelp()
        printListAfterCmd = false

      case printList: 
        printListAfterCmd = true // happens anyway

      case quit: 
        fmt.Println("Goodbye!")
        printListAfterCmd = false

      case upCurrent: 
        currentTask = moveUp(currentTask)

      case downCurrent: 
        currentTask = moveDown(currentTask)

      case addTask: 
        fmt.Print("Enter name of new task: ")
        taskName, _ := reader.ReadString('\n')
        t := NewTask(strings.TrimSpace(taskName))
        currentTask.AddChild(t)

      case renameTask: 
        fmt.Print("Enter new name of current task: ")
        taskName, _ := reader.ReadString('\n')
        currentTask.SetName(strings.TrimSpace(taskName))

      case deleteTask: 
        taskToKill := currentTask
        currentTask = moveUp(currentTask)
        if (!taskToKill.HasParents()) {
          fmt.Println("Can't delete top level task list")
        } else {
          taskToKill.Remove(masterTask)
        }

      case completeTask:
        currentTask.SetState(complete)

      case startTask:
        currentTask.SetState(inProgress)

      case resetTask:
        currentTask.SetState(notStarted)

      case holdTask:
        currentTask.SetState(onHold)

      case setStartTime:
        fmt.Print("Enter start date and time (MM/DD/YYYY HH:MMam/pm): ")
        strTime, _ := reader.ReadString('\n')
        strTime = strings.TrimSpace(strTime)
        startTime, err := time.Parse("1/2/2006 3:04pm", strTime)
        if (err == nil) {
          // fmt.Printf("startTime = %s\n", startTime)
          // for now the date is always today
          //y, m, d := time.Now().Date()
          //h := startTime.Hour()
          //m := startTime.Minute()
          //fullStartTime := time.Date(y, m, d, h, m, 0, 0, nil)
          currentTask.SetTargetStartTime(&startTime)  
        } else {
          fmt.Printf("err = %s\n", err)
        }
        
      case setEstimate:
        fmt.Print("Enter estimate (e.g. 45m or 1h30m): ")
        strEstimate, _ := reader.ReadString('\n')
        strEstimate = strings.TrimSpace(strEstimate)
        estimate, err := time.ParseDuration(strEstimate)
        if (err == nil) {
          currentTask.SetEstimate(estimate)
        } else {
          fmt.Printf("err = %s\n", err)         
        }

      case debugTask:
        fmt.Printf("name = %s\n", currentTask.GetName())
        fmt.Printf("id = %s\n", currentTask.GetId())
        fmt.Printf("state = %s\n", currentTask.GetState())
        fmt.Printf("startTime = %s\n", currentTask.GetTargetStartTime())
        fmt.Printf("estimate = %s\n", currentTask.GetEstimate())
    }

    // most commands want us to reprint the entire list in
    // its most recent form
    if printListAfterCmd {
      fmt.Println(masterTask)
    }
  }
  
  // save my immediate sub-tasks (won't save grouping master task)
  err = masterTask.Save(true)
  if err != nil {
    log.Fatal(err)
  }     
}

// for now keep a global - temp as we build our API
// eventually we'll move this in somewhere else
var storage TaskDataMapper
var master *Task
var commands *commandHistory

var users Users

func initKnownUsers(tdm TaskDataMapper) (Users, error) {
  return tdm.UserLoadAll()

  /* - useful code to hang onto until we're sure we have users working to auto-create the superuser
  // allocate the slice to collect the users
  us := make(Users, 0)

  // until we have the yaml user file...
  // for now auto-create one user - the admin
  // TBD: make users a real thing
  admin, _ := NewUser("", "admin", "dblock@alumni.brown.edu", "insecure", nil)
  us = append(us, admin)
  return us, nil
  */
}

func runServerApp(port string, files string, certs string, dbName string) {
  log.Printf("Will run as server soon...\n")

  // initialize the backend storage mechanism requests
  tdm, err := initStorage(dbName)
  if err != nil {
    log.Fatal(err)
  }
  storage = tdm

  // load up all known users - do first since tasks reference users
  users, err = initKnownUsers(tdm)  

  // initialize a master task (in a global for now)
  master, err = initMasterTask(tdm)
  if err != nil {
    log.Fatal(err)
  } 

  // create an instance of our router with path to files
  router := NewRouter(files)
  
  // use built-in file server to serve our client application at /
  // TBD: integrate this into our router???
  log.Printf("...serving static pages from %s\n", files)

  // start the server itself
  log.Printf("...serving certificates from %s\n", certs)
  log.Printf("...listening on port%s\n", port)
  err = http.ListenAndServeTLS(port, certs + "self-signed.crt", certs + "server.key", router)
  if err != nil {
    log.Fatal(err)
  } 

}


// main: for now this is the console app - in the future input args will choose
// console app vs. web server
func main() {

  var server                bool
  var static_files_location string
  var certs_location        string
  var listenport            string
  var dbName                string
  flag.BoolVar(&server, "server", false, "start pim as web server rather than console app")
  flag.StringVar(&static_files_location, "html", "./client", "specify path to static web files on this server")
  flag.StringVar(&certs_location, "certs", ".", "specify path to TLS certificates on this server")
  flag.StringVar(&listenport, "port", "4000", "specify port on which the server will take requests")
  flag.StringVar(&dbName, "db", DB_NAME, "specify the database to use on the server or YAML")
  flag.Parse()

  // if we're starting as a server
  if (server) {

    // normalize file locations and port syntax so we can be flexible with
    // what the user types in
    if !strings.HasSuffix(certs_location, "/") {
      certs_location = certs_location + "/"
    }
    if strings.HasSuffix(static_files_location, "/") {
      static_files_location = strings.TrimSuffix(static_files_location, "/")
    }
    if !strings.HasPrefix(listenport, ":") {
      listenport = ":" + listenport
    }

    runServerApp(listenport, static_files_location, certs_location, dbName)

  } else {

    runConsoleApp(dbName)

  } // else we started app as a console app
} // main
