package main

import "fmt"
import "errors"
import "log"
import "database/sql"


type Mapper interface {
	Load() string
	Save() string
}

type MapperSQL struct {
	TimesUsed int
}

func (ms MapperSQL) Load() string {
	ms.TimesUsed += 1
	return "Load"
}
func (ms *MapperSQL) Save() string {
	ms.TimesUsed += 1
	return "Save"
}

type TestTask struct {
	m Mapper
}

func (t* TestTask) SetMapper(newM Mapper) {
	t.m = newM
}

func tryMe() {
	m := &MapperSQL{TimesUsed:0}
	t := TestTask{}
	t.SetMapper(m)
	fmt.Println(m.Load())
	fmt.Println(m.Save())
	fmt.Println(m.TimesUsed)
}


type TaskDataMapperPostgreSQL struct {
	id int
	parentIds []int
}

func (tm *TaskDataMapperPostgreSQL) Id() int {
	return tm.id
}
func (tm *TaskDataMapperPostgreSQL) SetId(newId int) {
	tm.id = newId
}
func (tm *TaskDataMapperPostgreSQL) FindParentId(id int) int {
	return findInt(tm.parentIds, id)
}
func (tm *TaskDataMapperPostgreSQL) HasParentId(id int) bool {
	return tm.FindParentId(id) != -1
}
func (tm *TaskDataMapperPostgreSQL) AddParentId(newId int) {
	if (!tm.HasParentId(newId)) {
		tm.parentIds = append(tm.parentIds, newId)
	}
}
func (tm *TaskDataMapperPostgreSQL) RemoveParentId(id int) {
	tm.parentIds = removeIntByValue(tm.parentIds, id)
}



// global variable to hold a DB connection
// can we find a way to make it a static
// in the TaaskDataMapperPostgreSQL world?
type Env struct {
    db *sql.DB
}
var env *Env = nil

// initialize the database and hold in global variable env
// TBD - isolate this in a DB layer better and perhaps stop
// using global variables so we can later support multiple
// DB connections.
func NewTaskDataMapperPostgreSQL(id int) *TaskDataMapperPostgreSQL {

	// if global env not yet initialized then initialize it
	if env == nil {
	    db, err := NewDB("postgres://postgres:postgres@localhost/pim?sslmode=disable")
	    if err != nil {
	        log.Panic(err)
	    }
	    env = &Env{db: db}		
	}

	// create and return the mapper object
	return &TaskDataMapperPostgreSQL{id:id}
}

// NewDataMapper() implements for PostgreSQL the ability to return a new
// and unpersisted instance of the mapper that can later be filled in
// by the object with an id once it is saved
func (tm TaskDataMapperPostgreSQL) NewDataMapper() TaskDataMapper {
	return NewTaskDataMapperPostgreSQL(-1)
}

func (tm *TaskDataMapperPostgreSQL) Save(t *Task) error {

	// log.Printf("Save(): task = %s, id = %d, len(parentIds) = %d", t.name, tm.id, len(tm.parentIds))

	// upsert the task itself
	if tm.id >= 0 {
		_, err := dbExec(env, "UPDATE tasks SET name = $1, state = $2 WHERE ID = $3", t.name, t.state, tm.id)
		if (err != nil) {
			err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to update task %s: %s", t.name, err))
			return err
		}
	} else {

 		var newId int
    	err := env.db.QueryRow(`INSERT INTO tasks (name, state) VALUES ($1, $2) RETURNING id`, t.name, t.state).Scan(&newId)
		/// newId, err := dbInsert(env, "INSERT INTO tasks (name, state) VALUES ($1, $2) RETURNING id", t.name, t.state)
		if (err != nil)	{ 
			err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert task %s: %s", t.name, err))
			return err
		}
		tm.id = newId
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

		// for this parent get it's id - we have to cast any datamapper found
		var tmParent interface{}
		tmParent = p.DataMapper()
		tmParentPG := tmParent.(*TaskDataMapperPostgreSQL)
		id := tmParentPG.Id()

		// if we got an id then write the parent / child relationship

		// first see if it is already written - if so nothing to do
		// except remove from the previously saved ids
		if tm.HasParentId(id) {
			savedParentIds = removeIntByValue(savedParentIds, id)
		} else if id != -1 {  // -1 means it has not been saved so skip!
			// we need to create the new parent relationsip to ourselves
			_, err := dbInsert(env, "INSERT INTO task_parents (parent_id, child_id) VALUES ($1, $2)", id, tm.id)
			if err != nil	{ 
				err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert parent relationship between parent task %s and child task %s: %s", p.Name(), t.Name(), err))
				return err
			}
			tm.AddParentId(id)
		}
	}

	// once we've looped through all the parents, anything left needs to
	// be removed - it means the parentage that was once saved is no longer there
	for _, idParent := range savedParentIds {
		_, err := dbExec(env, "DELETE FROM task_parents WHERE parent_id = $1 AND child_id = $2", idParent, tm.id)
		if err != nil {
			err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to remove obsolete parent relationship with parent id %d to child task %s: %s", idParent, t.Name(), err))
			return err
		}
	}

	return nil
}

func (tm TaskDataMapperPostgreSQL) LoadChildren(parent *Task) error {
	// log.Printf("LoadChildren(): for parent task %s\n", parent.name)
	var (
		id int
		name string
		state TaskState
	)

	var baseQuery string = `SELECT t.id, t.name, t.state FROM tasks t
		                       LEFT JOIN task_parents tp ON tp.child_id = t.id
		                       WHERE tp.parent_id %s`

    // if no parent on this guy then we're at the top - load root tasks
    var sqlSelect string
    if tm.id == -1 {
    	sqlSelect = fmt.Sprintf(baseQuery, "IS NULL")
	} else {
		sqlSelect = fmt.Sprintf(baseQuery, fmt.Sprintf("= %d", tm.id))
	}

	// make the query for the kids
	rows, err := env.db.Query(sqlSelect)
	if err != nil {
		log.Fatal(err) // maybe we should not just die here
	}
	defer rows.Close()

	// for each child task in the DB
	for rows.Next() {
		err := rows.Scan(&id, &name, &state)
		if err != nil {
			log.Fatal(err)
		}

		// create the child task
		k := &Task{name:name, state:state}

		// set the data mapper onto the child explicitly with correct id
		k.SetDataMapper(NewTaskDataMapperPostgreSQL(id))

		// add the child to the supplied parent
		parent.AddChild(k) 

		// do some housekeeping to remember that this child came from
		// a relationship already in the DB so we don't try to recreate
		// it again later
		var tmKid interface{}
		tmKid = k.DataMapper()
		tmKidPG := tmKid.(*TaskDataMapperPostgreSQL)
		tmKidPG.AddParentId(tm.id)
		k.SetDataMapper(tmKidPG)

		// now that the child is fully loaded, recurse to go get it's children
		// assert that we've set the mapper id because if we haven't we'll
		// jump back to the top and recurse forever (we should throw an error?)
		if tmKidPG.Id() != -1 {
			err = tmKidPG.LoadChildren(k)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}	
	return nil

}

func (tm TaskDataMapperPostgreSQL) Load(parent *Task) error {
	return tm.LoadChildren(parent)
}

func (tm *TaskDataMapperPostgreSQL) Delete(t *Task, reparent *Task) error {

	// make sure we have an id for this task - if not it was never saved
	// and we have nothing to do here
	if tm.id == -1 {
		return nil
	}

	// if a reparenting is requested - update all tasks to have the new parent
	reparentId := -1
	if reparent != nil {
		var tmNewParent interface{}
		tmNewParent = reparent.DataMapper()
		tmNewParentPG := tmNewParent.(*TaskDataMapperPostgreSQL)
		reparentId = tmNewParentPG.Id()
	}
	if reparentId != -1 {
		// if reparenting is requested and that parent is in the DB already
		// then reparent this task to the requested new parent
		_, err := dbExec(env, "UPDATE task_parents SET parent_id = $1 WHERE parent_id = $2", reparentId, tm.id)
		if err != nil {
			err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to set new parent on children of task %s from id %d to id %d: %s", t.Name(), tm.id, reparentId, err))
			return err
		}
	} else {
		// delete all references to this task from task_parents table
		_, err := dbExec(env, "DELETE FROM task_parents WHERE parent_id = $1", tm.id)
		if err != nil {
			err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to delete parent references to task %s with id %d: %s", t.Name(), tm.id, err))
			return err
		}

	}

	// remove myself as a child from any parent tasks - no re-childing necessary
	_, err := dbExec(env, "DELETE FROM task_parents WHERE child_id = $1", tm.id)
	if err != nil {
		err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to delete child references to task %s with id %d: %s", t.Name(), tm.id, err))
		return err
	}

	// delete this task from the tasks table - must do this after deleting from
	// parent table
	_, err = dbExec(env, "DELETE FROM tasks WHERE id = $1", tm.id)
	if err != nil {
		err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove task %s with id %d: %s", t.Name(), tm.id, err))
		return err
	}


	// clean in-memory tm structures
	tm.id = -1
	tm.parentIds = nil

	return nil
}


