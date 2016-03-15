package main

import "fmt"
import "bufio"
import "os"
import "strings"

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
}
