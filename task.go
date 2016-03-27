package main

import "fmt"
import "errors"

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

// for persistence we have a mapper interface that can
// be implemented differently depending on the backend
// we select.  Anyone that wishes to store Tasks can
// implement the methods in this interface with its own
// storage backend.
type TaskDataMapper interface {
	NewDataMapper() TaskDataMapper // return an empty mapper of your implementation type
	Save(t *Task) error            // save a task - just the task and parent relationships
	Load(t *Task) error 		   // load a task - and all its children (note lack of symmetry)
	Delete(t *Task, p *Task) error // delete a task - optionally reparenting its children
}

// Task: our central type for the whole world here - will become quite large over time
type Task struct {
	name string          // name of the task
	state TaskState      // state of the task
	parents []*Task      // list of parent tasks (we support many parents)
	kids []*Task         // list of child tasks

	// for console app only!  hopefully won't need in the end
	current bool  // need to get rid of this
	currentParent bool // need to get rid of this too

	// data-mapper used to abstract persistence from this in-memory task object
	persist TaskDataMapper

	// iterator support to make iterating with multiple parents sane
	// TBD move iteration to its own class for concurrent usage in the future
	iterCurrParent int
	iterCurrChild int
}

// NewTask: create a new task with a name and default settings
// When we break Tasks into its own package we will rename this to just "New()"
func NewTask(name string) *Task {
	return &Task{name:name, state:notStarted}
}

// SetDataMapper: assign an implementation of data mapper to persist this task
func (t *Task) SetDataMapper(tdm TaskDataMapper) {
	t.persist = tdm
}

// GetDataMapper: get currently assigned data mapper if any
func (t *Task) DataMapper() TaskDataMapper {
	return t.persist
}

// stateChar: map from current state to the UTF-8 code for console display of the state
func (t Task) stateChar() string {
	return string(stateChars[t.state])
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
	s += fmt.Sprintf("[%v] %v (%d sub-tasks)", t.stateChar(), t.name, len(t.kids))
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

// SetState: sets the task state
func (t *Task) SetState(newState TaskState) {
	t.state = newState
}

// State: returns the current state of the task
func (t *Task) State() TaskState {
	return t.state
}

// SetName: sets the name field
func (t *Task) SetName(newName string) {
	t.name = newName
}

// Name: returns the name field
func (t *Task) Name() string {
	return t.name
}

// removeTaskFromSlive: worker function to remove from slice
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

// HasParents: return true if the task has any parents
func (t Task) HasParents() bool {
	return len(t.parents) > 0
}

// HasChildren: returns true of the task has any kids
func (t Task) HasChildren() bool {
	return len(t.kids) > 0
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
	if t.persist == nil {
		t.persist = p.persist.NewDataMapper()
	}
	if p.persist == nil {
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
	t.persist.Delete(t, nil)

	// remove from parent's child lists
	for _, p := range t.parents {
		p.RemoveChild(t)
	}

	// remove from kids parent lists
	// replacing with new parent if specified
	// and child would be orphaned otherwise
	for _, k := range t.kids {
		k.RemoveParent(t)
		if newParent != nil && !k.HasChildren() {
			k.AddParent(newParent)
		}
	}
	return nil
}

// Save: saves the task with whatever persistence has been set on the
// task via the DataMapper
func (t *Task) Save() error {
	// save myself
	err := t.persist.Save(t)
	if err != nil {
		return err
	}

	// recurse to save my children - SaveChildren will call Save()
	return t.SaveChildren()
}

/*
// Load reads in a task from the DB as long as the task's
// persistent_id has been set.  This function will replace
// the task's current in-memory information with what is
// in the database - will this be rare or will it recurse
// based on finding parent ids in the DB?
func (t *Task) Load(env *Env) error {

	// if no persistent_id then error
	if t.persistent_id < 0 {
		return errors.New("task.Load(): no identifier provided to load from")
	}

	// query for the item
	err := env.db.QueryRow("SELECT name, state FROM tasks where id = $1", t.persistent_id).Scan(&t.name, &t.state)
	if err != nil {
		log.Fatal(err)
	}

	return nil
} */

// SaveChildren only saves the children, which is a good and
// charitable thing to do.  Note that it does not save the
// grouping parent task.
func (t *Task) SaveChildren() error {
	var err error = nil
	for c := t.FirstChild(); c != nil && err == nil; c = t.NextChild() {
		err = c.Save()
	}
	return err
}

// needed? 
/*
func LoadTasks(env *Env, tasks []*Task) error {
	var err error = nil
	for _, t := range tasks {
		err = t.Load(env)
	}
	return err
} */

// load all tasks and add as children of parent task
// eventually we'll adjust this to load all
// tasks that are children of a particular
// parent task
func (t *Task) Load() error {
	return t.persist.Load(t)
}

