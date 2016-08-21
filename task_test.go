package main

import (
	"testing"
	"regexp"
	"strconv"
)

func validUUIDv4(id string) bool {
    r := regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[8|9|aA|bB][a-f0-9]{3}-[a-f0-9]{12}$")
    return r.MatchString(id)
}

func validateDefaultTask(task *Task, expectedName string, t *testing.T) {
	if task == nil {
		t.Error("Failed to even create a task - nill returned")
	}
	if !validUUIDv4(task.Id()) {
		t.Error("Task created with unexpected uiid value: ", task.Id())
	}
	if task.Name() != expectedName {
		t.Error("Task name expected ", expectedName, " but found: ", task.Name())
	}
	if task.State() != notStarted {
		t.Error("Task state expected <notStarted> but found: ", task.State())
	}
}

// basic simple task is created with correct default values and can be set / get
func TestTaskCreate(t *testing.T) {
	task := NewTask("Test Task")
	validateDefaultTask(task, "Test Task", t)

	task.SetName("Test Task Renamed")
	if task.Name() != "Test Task Renamed" {
		t.Error("Task rename failed, expected <Test Task Rename> but found: ", task.Name())
	}
	task.SetName("")
	if task.Name() != "" {
		t.Error("Task rename failed, expected empty name but found: ", task.Name())
	}

	task.SetState(complete)
	if task.State() != complete {
		t.Error("Task SetState() failed, expected <complete> but found: ", task.State())
	}
}


// multi-parent/multi-child relationship can be created and iterated and removed
func TestTaskHierarchy(t *testing.T) {
	// create the parent
	parent := NewTask("Parent Task")
	validateDefaultTask(parent, "Parent Task", t)
	if parent.HasChildren() {
		t.Error("Parent should not have children until we've added them.")
	}

	// create 5 children and hook each one on
	childrenToTest := 5
	for i := 0; i < childrenToTest; i++ {
		childTaskName := "Child Task #" + strconv.Itoa(i+1)
		currChild := NewTask(childTaskName)
		validateDefaultTask(currChild, childTaskName, t)
		parent.AddChild(currChild)
		if currChild.Parent() != parent {
			t.Error("Child", childTaskName, "not correctly linked to parent task.")
		}
	}

	// validate all kids were connected
	if parent.NumChildren() != childrenToTest {
		t.Error("Incorrect number of children found linked to parent.")
	}

	// validate that we can iterate over all the kids
	i := childrenToTest
	for c := parent.FirstChild(); c != nil; c = parent.NextChild() {
		i -= 1
	}
	if i != 0 {
		t.Error("Unable to iterate over all children linked to parent.")
	}

	// validate we can remove two of the children without reparenting
	c := parent.FirstChild();
	c.Remove(nil)
	c = parent.FirstChild();
	c.Remove(nil)
	if parent.NumChildren() != childrenToTest - 2 {
		t.Error("Unable to remove children from parent.")
	}

	// create 5 grand-parents
	grandparentsToTest := 5
	for i := 0; i < childrenToTest; i++ {
		grandTaskName := "Grandparent Task #" + strconv.Itoa(i+1)
		currGrand := NewTask(grandTaskName)
		validateDefaultTask(currGrand, grandTaskName, t)
		parent.AddParent(currGrand)
		if currGrand.FirstChild() != parent {
			t.Error("Grandparent", grandTaskName, "not correctly linked to parent task.")
		}
	}

	// validate all grand-parents were connected
	if parent.NumParents() != grandparentsToTest {
		t.Error("Incorrect number of grandparents found linked to parent.")
	}

	// validate that we can iterate over all the grandparents
	i = grandparentsToTest
	for g := parent.FirstParent(); g != nil; g = parent.NextParent() {
		i -= 1
	}
	if i != 0 {
		t.Error("Unable to iterate over all grandparents linked to parent.")
	}

	// validate we can remove two of the parents
	g := parent.FirstParent();
	g.Remove(nil)
	g = parent.FirstParent();
	g.Remove(nil)
	if parent.NumParents() != grandparentsToTest - 2 {
		t.Error("Unable to remove grandparents from parent.")
	}

	// validate we can reparent children when removing ourselves
	grandParent := parent.FirstParent()
	idToCheck := parent.FirstChild().Id()
	parent.Remove(grandParent)
	if grandParent.NumChildren() != childrenToTest - 2  || grandParent.FindChild(idToCheck) == nil {
		t.Error("Reparenting failed when removing a task with children.")
	}

}