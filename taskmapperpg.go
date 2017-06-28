package main

import "fmt"
import "errors"
import "log"
import "os"
import "database/sql"
import "time"
import "github.com/lib/pq"

const (
    DB_USER     = "postgres"
    DB_PASSWORD = "postgres"
    DB_NAME     = "pim"
    DB_HOST_ENV = "DAB_DB_HOST" // an environment variable i cause to be created
)

// TaskDataMapperPostgreSQL implements TaskDataMapper to persist tasks
type TaskDataMapperPostgreSQL struct {
	loaded bool
	id int // can we get rid of both of these id and parentIds?
	parentIds []string
}

// TBD: 8/16/16...
// currently all these unique ID functions are in the task mapper
// but the ID should not be unique to the persistence layer, rather
// it should be pushed up to the rask object.  I believe all we'll
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

// TaskDataMapperPostgreSQL accessor and setter functions
func (tm *TaskDataMapperPostgreSQL) IsInDB() bool {
	return tm.loaded
}
func (tm *TaskDataMapperPostgreSQL) MarkInDB() {
	tm.loaded = true
}
/*
func (tm *TaskDataMapperPostgreSQL) Id() int {
	return tm.id
}
func (tm *TaskDataMapperPostgreSQL) SetId(newId int) {
	tm.id = newId
}
*/
func (tm *TaskDataMapperPostgreSQL) FindParentId(id string) int {
	return findString(tm.parentIds, id)
}
func (tm *TaskDataMapperPostgreSQL) HasParentId(id string) bool {
	return tm.FindParentId(id) != -1
}
func (tm *TaskDataMapperPostgreSQL) AddParentId(newId string) {
	if (!tm.HasParentId(newId)) {
		tm.parentIds = append(tm.parentIds, newId)
	}
}
func (tm *TaskDataMapperPostgreSQL) RemoveParentId(id string) {
	tm.parentIds = removeStringByValue(tm.parentIds, id)
}



// global variable to hold a DB connection
// can we find a way to make it a static
// in the TaaskDataMapperPostgreSQL world?
type Env struct {
    db *sql.DB
}
var env *Env = nil

// Function used when we discover a brand new postgres server that has no
// PIM database.  This isolates the creation of the new DB.  Note that
// empty tables are added outside this function for a more elegant use of
// defer db.Close(). 
func CreateEmptyPIMDatabase(dbHost string) error {

	// output to the console (log?) what's going on
	fmt.Println("PIM Database not found - creating empty database...")

	// open db without the database - this links to postgres' default DB
	// NOTE: there is no way in Postgres to connect without linking to _some_ DB
    dbinfo := fmt.Sprintf("host=%s user=%s password=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD)
    db, dberr := NewDB(dbinfo)
    if dberr != nil {
    	fmt.Printf(" could not open connection to PostgreSQL: %s\n", dberr)
    	return dberr
    }
    defer db.Close()

    // set the an env now so we can use our dbExec function
    tmpenv := &Env{db: db}

    // dbCreate(tmpenv, DB_NAME)  // string substitution is failing for a reason I don't understand!
	_, dberr = dbExec(tmpenv, "CREATE DATABASE pim")
	if dberr != nil {
		fmt.Printf(" CREATE DATABASE failed with error: %s\n", dberr)
		return dberr
	}

	// all set with no errors
	return nil
}


// initialize the database and hold in global variable env
// TBD - isolate this in a DB layer better and perhaps stop
// using global variables so we can later support multiple
// DB connections.
func NewTaskDataMapperPostgreSQL(saved bool) *TaskDataMapperPostgreSQL {

	// if global env not yet initialized then initialize it
	if env == nil {

	    // find the DB host from env variable
	    dbHost := os.Getenv(DB_HOST_ENV)
	    if len(dbHost) == 0 {
		    fmt.Printf(" Aborting: DB host not found in expected environment variable: DB_PORT_5432_TCP_ADDR\n")
		    return nil // we should change this function to return an error
	    }

	    // connect to the PIM database
	    dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD, DB_NAME)
	    db, dberr := NewDB(dbinfo)

	    // if we get an error it may be because this if first time and we need to create DB
	    if dberr != nil {

	    	// create the empty PIM database and then retry connection
	    	dberr = CreateEmptyPIMDatabase(dbHost)
	    	if dberr != nil {
		    	return nil // we've got bigger problems - couldn't create empty DB
	    	}

			// connect to the newly created DB and set the env so we can run our commands
		    dbinfo = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD, DB_NAME)
		    db, dberr = NewDB(dbinfo)
		    if dberr != nil {
		    	fmt.Printf(" could not open connection to new pim DB: %s\n", dberr)
		    	return nil // we should change this function to return an error
		    }
		    env = &Env{db: db}

			// create the tables we need empty
   			_, dberr = dbExec(env, `CREATE TABLE tasks ( 
	        	         id CHAR(36) PRIMARY KEY,
	            	     name VARCHAR(1024) NOT NULL,
	                	 state INT NOT NULL,
	                	 start_time TIMESTAMP,
	                	 estimate_minutes INT,
		                 created_at TIMESTAMP,
		                 modified_at TIMESTAMP)`)
			if dberr != nil {
				fmt.Printf(" CREATE TABLE TASKS failed: %s\n", dberr)
				return nil
			}
   			_, dberr = dbExec(env, `CREATE TABLE task_parents (
						parent_id CHAR(36) NOT NULL,
						child_id CHAR(36) NOT NULL,
						created_at TIMESTAMP,
						modified_at TIMESTAMP,
						CONSTRAINT pk_parents PRIMARY KEY (parent_id,child_id),
						FOREIGN KEY (parent_id) REFERENCES tasks(id),
						FOREIGN KEY (child_id) REFERENCES tasks(id) )`)
			if dberr != nil {
				fmt.Printf(" CREATE TABLE TASKS failed: %s\n", dberr)
				return nil
			}
   			fmt.Println(" successful.")
	    } else {
		    env = &Env{db: db}
	    }
	} // if we need to initialize the db

	// create and return the mapper object
	return &TaskDataMapperPostgreSQL{loaded:saved}
}

// NewDataMapper() implements for PostgreSQL the ability to return a new
// and unpersisted instance of the mapper that can later be filled in
// by the object with the list of saved parent ids
func (tm TaskDataMapperPostgreSQL) NewDataMapper() TaskDataMapper {
	return NewTaskDataMapperPostgreSQL(false)
}

func (tm *TaskDataMapperPostgreSQL) Save(t *Task, saveChildren bool, saveMyself bool) error {

	// log.Printf("Save(%t, %t): task = %s, id = %s, loaded = %t len(parentIds) = %d", saveChildren, saveMyself, t.name, t.id, tm.loaded, len(tm.parentIds))

	// only save myself if requested
	if saveMyself {

		// upsert the task itself
		if tm.loaded {
			_, err := dbExec(env, `UPDATE tasks SET name = $1, state = $2, start_time = $3, estimate_minutes = $4  
				                   WHERE ID = $5`, t.Name, t.State, t.StartTime, int(t.Estimate.Minutes()), t.Id)
			if (err != nil) {
				err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to update task %s: %s", t.Name, err))
				return err
			}
		} else {
	    	_, err := dbExec(env, `INSERT INTO tasks (id, name, state, start_time, estimate_minutes) VALUES ($1, $2, $3, $4, $5) RETURNING id`, 
	    		             t.Id, t.Name, t.State, t.StartTime, int(t.Estimate.Minutes()))
			if (err != nil)	{ 
				err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert task %s: %s", t.Name, err))
				return err
			}
			tm.MarkInDB()
		}

		// now set the parent relationships - do we break datamapper pattern
		// by not also saving our children? note that we have our own
		// id if we get here - we just need the ids of our parents which is
		// super ugly - they may not have been saved - we must assume they have been
		// and that saving happens top-down.

		// first collect the parent ids we already have recorded as saved
		// we'll remove them as we go and anything no longer in the parent
		// list we'll remove at the end
		savedParentIds := tm.parentIds

		for p := t.FirstParent(); p != nil; p = t.NextParent() {

			// for this parent get it's id
			id := p.GetId()

			// for this parent make sure it is in the database
			// and if it is not we will skip it!  we assume saves
			// are done top-down, so if it isn't "loaded" then
			// the parent must be a memory-only task and the
			// task we are saving should be set to be a root task
			var tmParent interface{}
			tmParent = p.DataMapper()
			tmParentPG := tmParent.(*TaskDataMapperPostgreSQL)
			skipParent := !tmParentPG.loaded

			// if we got an id then write the parent / child relationship
			if !skipParent {

				// first see if it is already written - if so nothing to do
				// except remove from the previously saved ids
				if tm.HasParentId(id) {
					savedParentIds = removeStringByValue(savedParentIds, id)
				} else {
					// if it wasn't so much work to check, i'd say we should assert that
					// saves are happening top-down and that the parent is already in the
					// DB before we call insert - because the insert will fail for data
					// integrity reasons if the parent is not already in the DB
					// assert tmParent.IsInDB()
					_, err := dbInsert(env, "INSERT INTO task_parents (parent_id, child_id) VALUES ($1, $2)", id, t.GetId())
					if err != nil	{ 
						err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert parent relationship between parent task %s and child task %s: %s", p.GetName(), t.GetName(), err))
						return err
					}
					tm.AddParentId(id)
				}
			} // skip this parent if not loaded
		}

		// once we've looped through all the parents, anything left needs to
		// be removed - it means the parentage that was once saved is no longer there
		for _, idParent := range savedParentIds {
			_, err := dbExec(env, "DELETE FROM task_parents WHERE parent_id = $1 AND child_id = $2", idParent, t.GetId())
			if err != nil {
				err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to remove obsolete parent relationship with parent id %d to child task %s: %s", idParent, t.GetName(), err))
				return err
			}
		}
	} // if saveMyself

	// now recurse if requested
	if saveChildren {
		var err error = nil
		for c := t.FirstChild(); c != nil && err == nil; c = t.NextChild() {
			// log.Printf("about to save id=%s, name=%s\n", c.Id(), c.Name())
			// err = c.Save(true, true) - could use this to enforce memory-only objects within hierarchy
			err = c.persist.Save(c, true, true)
		}
		if err != nil {
			return err
		}
	} // if saveChildren

	return nil
}

func (tm TaskDataMapperPostgreSQL) setTaskFields(t *Task, 
												 db_start_time pq.NullTime,
												 db_estimate_minutes sql.NullInt64) error {
	// go is terrible dealing with null values from databases
	// must use this intervening struct to track if null or not
	db_value_start_time, _ := db_start_time.Value()
	if (db_value_start_time != nil) {
		start_time := db_value_start_time.(time.Time)
		t.SetStartTime(&start_time)
	}

	db_value_estimate_minutes, _ := db_estimate_minutes.Value()
	if (db_value_estimate_minutes != nil) {
		estimate_minutes := db_value_estimate_minutes.(int64)
		t.SetEstimate(time.Duration(estimate_minutes) * time.Minute)
	}

	return nil	
}

/*
=============================================================================
 Load()
-----------------------------------------------------------------------------
 Inputs: loadChildren bool - true to recurse to laod children
         root         bool - true to ONLY load children (don't load this task)

 Load a task given its id, and optionally recurse to load all its children.

 OK, this is confusing, so listen up.  In the normal case of reloading our
 task database we actually don't want to load an individual task by id,
 because we don't know the ids of all the tasks.  So you CAN use this to
 load an individual task if you happen to know it's ID, but the usual cases
 are to 
  - load one task: call this function without loading children if you're 
    trying to load just one task by known id, OR
  - load all tasks: call it with root set to true which will cause it to
    skip loading the task itself, and just load all "root" tasks into the 
    provided (presumably) in-memory task.

 Why do I say it is confusing?  Because a normal recursion would Load()
 each individual task.  We don't want to do so because it is so much
 more efficient to load all "peers" at the same time.  So only the
 top-level parent is loaded individually.  All kids are loaded based
 on their parent lists within loadChildren().

 See loadChildren() which does the heavy-lifting of recursion, but is not
 exposed.  Control the behavior you want using the parameters on Load():
   - provide a memory-only task to skip loading the top-level task
   - set loadChildren to true with the memory-only task to load entire DB
   - set loadChildren to false with a non-memory-only task to load just
     that task.
   - if you actually know the root of the subset of the hierarchy you want
     to load, then set loadChildren to false and provide a non-memory-only
     task.
===========================================================================*/
func (tm TaskDataMapperPostgreSQL) Load(t *Task, loadChildren bool, root bool) error {

	var (
		name string
		state TaskState
		db_start_time pq.NullTime
		db_estimate_minutes sql.NullInt64
	)

	// if we're told this is "root" that means we should not attempt
	// to load into "this" task, just move on to the kids
	if (!root) {

		// build and execute the query for the task
		taskQuery := "SELECT name, state, start_time, estimate_minutes FROM tasks WHERE id = '" + t.GetId() + "'"
		err := env.db.QueryRow(taskQuery).Scan(&name, &state, &db_start_time, &db_estimate_minutes)
		if err != nil {
			// log.Printf("query for a task failed: %s, err: %s\n", taskQuery, err)
			return err
		}

		// overwrite my in-memory values
		t.SetName(name)
		t.SetState(state)
		tm.setTaskFields(t, db_start_time, db_estimate_minutes)

		// go is terrible dealing with null values from databases
		// must use this intervening struct to track if null or not
		/*
		db_value_start_time, _ := db_start_time.Value()
		if (db_value_start_time != nil) {
			start_time := db_value_start_time.(time.Time)
			t.SetStartTime(start_time)
		}

		db_value_estimate_minutes, _ := db_estimate_minutes.Value()
		if (db_value_estimate_minutes != nil) {
			estimate_minutes := db_value_estimate_minutes.(int64)
			t.SetEstimate(time.Duration(estimate_minutes) * time.Minute)
		} */

		// set myself as loaded from the DB
		tm.loaded = true
	}

	// if we're being ask to load the hierarchy below this task then do it
	// passing along "root=true" will tell loadChildren to attach all tasks in
	// the DB whose parents are NULL (top-level tasks) to me
	if loadChildren {
		return tm.loadChildren(t, root)
	} else {
		return nil
	}
}


func (tm TaskDataMapperPostgreSQL) loadChildren(parent *Task, root bool) error {
	// log.Printf("LoadChildren(): for parent task %s\n", parent.name)
	var (
		id string
		name string
		state TaskState
		db_start_time pq.NullTime
		db_estimate_minutes sql.NullInt64
	)

	var baseQuery string = `SELECT t.id, t.name, t.state, t.start_time, t.estimate_minutes FROM tasks t
		                       LEFT JOIN task_parents tp ON tp.child_id = t.id
		                       WHERE tp.parent_id %s`

    // if no parent on this guy then we're at the top - load root tasks
    var sqlSelect string
    if root {
    	sqlSelect = fmt.Sprintf(baseQuery, "IS NULL")
	} else {
		sqlSelect = fmt.Sprintf(baseQuery, fmt.Sprintf("= '%s'", parent.GetId()))
	}

	// make the query for the kids
	rows, err := env.db.Query(sqlSelect)
	if err != nil {
		log.Printf("query for the kids failed: %s\n", sqlSelect)
		log.Fatal(err) // maybe we should not just die here
	}
	defer rows.Close()

	// for each child task in the DB
	for rows.Next() {
		err := rows.Scan(&id, &name, &state, &db_start_time, &db_estimate_minutes)
		if err != nil {
			log.Printf("row scan failed\n")
			log.Fatal(err)
		}
		// log.Printf("LoadChildren(): read id=%s, name=%s\n", id, name)

		// create the child task
		k := &Task{Id:id, Name:name, State:state}
		tm.setTaskFields(k, db_start_time, db_estimate_minutes)

		// set the data mapper onto the child indicating that it was loaded from DB
		kdm := NewTaskDataMapperPostgreSQL(true)
		k.SetDataMapper(kdm)

		// add the child to the supplied parent
		parent.AddChild(k) 

		// do some housekeeping to remember that this child came from
		// a relationship already in the DB so we don't try to recreate
		// it again later
		kdm.AddParentId(parent.GetId())

		// now that the child is fully loaded, recurse to go get it's children
		err = kdm.loadChildren(k, false)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}	
	return nil

}

func (tm *TaskDataMapperPostgreSQL) Delete(t *Task, reparent *Task) error {

	// if the task has never been saved then no work to here
	if !tm.IsInDB() {
		return nil
	}

	// if a reparenting is requested - update all tasks to have the new parent
	bReparent := false
	if reparent != nil {
		// need to get the reparent's mapper to know if it is in the DB
		var tmNewParent interface{}
		tmNewParent = reparent.DataMapper()
		tmNewParentPG := tmNewParent.(*TaskDataMapperPostgreSQL)
		bReparent = tmNewParentPG.IsInDB()
	}
	if bReparent {
		// if reparenting is requested and that parent is in the DB already
		// then reparent this task to the requested new parent
		_, err := dbExec(env, "UPDATE task_parents SET parent_id = $1 WHERE parent_id = $2", reparent.GetId(), t.GetId())
		if err != nil {
			err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to set new parent on children of task %s from id %d to id %d: %s", t.GetName(), t.GetId(), reparent.GetId(), err))
			return err
		}
	} else {
		// delete all references to this task from task_parents table
		_, err := dbExec(env, "DELETE FROM task_parents WHERE parent_id = $1", t.GetId())
		if err != nil {
			err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to delete parent references to task %s with id %d: %s", t.GetName(), t.GetId(), err))
			return err
		}

	}

	// remove myself as a child from any parent tasks - no re-childing necessary
	_, err := dbExec(env, "DELETE FROM task_parents WHERE child_id = $1", t.GetId())
	if err != nil {
		err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to delete child references to task %s with id %d: %s", t.GetName(), t.GetId(), err))
		return err
	}

	// delete this task from the tasks table - must do this after deleting from
	// parent table
	_, err = dbExec(env, "DELETE FROM tasks WHERE id = $1", t.GetId())
	if err != nil {
		err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove task %s with id %d: %s", t.GetName(), t.GetId(), err))
		return err
	}

	// clean in-memory tm structures
	tm.loaded = false
	tm.parentIds = nil

	return nil
}


