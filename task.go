package main

import "fmt"
import "errors"

type ErrTaskNotFoundInList int
func (e ErrTaskNotFoundInList) Error() string {
	return fmt.Sprintf("Could not find task in parent or child list.")
}

type TaskState int
const (
	notStarted TaskState = iota 
	complete
	inProgress
	onHold
)
var stateChars = []rune{ '\u0020', '\u2713', '\u27a0', '\u2394'}

type Task struct {
	name string
	state TaskState
	parents []*Task
	kids []*Task

	// for console app only!  hopefully won't need in the end
	current bool
}

func (t Task) stateChar() string {
	return string(stateChars[t.state])
}

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

func (t Task) StringHierarchy(level int) string {
	s := t.StringSingle(level)
	for _, k := range t.kids {
		s += "\n" + k.StringHierarchy(level + 1)
	}
	return s	
}

func (t Task) String() string {
	return t.StringHierarchy(0)
}

func (t *Task) SetState(newState TaskState) {
	t.state = newState
}

func removeTaskFromSlice(s []*Task, i int) []*Task {
	return append(s[:i], s[i+1:]...)
}

func findTaskInSlice(s []*Task, f *Task) int {
	for i, curr := range s {
		if f == curr {
			return i
		}
	}
	return -1
}

func (t Task) findChild(k *Task) int {
	return findTaskInSlice(t.kids, k)
}

func (t Task) findParent(p *Task) int {
	return findTaskInSlice(t.parents, p)
}

func (t Task) HasParents() bool {
	return len(t.parents) > 0
}

func (t Task) HasChildren() bool {
	return len(t.kids) > 0
}

func (t *Task) FirstChild() *Task {
	if t.HasChildren() {
		return t.kids[0]
	}
	return nil
}

// for now we'll assume a single parent - take first one
func (t *Task) Parent() *Task {
	if t.HasParents() {
		return t.parents[0]
	}
	return nil
}
func (t *Task) PrevSibling() *Task {
	if t.HasParents() {
		myIdx := t.Parent().findChild(t)
		if myIdx >= 1 { // if i'm the first child return nil
			return t.Parent().kids[myIdx-1]
		}
	}
	return nil
}
func (t *Task) NextSibling() *Task {
	if t.HasParents() {
		myIdx := t.Parent().findChild(t)
		if myIdx < len(t.Parent().kids)-1 { 
			return t.Parent().kids[myIdx+1]
		}
	}
	return nil
}


func (t *Task) AddParent(p *Task) error {
	t.parents = append(t.parents, p)
	p.kids = append(p.kids, t)
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

// remove me from my parent's child lists
// and from any children's parent lists
// if my children are orphaned, optinally make them
// children of a supplied parent
func (t *Task) Remove(newParent *Task) error {
	// remove from parent's child list - assumes 1 parent! ???
	if t.HasParents() {
		t.Parent().RemoveChild(t)
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


