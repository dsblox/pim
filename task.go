package main

import "fmt"
import "errors"
import "github.com/satori/go.uuid"
import "time"


// didn't understand error types as values before - will rework
// this for proper Go error processing TBD
type ErrTaskNotFoundInList int
func (e ErrTaskNotFoundInList) Error() string {
	return fmt.Sprintf("Could not find task in parent or child list.")
}

// TaskState: enum type to track possible states of tasks
// along with UTF-8 characters to represent each state in the console app
type TaskState int
const (
	notStarted TaskState = iota 
	complete
	inProgress
	onHold
)
var stateChars = []rune{ '\u0020', '\u2713', '\u27a0', '\u2394'}
var stateStrings = []string{"notStarted", "complete", "inProgress", "onHold"} // for tests only
func (ts TaskState) String() string {
	return stateStrings[ts]
}

// for persistence we have a mapper interface that can
// be implemented differently depending on the backend
// we select.  Anyone that wishes to store Tasks can
// implement the methods in this interface with its own
// storage backend.
type TaskDataMapper interface {
	NewDataMapper() TaskDataMapper // return an empty mapper of your implementation type
	Save(t *Task, saveChildren bool, saveMyself bool) error            // save a task - just the task and parent relationships
	Load(t *Task, loadChildren bool, root bool) error 		   // load a task - and all its children (note lack of symmetry)
	Delete(t *Task, p *Task) error // delete a task - optionally reparenting its children
}

// Task: our central type for the whole world here - will become quite large over time
type Task struct {
	Id string  `json:"id"`        // unique id of the task - TBD make this pass through to mapper!!!
	Name string `json:"name"`     // name of the task
	State TaskState `json:"state"` // state of the task
	StartTime *time.Time `json:"startTime,omitempty"` // start time of the task
	Estimate time.Duration `json:"estimate"` // estimated duration of the task

	parents []*Task      // list of parent tasks (we support many parents)
	kids []*Task         // list of child tasks

	// for console app only!  hopefully won't need in the end
	current bool  // need to get rid of this
	currentParent bool // need to get rid of this too

	// data-mapper used to abstract persistence from this in-memory task object
	persist TaskDataMapper
	memoryonly bool // if true, don't allow this task to be saved

	// iterator support to make iterating with multiple parents sane
	// TBD move iteration to its own class for concurrent usage in the future
	iterCurrParent int
	iterCurrChild int
}
type Tasks []*Task

func (list Tasks) FindById(id string) *Task {
	for _, curr := range list {
		if id == curr.GetId() {
			return curr
		}
	}
	return nil
}



// NewTask: create a new task with a name, assign a unique id, and default settings
// When we break Tasks into its own package we will rename this to just "New()"
func NewTask(name string) *Task {
	return &Task{Id:uuid.NewV4().String(), Name:name, State:notStarted, memoryonly:false}
}

// an in-memory-only task that will never be saved - used to group other tasks
// so you can iterate over them or manipulate them, but designated never to be
// saved- note that it does have an id
func NewTaskMemoryOnly(name string) *Task {
	return &Task{Id:uuid.NewV4().String(), Name:name, State:notStarted, memoryonly:true}	
}

func (taska *Task) Equal(taskb *Task) bool {
	return taska.GetId() == taskb.GetId()
}
func (taska *Task) DeepEqual(taskb *Task) bool {
	return (taska.Id == taskb.Id && taska.State == taskb.State && taska.Name == taskb.Name)
	// tbd: run child and parent lists and at least ensure their ids match
	// avoid recursing here as you'll end up running the entire tree in both directions
}

// SetDataMapper: assign an implementation of data mapper to persist this task
// or (if memory only) all child tasks that are no memoryonly
func (t *Task) SetDataMapper(tdm TaskDataMapper) {
	t.persist = tdm
}

// GetDataMapper: get currently assigned data mapper if any
func (t *Task) DataMapper() TaskDataMapper {
	return t.persist
}

// stateChar: map from current state to the UTF-8 code for console display of the state
func (t Task) stateChar() string {
	return string(stateChars[t.State])
}

func RenderTime(t* time.Time) string {
  if t != nil {
  	return t.Format("20060102-15:04:05.000")
  } else {
  	return ""
  }
}

// StringSingle: display an individual task indented to the requested level
// and highlighting the current task with an "*" 
func (t Task) StringSingle(level int) string {
	var s string
	for i := 0; i < level; i++ {
		s += "   "
	}
	// for console app only - note the console app controls
	// current task - the task object makes no attempt to
	// make sure there is only one task marked current
	if t.current {
		s += "*"
	} else {
		s += " "
	}
	s += fmt.Sprintf("[%v] <%v> %v <%d> (%d sub-tasks)", t.stateChar(), RenderTime(t.StartTime), t.Name, t.Estimate, len(t.kids))
	return s
}

// StringChildren recurses to print the task
// hierarchy below the current task starting
// from the indent level specified.  It is 
// separate from StringHierarchy to allow clients
// to print their children without including the
// parent task.
func (t Task) StringChildren(level int) string {
	var s string
	for _, k := range t.kids {
		s += "\n" + k.StringHierarchy(level + 1)
	}
	return s		
}

// StringHierarhcy recurses via StringChildren
// to print the entire hierarchy below this
// task, beginning the indenting at the level
// specified
func (t Task) StringHierarchy(level int) string {
	s := t.StringSingle(level)
	s += t.StringChildren(level)
	return s	
}

// Stringer-invoked String function uses the hierarhcy
// recursion to print the full list of tasks, always
// beginning indenting from zero.
func (t Task) String() string {
	return t.StringHierarchy(0)
}

// SetId - comment out - id cannot be changed
/*
func (t *Task) SetId(newId string) {
	t.id = newId
} */

// Id
func (t *Task) GetId() string {
	return t.Id
}

// SetState: sets the task state
func (t *Task) SetState(newState TaskState) {
	t.State = newState
}

// State: returns the current state of the task
func (t *Task) GetState() TaskState {
	return t.State
}

// SetName: sets the name field
func (t *Task) SetName(newName string) {
	t.Name = newName
}

// Name: returns the name field
func (t *Task) GetName() string {
	return t.Name
}

func (t *Task) SetStartTime(start *time.Time) {
	t.StartTime = start;
}
func (t *Task) GetStartTime() *time.Time {
	return t.StartTime;
}
func (t *Task) SetEstimate(estimate time.Duration) {
	t.Estimate = estimate;
}
func (t *Task) GetEstimate() time.Duration {
	return t.Estimate;
}

// IsMemoryOnly: returns memory-only state indicating
// if this task should ever be saved or read from storage
func (t *Task) IsMemoryOnly() bool {
	return t.memoryonly
}

func (t *Task) Kids() Tasks {
	return t.kids
}

// removeTaskFromSlice: worker function to remove from slice
// note: not part of the Task object
func removeTaskFromSlice(s []*Task, i int) []*Task {
	return append(s[:i], s[i+1:]...)
}

// findTaskInSlice: returns the index of the matching (identical task memory)
// task within a slice of tasks
func findTaskInSlice(s []*Task, f *Task) int {
	for i, curr := range s {
		if f == curr {
			return i
		}
	}
	return -1
}

// findChild: returns index of where the task provided lives in the child list
// or -1 if not in the list
func (t Task) findChild(k *Task) int {
	return findTaskInSlice(t.kids, k)
}

// findParent: returns index of where the task provided lives in the parent list
// or -1 if not in the list
func (t Task) findParent(p *Task) int {
	return findTaskInSlice(t.parents, p)
}

// NumParents: returns the number of parents on a task
func (t Task) NumParents() int {
	return len(t.parents)
}

// NumChildren: returns the number of children on a task
func (t Task) NumChildren() int {
	return len(t.kids)
}

// HasParents: return true if the task has any parents
func (t Task) HasParents() bool {
	return t.NumParents() > 0
}

// HasChildren: returns true of the task has any kids
func (t Task) HasChildren() bool {
	return t.NumChildren() > 0
}

func (t Task) FindChild(id string) *Task {
	for _, curr := range t.kids {
		if id == curr.GetId() {
			return curr
		}
	}
	return nil
}

func (t Task) FindParent(id string) *Task {
	for _, curr := range t.parents {
		if id == curr.GetId() {
			return curr
		}
	}
	return nil	
}


// CurrentParent: returns the first parent found with its
// current flag set to true.  Used by UIs to track a
// particular path through the parent / child chain
// returns nil if no parent is marked as current
func (t *Task) CurrentParent() *Task {
	for _, curr := range t.parents {
		if curr.current {
			return curr
		}
	}
	return nil
}


// IterChild: returns the child delta number of
// items from the last returned child.  Provide
// delta = 0 to get the first child.

// IterTasks: worker that operates over any slice of tasks
// and given a current index i will return the task ahead
// or behind i by that delta slots.  Since it is assumed to
// be used for iteration, any delta of zero indicates that
// we should return the first item in the list and "reset"
// iteration to the front of the list
func IterTasks(tasks []*Task, i int, delta int) (int, *Task) {
	var numTasks = len(tasks)
	var j int

	// zero means to reset iterator to start of list
	// and anything else means to jump that distince
	// ahead or behind (delta can be negative)
	if delta == 0 {
		j = 0
	} else {
		j = i + delta
	}

	// if we ended up outside the list boundaries then return nil
	// and an unchanged new index.  Note this also catches the
	// case of an empty list because numTasks will be 0 here
	if j < 0 || j >= numTasks {
		return i, nil
	}

	// return the new index and the task at that index
	return j, tasks[j]
}

func (t *Task) IterChild(delta int) *Task {
	var nextChild *Task = nil
	t.iterCurrChild, nextChild = IterTasks(t.kids, t.iterCurrChild, delta)
	return nextChild
}

func (t *Task) IterParent(delta int) *Task {
	var nextParent *Task = nil
	t.iterCurrParent, nextParent = IterTasks(t.parents, t.iterCurrParent, delta)
	return nextParent
}

func (t *Task) FirstChild() *Task {
	return t.IterChild(0)
}

func (t *Task) NextChild() *Task {
	return t.IterChild(1)
}

func (t *Task) PrevChild() *Task {
	return t.IterChild(-1)
}

func (t *Task) FirstParent() *Task {
	return t.IterParent(0)
}

func (t *Task) NextParent() *Task {
	return t.IterParent(1)
}

func (t *Task) PrevParent() *Task {
	return t.IterParent(-1)
}


// PrevSibling: return the previous sibling from the specified
// parent.  If no parent is provided then we first attempt to
// find a current parent.  If no current parent is found then
// we return the previous sibling from first parent 
// (useful if you've only built a single parent hierarchy).
func (t *Task) PrevSibling(parent *Task) *Task {
	if t.HasParents() {
		if parent == nil {
			parent = t.CurrentParent()
		}
		if parent == nil {
			parent = t.FirstParent()
		}
		myIdx := parent.findChild(t)
		if myIdx >= 1 { // if i'm the first child return nil
			return parent.kids[myIdx-1]
		}
	}
	return nil
}

// NextSibling: return my next sibling.  Since I may have multiple
// parents you can specify which parent's sibling list should be
// traversed.  If nil is specified we take a good guess:
//   - first we check for a "current" parent to traverse - current is 
//     a helper function we provide to allow UIs to track a hierarchy path
//   - if that fails we'll just traverse the list of the first parent we
//     see (useful if the callers of this are making a single parent hierarchy)
func (t *Task) NextSibling(parent *Task) *Task {
	if t.HasParents() {
		if parent == nil {
			parent = t.CurrentParent()
		}
		if parent == nil {
			parent = t.FirstParent()
		}
		myIdx := parent.findChild(t)
		if myIdx < len(parent.kids)-1 { 
			return parent.kids[myIdx+1]
		}
	}
	return nil
}

/*
// PrevParent: return my previous parent.
// parent.  If no parent is provided then we assume a single-parent
// heriarchy and simply use the first parent as the only parent.
func (t *Task) PrevParent(child *Task) *Task {
	if child == nil {
		return nil
	}
	if t.HasChildren() {
		myIdx := child.findParent(t)
		if myIdx >= 1 { // if i'm the first parent return nil
			return child.parents[myIdx-1]
		}
	}
	return nil
}

// NextParent: return the next parent from the specified
// child.  Child task to match against must be specified.
func (t *Task) NextParent(child *Task) *Task {
	if child == nil {
		return nil
	}
	if t.HasChildren() {
		myIdx := child.findParent(t)
		if myIdx < len(child.parents)-1 { 
			return child.parents[myIdx+1]
		}
	}
	return nil
}
*/

// for now we'll assume a single parent - take first one
func (t *Task) Parent() *Task {
	return t.FirstParent()
}

func (t *Task) AddParent(p *Task) error {
	t.parents = append(t.parents, p)
	p.kids = append(p.kids, t)

	// if either is missing a mapper then create it
	if t.persist == nil && p.persist != nil {
		t.persist = p.persist.NewDataMapper()
	}
	if p.persist == nil && t.persist != nil {
		p.persist = t.persist.NewDataMapper()
	}
	return nil
}

func (t *Task) AddChild(k *Task) error {
	return k.AddParent(t)
}

func (t *Task) removeParent(p *Task, bParentExpected bool, bChildExpected bool) error {
	indexParent := t.findParent(p)
	if indexParent >= 0 {
		t.parents = removeTaskFromSlice(t.parents, indexParent)
	} else if bParentExpected {
		return errors.New("pim: parent to remove not found on task")
	}

	indexChild := p.findChild(t)
	if indexChild >= 0 {
		p.kids = removeTaskFromSlice(p.kids, indexChild)
	} else if bChildExpected {
		return errors.New("pim: child to remove not found on task")	
	}

	return nil
}

func (t *Task) RemoveParent(p *Task) error {
	return t.removeParent(p, true, false)
}


func (t *Task) RemoveChild(k *Task) error {
	return k.removeParent(t, false, true)
}

// Remove: remove me from my parent's child lists and from any children's parent lists
// if my children are orphaned, optinally make them children of a supplied parent
func (t *Task) Remove(newParent *Task) error {

	// delete from storage before we remove from
	// memory since the data-mapper will need the
	// loaded object to find the persisted data to delete.
	// don't bother with reparenting my kids - too
	// hard to coordinate with the logic of only doing
	// so for orphans, and reparenting should get
	// fixed on the next save (will it?)
	if t.persist != nil {
		t.persist.Delete(t, nil)
	}

	// remove from parent's child lists
	for _, p := range t.parents {
		p.RemoveChild(t)
	}

	// remove from kids parent lists
	// replacing with new parent if specified
	// and child will be orphaned otherwise
	for _, k := range t.kids {
		k.RemoveParent(t)
		if newParent != nil && !k.HasChildren() {
			k.AddParent(newParent)
		}
	}
	return nil
}

/*
=================================================================================
 Task.Save()
---------------------------------------------------------------------------------
 Inputs: saveChildren bool - true: recursively save all children
                           - false: save just this task and parent relationships

 Save the task to storage and (optionally) all of its children.  It is important
 to note that the "storage unit" assumed is the task itself and all of this
 task's parent relationships.

 Save is where we enforce the memory-only setting on a task, but it is ONLY
 ENFORCED AT THE TOP LEVEL - Mappers are free to ignore the setting for children 
 in the middle of the hierarhcy!

 Note that we leave the implementation of the recursion to the Mapper so it
 can take advantage of clever ways to load objects in bulk from whatever its
 storage mechanism is.     
===============================================================================*/
func (t *Task) Save(saveChildren bool) error {

	// log.Printf("Task.Save() id:%s, name:%s, memoryonly:%t\n", t.Id(), t.Name(), t.memoryonly)

	// if memory only and not saving children then nothing to do
	if t.memoryonly && !saveChildren {
		return nil
	}

	// save my children if requested, and save myself if not memory only
	err := t.persist.Save(t, saveChildren, !t.memoryonly)
	if err != nil {
		return err
	}

	return nil
}



/*
=================================================================================
 Task.Load()
---------------------------------------------------------------------------------
 Inputs: loadChildren bool - true to recursively load all children
                           - false loads the task (not its parent relationships!)

 Load the task with this id from storage and (optionally) all of its children.  
 Note the versatility of this function based on the kind / state of this task.
   - if this task is memory-only then it will simply attach "root" tasks
     as children to this task (and their children if specified)
   - if this task is not memory-only then it will load from storage any
     task with a matching id (and its children if specified)

 If loadChildren is not specified then no parent relationships are loaded.
 This is because the assumption is always that tasks are loaded "top down"
 so it makes no sense to load parent tasks "bottom up".

 Note that we leave the implementation of the recursion to the Mapper so it
 can take advantage of clever ways to load objects in bulk from whatever its
 storage mechanism is.
===============================================================================*/
func (t *Task) Load(loadChildren bool) error {
	return t.persist.Load(t, loadChildren, t.memoryonly)
}

