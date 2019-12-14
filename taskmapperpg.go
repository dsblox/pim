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
    DB_HOST_ENV = "DAB_DB_HOST" // an environment variable i cause to be created
    // note the database name is always passed in from above
    // to allow easy switching between databases

    // the migration version is used with my homemade migration code
    // and maps to a 4-digit set of migration files for Origin, Up
    // and Down files to be run on clean DBs, to upgrade or rollback.
    DB_MIGRATION_VERSION = 3
)

// TaskDataMapperPostgreSQL implements TaskDataMapper to persist tasks
type TaskDataMapperPostgreSQL struct {
	loaded bool
	dbName string
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
    migrationVersion int
}
var env *Env = nil

// Function used when we discover a brand new postgres server that has no
// PIM database.  This isolates the creation of the new DB.  Note that
// empty tables are added outside this function for a more elegant use of
// defer db.Close(). 
func CreateEmptyPIMDatabase(dbHost string, dbName string) error {

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
    tmpenv := &Env{db: db, migrationVersion: DB_MIGRATION_VERSION}

    // dbCreate(tmpenv, dbName)  // string substitution is failing for a reason I don't understand!
	_, dberr = dbExec(tmpenv, "CREATE DATABASE " + dbName)
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
func NewTaskDataMapperPostgreSQL(saved bool, dbName string) *TaskDataMapperPostgreSQL {

	// if global env not yet initialized then initialize it
	if env == nil {

	    // find the DB host from env variable
	    dbHost := os.Getenv(DB_HOST_ENV)
	    if len(dbHost) == 0 {
		    fmt.Printf(" Aborting: DB host not found in expected environment variable: DB_PORT_5432_TCP_ADDR\n")
		    return nil // we should change this function to return an error
	    }

	    // if dbName not provided then use the default
	    if len(dbName) == 0 {
		    fmt.Printf(" Aborting: DB name not provided\n")
	    	return nil
	    }

	    // connect to the PIM database
	    dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD, dbName)
	    db, dberr := NewDB(dbinfo)

	    // if we get an error it may be because this if first time and we need to create DB
	    if dberr != nil {

	    	// create the empty PIM database and then retry connection
	    	dberr = CreateEmptyPIMDatabase(dbHost, dbName)
	    	if dberr != nil {
		    	return nil // we've got bigger problems - couldn't create empty DB
	    	}

			// connect to the newly created DB and set the env so we can run our commands
		    dbinfo = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD, dbName)
		    db, dberr = NewDB(dbinfo)
		    if dberr != nil {
		    	fmt.Printf(" could not open connection to new pim DB: %s\n", dberr)
		    	return nil // we should change this function to return an error
		    }
		    env = &Env{db: db, migrationVersion: DB_MIGRATION_VERSION}

			// create the tables we need empty
		    dberr = dbMigrateOrigin(env);

			if dberr != nil {
				fmt.Printf(" Initial table creation in empty database failed: %s\n", dberr)
				return nil
			}
   			fmt.Println(" successful.")
	    } else {
		    env = &Env{db: db, migrationVersion: DB_MIGRATION_VERSION}
		    dberr = dbMigrateUp(env) // TBD: TEST THIS CODE!!!
		    if dberr != nil {
				fmt.Printf(" Database migrations failed: %s\n", dberr)
				return nil
			}

	    }
	} // if we need to initialize the db

	// create and return the mapper object
	return &TaskDataMapperPostgreSQL{loaded:saved,dbName:dbName}
}

// NewDataMapper() implements for PostgreSQL the ability to return a new
// and unpersisted instance of the mapper that can later be filled in
// by the object with the list of saved parent ids
func (tm TaskDataMapperPostgreSQL) NewDataMapper(storageName string) TaskDataMapper {
	return NewTaskDataMapperPostgreSQL(false, storageName)
}
func (tm TaskDataMapperPostgreSQL) CopyDataMapper() TaskDataMapper {
	return NewTaskDataMapperPostgreSQL(false, tm.dbName)
}

// this function is used to update system tags - mapping a boolean that has been set
// on the task into the task_tags db table.  Now that tags are stored as tags in
// memory as well, we should change this stuff to just dump all the tags into the
// DB but we need a good way to RESET a tag that was removed.
func (tm TaskDataMapperPostgreSQL) syncSystemTag(tagBool bool, t *Task, tagName string, tagId int, newTask bool) error {
	if (tagBool) {
		_, err := dbExec(env, `INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2) 
			                   ON CONFLICT ON CONSTRAINT pk_tasktags DO NOTHING`, t.GetId(), tagId)
		if (err != nil)	{ 
			err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert %s tag over task %s: %s", tagName, t.GetName(), err))
			return err
		}
	} else if !newTask {
		_, err := dbExec(env, `DELETE FROM task_tags WHERE task_id = $1 AND tag_id = $2`, t.GetId(), tagId)
		if (err != nil)	{ 
			// eat this error - we always try to delete if system tag is not set on the object
			// this is dangerous - if other DB errors cause this we will eat the error!
		}		
	}
	return nil;
}

func (tm TaskDataMapperPostgreSQL) syncSystemTags(t *Task, newTask bool) error {
	err := tm.syncSystemTag(t.IsTagSet("today"), t, "today", 1, newTask)
	if (err != nil) {
		return err;
	}
	err = tm.syncSystemTag(t.IsTagSet("thisweek"), t, "thisweek", 2, newTask)
	if (err != nil) {
		return err;
	}
	err = tm.syncSystemTag(t.IsDontForget(), t, "dontforget", 3, newTask)
	if (err != nil) {
		return err;
	}
	return nil;
}


func (tm *TaskDataMapperPostgreSQL) Save(t *Task, saveChildren bool, saveMyself bool) error {

	// log.Printf("Save(%t, %t): task = %s, id = %s, loaded = %t len(parentIds) = %d", saveChildren, saveMyself, t.name, t.id, tm.loaded, len(tm.parentIds))

	// only save myself if requested
	if saveMyself {

		// upsert the task itself
		if tm.loaded {
			_, err := dbExec(env, `UPDATE tasks SET name = $1, state = $2, target_start_time = $3, actual_start_time = $4, actual_completion_time = $5, estimate_minutes = $6 
				                   WHERE ID = $7`, t.GetName(), t.GetState(), t.TargetStartTime, t.ActualStartTime, t.ActualCompletionTime, int(t.Estimate.Minutes()), t.GetId())
			if (err != nil) {
				err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to update task %s: %s", t.GetName(), err))
				return err
			}

			// update today, thisweek and dontforget tags - false means its an update
			tm.syncSystemTags(t, false)
			if (err != nil) {
				return err;
			}

		} else {
	    	_, err := dbExec(env, `INSERT INTO tasks (id, name, state, target_start_time, actual_start_time, actual_completion_time, estimate_minutes) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, 
	    		             t.GetId(), t.GetName(), t.GetState(), t.TargetStartTime, t.ActualStartTime, t.ActualCompletionTime, int(t.Estimate.Minutes()))
			if (err != nil)	{ 
				err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert task %s: %s", t.GetName(), err))
				return err
			}

			// update today, thisweek and dontforget tags - true means its a new task
			tm.syncSystemTags(t, true)
			if (err != nil) {
				return err;
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

func dbTimeCheck(db_time pq.NullTime) *time.Time {
	db_time_value, _ := db_time.Value()
	if db_time_value != nil {
		my_time := db_time_value.(time.Time)
		return &my_time
	} else {
		return nil
	}
}

func (tm TaskDataMapperPostgreSQL) setTaskFields(t *Task, 
												 db_target_start_time      pq.NullTime,
												 db_actual_start_time      pq.NullTime,
												 db_actual_completion_time pq.NullTime,
												 db_estimate_minutes       sql.NullInt64,
												 db_today				   sql.NullBool,
												 db_thisweek               sql.NullBool,
												 db_dontforget             sql.NullBool) error {
	// go is terrible dealing with null values from databases
	// must use this intervening struct to track if null or not
	t.SetTargetStartTime(dbTimeCheck(db_target_start_time))
	t.SetActualStartTime(dbTimeCheck(db_actual_start_time))
	t.SetActualCompletionTime(dbTimeCheck(db_actual_completion_time))

    /* -- this was the old way - keeping in case I have to debug dbTimeCheck()
	db_value_start_time, _ := db_start_time.Value()
	if (db_value_start_time != nil) {
		start_time := db_value_start_time.(time.Time)
		t.SetStartTime(&start_time)
	}
	*/

	db_value_estimate_minutes, _ := db_estimate_minutes.Value()
	if (db_value_estimate_minutes != nil) {
		estimate_minutes := db_value_estimate_minutes.(int64)
		t.SetEstimate(time.Duration(estimate_minutes) * time.Minute)
	}

	db_value_today, _ := db_today.Value()
	if (db_value_today != nil) {
		today := db_value_today.(bool)
		t.SetToday(today)
	}

	db_value_thisweek, _ := db_thisweek.Value()
	if (db_value_thisweek != nil) {
		thisweek := db_value_thisweek.(bool)
		t.SetThisWeek(thisweek)
	}

	db_value_dontforget, _ := db_dontforget.Value()
	if (db_value_dontforget != nil) {
		dontforget := db_value_dontforget.(bool)
		t.SetDontForget(dontforget)
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
		db_target_start_time pq.NullTime
		db_actual_start_time pq.NullTime
		db_actual_completion_time pq.NullTime
		db_estimate_minutes sql.NullInt64
		db_today sql.NullBool
		db_thisweek sql.NullBool
		db_dontforget sql.NullBool
	)

	// if we're told this is "root" that means we should not attempt
	// to load into "this" task, just move on to the kids
	if (!root) {

		// build and execute the query for the task
		taskQuery := `SELECT t.name, t.state, t.target_start_time, t.actual_start_time, t.actual_completion_time, t.estimate_minutes, 
		              COUNT(tt.tag_id = 1) > 0 AS today, 
		              COUNT(tt.tag_id = 2) > 0 AS thisweek,
		              COUNT(tt.tag_id = 3) > 0 AS dontforget
		              FROM tasks t
		              LEFT JOIN task_tags tt on tt.task_id = t.id
		              WHERE id = '" + t.GetId() + "'" + "
		              GROUP BY t.id`
		err := env.db.QueryRow(taskQuery).Scan(&name, &state, &db_target_start_time, &db_actual_start_time, &db_actual_completion_time, &db_estimate_minutes, &db_today, &db_thisweek, &db_dontforget)
		if err != nil {
			// log.Printf("query for a task failed: %s, err: %s\n", taskQuery, err)
			return err
		}

		// overwrite my in-memory values
		t.SetName(name)
		t.SetState(state)
		tm.setTaskFields(t, db_target_start_time, db_actual_start_time, db_actual_completion_time, db_estimate_minutes, db_today, db_thisweek, db_dontforget)

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
		dbid string
		dbname string
		dbstate TaskState
		db_target_start_time pq.NullTime
		db_actual_start_time pq.NullTime
		db_actual_completion_time pq.NullTime
		db_estimate_minutes sql.NullInt64
		db_today sql.NullBool
		db_thisweek sql.NullBool
		db_dontforget sql.NullBool
	)

	var baseQuery string = `SELECT t.id, t.name, t.state, t.target_start_time, t.actual_start_time, t.actual_completion_time, t.estimate_minutes,
				            COUNT(tt.tag_id = 1) > 0 AS today, 
				            COUNT(tt.tag_id = 2) > 0 AS thisweek,
				            COUNT(tt.tag_id = 3) > 0 AS dontforget
	                        FROM tasks t
		                    LEFT JOIN task_parents tp ON tp.child_id = t.id
		                    LEFT JOIN task_tags tt ON tt.task_id = t.id
		                    WHERE tp.parent_id %s
		                    GROUP BY t.id`

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
		err := rows.Scan(&dbid, &dbname, &dbstate, &db_target_start_time, &db_actual_start_time, &db_actual_completion_time, &db_estimate_minutes, &db_today, &db_thisweek, &db_dontforget)
		if err != nil {
			log.Printf("tmpg.loadChildren(): row scan failed\n")
			log.Fatal(err)
		}
		// log.Printf("LoadChildren(): read id=%s, name=%s\n", id, name)

		// create the child task
		k := &Task{id:dbid, name:dbname, state:dbstate}
		tm.setTaskFields(k, db_target_start_time, db_actual_start_time, db_actual_completion_time, db_estimate_minutes, db_today, db_thisweek, db_dontforget)

		// set the data mapper onto the child indicating that it was loaded from DB
		// TBD - this should copy the mapper from tm - shouldn't it??? I think DB_NAME will be
		// wrong if we don't use the default database name
		kdm := NewTaskDataMapperPostgreSQL(true, DB_NAME)
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

	// delete all tag references from this tasks_tags table
	_, err = dbExec(env, "DELETE FROM task_tags WHERE task_id = $1", t.GetId())
	if err != nil { // we should make sure this doesn't return an error if no tags are on the task - if it does we should eat that error
		err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove tags from task %s with id %d: %s", t.GetName(), t.GetId(), err))
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


