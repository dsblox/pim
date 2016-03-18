package main

import "fmt"
import "bufio"
import "os"
import "strings"
import "database/sql"
import "log"

// global variable to hold a DB connection
type Env struct {
    db *sql.DB
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
)
var cmdChars = []rune{ '?', 'q', 'p', 'a', 'x', 'u', 'd', 'c', 's', 'r', 'h'}

func findCommand(entered rune) int {
	for i, c := range cmdChars {
		if entered == c {
			return i
		}
	}
	return help
}

func printHelp() {
	fmt.Println("PIM Console Help")
	fmt.Println("  p = print task list")
	fmt.Println("  a = add task as child of current task")
	fmt.Println("  x = delete current task")
	fmt.Println("  u = move current task up")
	fmt.Println("  d = move current task down")
	fmt.Println("  c = complete current task")
	fmt.Println("  s = start current task")
	fmt.Println("  r = reset current task to not started")
	fmt.Println("  h = put current task on hold")
	fmt.Println("  ? = help")
	fmt.Println("  q = quit")
}

func moveUp(oldCurrentTask *Task) *Task {
	var newCurrentTask = oldCurrentTask.PrevSibling()
	if newCurrentTask == nil {
		newCurrentTask = oldCurrentTask.Parent()
	}
	if newCurrentTask != nil {
		oldCurrentTask.current = false
		newCurrentTask.current = true
	} else {
		newCurrentTask = oldCurrentTask
	}
	return newCurrentTask // may be unchanged if at top of list
}

func moveDown(oldCurrentTask *Task) *Task {
	var newCurrentTask = oldCurrentTask.FirstChild()
	if (newCurrentTask == nil) {
		newCurrentTask = oldCurrentTask.NextSibling()
		if newCurrentTask == nil && oldCurrentTask.HasParents() {
			newCurrentTask = oldCurrentTask.Parent().NextSibling()
		}
	}
	if newCurrentTask != nil {
		oldCurrentTask.current = false
		newCurrentTask.current = true
	} else {
		newCurrentTask = oldCurrentTask
	}
	return newCurrentTask
}

func dbExec(env *Env, sqlStr string, args ...interface{}) (sql.Result, error) {
	result, err := env.db.Exec(sqlStr, args...)
	if err != nil {
		log.Fatal(err)
	}
	return result, err
}

// returns id of inserted row if no error
func dbInsert(env *Env, sqlStr string, args ...interface{}) (int, error) {
	var id int
	err := env.db.QueryRow(sqlStr, args...).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}
	return id, err
}

// note the weirdness that we want the app - not the database
// to hold the top-level grouping-task, so bTop used only at
// the first level can be used to not save the first task.
func saveTask(env *Env, t *Task, parentId int, bTop bool) error {

	var myId int = -1
	if (!bTop) {

		// save myself - with null parent id if none specified
		// but save my own id so I can pass it on to my kids
		var err error
		if (parentId != -1) {
			myId, err = dbInsert(env, "INSERT INTO tasks (name, state, parent_id) VALUES ($1, $2, $3)", t.name, t.state, parentId)
		} else {
			myId, err = dbInsert(env, "INSERT INTO tasks (name, state, parent_id) VALUES ($1, $2, NULL)", t.name, t.state)
		}
		if err != nil {
			log.Fatal(err)
		}

	}

	// save my children
	for curr := t.FirstChild(); curr != nil; curr = curr.NextSibling() {
		err := saveTask(env, curr, myId, false)
		if err != nil {
			log.Fatal(err)
		}		
	}
	return nil		
}

// TBD - find a way to load the tasks - this will probably
// require that we start to store the task id on the task
// object.  Probably time to just make the object's
// persistence some kind of mixin to the Task
func loadTask(env *Env, t *Task, bTop bool) *Task, error {

	return t, nil
}

func main() {
    fmt.Printf("*** Welcome to PIM - The Task Manager for Your Life ***\n")

/*
    t1 := Task{name:"Clean dishes", state:notStarted}
    t2 := Task{name:"Write program", state:inProgress}

    t3 := Task{name:"Organize for work", state:onHold }
    t3a := Task{name:"Weekly Planning", state:onHold}
    t3b := Task{name:"Review Calendar", state:notStarted}
    t3c := Task{name:"Schedule Robert Catchup", state:notStarted}
    t3.AddChild(&t3a)
    t3.AddChild(&t3b)
    t3.AddChild(&t3c)

    t4 := Task{name:"Have breakfast", state:complete }

    fmt.Printf("\nFirst List...\n")
    fmt.Println(t1)
    fmt.Println(t2)
    fmt.Println(t3)
    fmt.Println(t4)

    e := t3.RemoveChild(&t3b)
    if (e != nil) {
    	fmt.Println(e)
    }
    t4.AddChild(&t3b)

    fmt.Printf("\nSecond List...\n")
    fmt.Println(t1)
    fmt.Println(t2)
    fmt.Println(t3)
    fmt.Println(t4)
    */

    var masterTask *Task = &Task{name:"Your Console Task List", state:notStarted}


    // initialize the database and hold in global variable env
    db, err := NewDB("postgres://postgres:postgres@localhost/pim?sslmode=disable")
    if err != nil {
        log.Panic(err)
    }
    env := &Env{db: db}

    // test DB - see how much data is in task table
	var (
		id int
		name string
		state TaskState
	)
	rows, err := env.db.Query("select id, name, state from tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name, &state)
		if err != nil {
			log.Fatal(err)
		}
		t := Task{name:name, state:state}
		masterTask.AddChild(&t)
		log.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}    


    var currentTask *Task = masterTask
    currentTask.current = true
	printHelp()


	reader := bufio.NewReader(os.Stdin)
    for cmd := help; cmd != quit;  {

		fmt.Print("Command: ")
		text, _ := reader.ReadString('\n')
		cmd = findCommand(([]rune(text))[0])
		switch cmd {
			case help: 
				printHelp()

			case printList: 
				fmt.Println(masterTask)

			case quit: 
				fmt.Println("Goodbye!")

			case upCurrent: 
				currentTask = moveUp(currentTask)
				fmt.Println(masterTask)

			case downCurrent: 
				currentTask = moveDown(currentTask)
				fmt.Println(masterTask)

			case addTask: 
				fmt.Print("Enter name of new task: ")
				taskName, _ := reader.ReadString('\n')
				currentTask.AddChild(&Task{name:strings.TrimSpace(taskName), state:notStarted})
				fmt.Println(masterTask)

			case deleteTask: 
				taskToKill := currentTask
				currentTask = moveUp(currentTask)
				if (!taskToKill.HasParents()) {
					fmt.Println("Can't delete top level task list")
				} else {
					taskToKill.Remove(masterTask)
					fmt.Println(masterTask)
				}

			case completeTask:
				currentTask.SetState(complete)
				fmt.Println(masterTask)

			case startTask:
				currentTask.SetState(inProgress)
				fmt.Println(masterTask)

			case resetTask:
				currentTask.SetState(notStarted)
				fmt.Println(masterTask)

			case holdTask:
				currentTask.SetState(onHold)
				fmt.Println(masterTask)
		}
	}

	// when quitting - clear tasks table and save current tasks
	/*
	_, err = env.db.Exec("TRUNCATE TABLE tasks")
	if err != nil {
		log.Fatal(err)
	} */
	
	// save my sub-tasks and their sub-tasks
	err = saveTask(env, masterTask, -1, true)
	if err != nil {
		log.Fatal(err)
	}		
}
