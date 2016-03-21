package main

import "fmt"
import "bufio"
import "os"
import "strings"
import "log"

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
	fmt.Println("  u = move current task up")
	fmt.Println("  d = move current task down")
	fmt.Println("  c = complete current task")
	fmt.Println("  s = start current task")
	fmt.Println("  r = reset current task to not started")
	fmt.Println("  h = put current task on hold")
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

// main: for now this is the console app - in the future input args will choose
// console app vs. web server
func main() {
    fmt.Printf("*** Welcome to PIM - The Task Manager for Your Life ***\n")

    // dummy master task to hold all tasks - this is for conveniece
    // and should not be printed out
    var masterTask *Task = &Task{name:"Your Console Task List", state:notStarted}

    // initialize the persistence layer - use PostgreSQL
    // and assign to masterTask - creating the first data
    // mapper initializes the DB - I chose to test a data-mapper
    // pattern to fully abstract the persistence layer from the
    // task functionality.  This should be the only place the
    // Tasks know how they are stored.
    tdm := NewTaskDataMapperPostgreSQL(-1)
    masterTask.SetDataMapper(tdm)

    // load the task list recursively
    masterTask.Load()

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

			case deleteTask: 
				// TBD: update data-mapper to remove things from the DB
				// we do updates on quit, but we should do deletes as
				// they happen - will need a delete op on data mapper
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
		}

		// most commands want us to reprint the entire list in
		// its most recent form
		if printListAfterCmd {
			fmt.Println(masterTask)
		}
	}
	
	// save my immediate sub-tasks (don't save grouping master task)
	err := masterTask.SaveChildren()
	if err != nil {
		log.Fatal(err)
	}		
}
