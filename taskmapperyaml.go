package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)


// TaskDataMapperYAML implements TaskDataMapper to persist tasks
type TaskDataMapperYAML struct {
	fileName string
}

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.  Note that for YAML we simply
// support a straight list of tasks for now
type TaskYAML struct {
	Id string
	Name string
	State string					// notStarted, completed, inProgress, onHold
	TargetStartTime *time.Time 		// targeted start time of the task
	ActualStartTime *time.Time 		// actual start time of the task
	ActualCompletionTime *time.Time // time task is marked done
	Estimate int					// estimate for the task in minutes
	Tags []string 			 		// attributes of the task
	Parents []string 		 		// ids of the parents for later hookup
}
type TasksYAML struct {
	Tasks []TaskYAML
}

// TBD: 8/16/16...
// currently all these unique ID functions are in the task mapper
// but the ID should not be unique to the persistence layer, rather
// it should be pushed up to the task object.  I believe all we'll
// need to do is to move the accessor and settor functions, and
// perhaps change the code that "pulls" the Postgres assigned id out
// since it should no longer be needed once we're setting the id
// as the primary key - we may also have to change the DB to use a
// string as the primary key instead of a number.
// you see the original code assumed the DB would assign an ID and
// therefore a task had no ID until it was initially saved.  We're
// about to change all that.  So far, all I've done is added the UUID
// to the Task object in task.go.  (I've not even set up to properly
// download and go get the uuid library)

func NewTaskDataMapperYAML(fileName string) *TaskDataMapperYAML {
	return &TaskDataMapperYAML{fileName:fileName}
}


// NewDataMapper() implements for YAML the ability to return a new
// and unpersisted instance of the mapper that can later be filled in
// by the object with the list of saved parent ids
func (tm TaskDataMapperYAML) NewDataMapper(fileName string) TaskDataMapper {
	return &TaskDataMapperYAML{fileName:fileName}
}

// not sure anymore what CopyDataMapper is for - so this implementation may be wrong
func (tm TaskDataMapperYAML) CopyDataMapper() TaskDataMapper {
	return &TaskDataMapperYAML{fileName:tm.fileName}
}

func TimeYAML(t* time.Time) string {
  if t != nil {
  	// return t.Format("2006-01-02T15:04:05-0700")
  	return t.Format(time.RFC3339)
  } else {
  	return "null"
  }
}

func (tm *TaskDataMapperYAML) writeTask(f *os.File , t *Task) error {
	var parentIds []string = t.GetParentIds(false)
	var estimate int = int(t.Estimate.Minutes())
	var tags []string = t.GetTags()
	strTags := ""
	if len(tags) > 0 {
		strTags = "'" + strings.Join(tags, "', '") + "'"
	}

	_, err := fmt.Fprintf(f, "- {id: %s, parents: %v, name: %s, state: %s, estimate: %d, tags: [%s], targetstarttime: %s, actualstarttime: %s, actualcompletiontime: %s }\n", 
		                  t.GetId(), parentIds, t.GetName(), t.GetState(), estimate, strTags, 
		                  TimeYAML(t.GetTargetStartTime()),
		              	  TimeYAML(t.GetActualStartTime()),
		              	  TimeYAML(t.GetActualCompletionTime()))
	return err
}

func (tm *TaskDataMapperYAML) saveTask(f *os.File, t *Task) error {
	var err error = nil

	// unless i'm the master root, save myself
	if t.HasParents() {
		err = tm.writeTask(f, t)
		if err != nil {
			return err
		}
	}

	// now save all my kids
	for c := t.FirstChild(); c != nil && err == nil; c = t.NextChild() {
		err = tm.saveTask(f, c)
	}
	if err != nil {
		return err
	}

	return err
}



// dump our tasks into the YAML file
func (tm *TaskDataMapperYAML) Save(t *Task, saveChildren bool, saveMyself bool) error {

	// log.Printf("Save(%t, %t): task = %s, id = %s, loaded = %t len(parentIds) = %d", saveChildren, saveMyself, t.name, t.id, tm.loaded, len(tm.parentIds))

	// open the file and clear it's contents if it exists
    f, err := os.Create(tm.fileName)
    if err != nil {
    	log.Printf("unable to create YAML file: %s\n", tm.fileName)
    	return err
    }

    // initialize the file with the header
    _, err = fmt.Fprintln(f, "tasks:")
	if err != nil {
    	log.Printf("unable to write header to YAML file: %s\n", tm.fileName)
		f.Close()
		return err
	}

	// for YAMLMapper, since we save the entire file each time, we have to jump
	// to the root task on every save to save everything - so we "recurse" to
	// the root and allow it to come back around here - unless we're already
	// at the root

	// since we're about to change this (soon we'll have a tag
	// identify the root) we'll just assume a single hierarchy
	var root *Task = t
	for root.HasParents() {
		root = root.FirstParent()
	}
	err = tm.saveTask(f, root);
	if err != nil {
	   	log.Printf("unable to write tasks to YAML file: %s\n", tm.fileName)
		f.Close()
		return err
	}

	f.Close()
	return nil
}

func (tm *TaskDataMapperYAML) addChildTask(parent* Task, yt* TaskYAML) (error, *Task) {
	
	child := &Task{id:yt.Id}
	child.SetName(yt.Name)
	child.Estimate = time.Duration(yt.Estimate) * time.Minute
	child.ActualStartTime = yt.ActualStartTime
	child.ActualCompletionTime = yt.ActualCompletionTime
	child.TargetStartTime = yt.TargetStartTime
	child.SetState(TaskStateFromString(yt.State))
	/*
	switch yt.State {
		case "notStarted": child.SetState(notStarted)
		case "complete": child.SetState(complete)
		case "inProgress": child.SetState(inProgress)
		case "onHold": child.SetState(onHold)
		default: child.SetState(notStarted)
	}
	*/

	for _,v := range yt.Tags {
		child.SetTag(v)
	}

	parent.AddChild(child)
	return nil, child
}


/*
=============================================================================
 Load()
-----------------------------------------------------------------------------
 Inputs: loadChildren bool - true to recurse to load children
         root         bool - true to ONLY load children (don't load this task)

 For YAML, only one combination of these parameters is supported: where both
 root and loadChildren are true.  This is because YAML Mapper assumes we
 wish to always load (and save) the entire task hierarchy into memory.
===========================================================================*/
func (tm *TaskDataMapperYAML) Load(t *Task, loadChildren bool, root bool) error {

	// log.Printf("Load(%s, %t, %t)", t.Name, loadChildren, root)

	// make sure we got a valid request
	if !root || !loadChildren {
		log.Printf("YAML data mapper only supports loading entire file into memory at once\n")
		return nil
	}

	// read the entire file into a buffer
	// readFile := "yaml/test.yaml"
	readFile := tm.fileName
	data, err := ioutil.ReadFile(readFile)
	if err != nil {
		log.Printf("Unable to open or read YAML file: %s\n", readFile)
  		return nil
	}

	// read the yaml data
	// log.Printf("YAML file: %s\n", yaml)
	var yamlTasks TasksYAML
	source := []byte(data) 
    err = yaml.Unmarshal(source, &yamlTasks)
    if err != nil {

    	// if we hit an error, we report it to the console and stop
    	// processing - but we may have partially loaded the file
    	// up until the error was hit.

	    // for now log the error and continue - not sure what more robust
	    // action we could take (maybe find a way to skip the bad line and
	   	// continue parsing - or set a flag so the UI can let users know
	   	// the file wasn't fully loaded?)
        log.Printf("YAML parsing error: %v", err)
    }
 
    // log.Printf("--- YAML Tasks:\n%+v\n\n", yamlTasks)

    // convert yaml tasks into real tasks
    for _, v := range yamlTasks.Tasks {

    	// if no parents, then add this task to the master
    	if len(v.Parents) == 0 {
    		tm.addChildTask(t, &v)
    	} else {
    		var parent *Task
    		var child *Task
    		child = nil
    		for _, parentId := range v.Parents {
    			parent = t.FindDescendent(parentId)

    			// if this parent was found
    			if parent != nil {

    				// if not already created, create it and add it
    				// but if already created then just link it to the additional parent
    				if child == nil {
    					_, child = tm.addChildTask(parent, &v)
					} else {
						parent.AddChild(child)
					}
    			} else {
    				log.Printf("TaskDataMapperYAML.Load(): Unable to find parent id <%s> for <%s>.\n", parentId, v.Id)
    			}
    		} // for each requested parent id

    		// parents were specified but none found, so don't orphan the task - link to the root
    		if child == nil {
    			tm.addChildTask(t, &v)
    			log.Printf("TaskDataMapperYAML.Load(): Unable to find any parents for <%s> so linked <%s>to root task.\n", v.Id, v.Name)
    		}

    	} // else there are parents to be sought
    }

	return nil
}

// Delete is not supported on the YAML mapper - you can only save or load the entire task list
func (tm *TaskDataMapperYAML) Delete(t *Task, reparent *Task) error {
	return nil
}


