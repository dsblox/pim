package main

import (
	"testing"
)

// create, save and load a basic task - note that since we leverage the
// real database here we risk corrupting real data when running our tests
// and we need to be extra careful about cleaning up after ourselves.
// TBD: create a completely separate database for testing and fully
// initialize it before each test run.
func TestSingleTask(t *testing.T) {

	task_one := NewTask("Test Task Version One")
	if task_one == nil {
		t.Error("Could not test - unable to create a basic task")
		return
	}

	// initialize the database connection
    tdmpg1 := NewTaskDataMapperPostgreSQL(false)
    if tdmpg1 == nil {
    	t.Error("PIM-Testing requires a local PostgreSQL database to running.  Exiting...")
    	return
    }
    task_one.SetDataMapper(tdmpg1)

	// save the task without trying to save children
	err := task_one.Save(false)
	if err != nil {
		t.Error("Could not test - unable to save a basic task. err:", err)
	}

	// load a new version of the task
	task_two := NewTask("Test Task Version Two")
	if task_two == nil {
		t.Error("Could not test - unable to create a basic task")
		return
	}
    tdmpg2 := NewTaskDataMapperPostgreSQL(false)
    task_two.SetDataMapper(tdmpg2)
    task_two.id = task_one.Id() // never in real life, but useful for testing
    err = task_two.Load(false) // don't load children
    if err != nil {
    	t.Error("Failed to load simple task we just saved with id: ", task_one.Id(), "err: ", err)
    }

    if !task_one.DeepEqual(task_two) {
    	t.Error("Failed to save and load a simple Task.")
    }

    // clean up by removing
    err = task_one.Remove(nil)
    if err != nil {
    	t.Error("Failed to remove simple test task err: ", err)
    }

    // make sure it is gone
    task_two.id = task_one.Id()
    err = task_two.Load(false)
    if err == nil {
    	t.Error("Was able to load a task that should have been deleted: ", task_one.Id())
    }
}


