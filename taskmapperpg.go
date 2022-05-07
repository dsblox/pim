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
    DB_MIGRATION_VERSION = 5
)

type PimPersistPostgreSQL struct {
  dbName string
  err error
}


// TaskDataMapperPostgreSQL implements TaskDataMapper to persist tasks
type TaskDataMapperPostgreSQL struct {
  loaded bool
  dbName string
  id int // can we get rid of both of these id and parentIds?
  parentIds []string
  err error
}

func (tm *TaskDataMapperPostgreSQL) Error() error {
  return tm.err
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
func NewPimPersistPostgreSQL(dbName string) (*PimPersistPostgreSQL, error) {

  // if global env not yet initialized then initialize it
  if env == nil {

      // find the DB host from env variable
      dbHost := os.Getenv(DB_HOST_ENV)
      if len(dbHost) == 0 {
        fmt.Printf(" Aborting: DB host not found in expected environment variable: DB_PORT_5432_TCP_ADDR\n")
        return nil, errors.New("Aborting: DB host not found in expected environment variable: DB_PORT_5432_TCP_ADDR")
      }

      // if dbName not provided then use the default
      if len(dbName) == 0 {
        fmt.Printf(" Aborting: DB name not provided\n")
        return nil, errors.New("Aborting: DB name not provided")
      }

      // connect to the PIM database
      dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD, dbName)
      db, dberr := NewDB(dbinfo)

      // if we get an error it may be because this if first time and we need to create DB
      if dberr != nil {

        // create the empty PIM database and then retry connection
        dberr = CreateEmptyPIMDatabase(dbHost, dbName)
        if dberr != nil {
          return nil, errors.New("Couldn't even create empty DB")
          // we've got bigger problems - couldn't create empty DB
        }

      // connect to the newly created DB and set the env so we can run our commands
        dbinfo = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, DB_USER, DB_PASSWORD, dbName)
        db, dberr = NewDB(dbinfo)
        if dberr != nil {
          fmt.Printf(" could not open connection to new pim DB: %s\n", dberr)
          return nil, errors.New("could not open connection to new pim DB")
        }
        env = &Env{db: db, migrationVersion: DB_MIGRATION_VERSION}

      // create the tables we need empty
        dberr = dbMigrateOrigin(env);

      if dberr != nil {
        fmt.Printf(" Initial table creation in empty database failed: %s\n", dberr)
        return nil, errors.New("Initial table creation in empty database failed")
      }
        fmt.Println(" successful.")
      } else {
        env = &Env{db: db, migrationVersion: DB_MIGRATION_VERSION}
        dberr = dbMigrateUp(env) // TBD: TEST THIS CODE!!!
        if dberr != nil {
        fmt.Printf(" Database migrations failed: %s\n", dberr)
        return nil, errors.New("Database migrations failed")
      }

      }
  } // if we need to initialize the db

  // create and return the persistence object
  return &PimPersistPostgreSQL{dbName:dbName, err:nil}, nil
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
  return &TaskDataMapperPostgreSQL{loaded:saved,dbName:dbName, err:nil}
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
    if (err != nil) { 
      err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert %s tag over task %s: %s", tagName, t.GetName(), err))
      return err
    }
  } else if !newTask {
    _, err := dbExec(env, `DELETE FROM task_tags WHERE task_id = $1 AND tag_id = $2`, t.GetId(), tagId)
    if (err != nil) { 
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

/*
==================================================================================
 syncTags()
----------------------------------------------------------------------------------
 This function takes a memory-based task and makes sure the tags on that task
 match the database version of the task.  Since we've not implemented "dirty"
 flags on the task to know what has changed (yet), we implement this by making DB
 calls to compare the tag state, and then making DB calls to force them to
 match.

 Two helper functions encapsulate some database functions, one to collect the
 tags from a task, another to collect all tags in the database.  Both of these
 functions returns a map where the key is the tag-name and the value is the id
 of the tag in the database.  Since the in-memory task only has strings, this
 helps us know the tag's id to set / remove individual tags.

 We may need to add some transactional control as to make sure we don't 
 so dumb things like create the same tag more than once if this code is
 executed in parallel.

 NOTE that when we're done here we'll create similar (not identical) code to sync 
 hyperlinks.  This will require the new task_links table coded (but never tested)
 in migration version 0004.
================================================================================*/
func (tm TaskDataMapperPostgreSQL) loadTaskTags(t *Task) (map[string]int, error) {
  taskTagsDB := make(map[string]int)
  taskQuery := fmt.Sprintf(`SELECT tags.name, tags.id FROM tags JOIN task_tags ON task_tags.tag_id = tags.id WHERE task_tags.task_id = '%s'`, 
                   t.GetId())
  taskTags, err := env.db.Query(taskQuery)
  if err != nil {
    log.Printf("query for the task tags failed: %s\n", taskQuery)
    return nil, err
    // log.Fatal(err) // maybe we should not just die here
  }
  defer taskTags.Close()
  for taskTags.Next() {
    var tagName string
    var tagId int
    err := taskTags.Scan(&tagName, &tagId)
    if err != nil {
      log.Printf("tmpg.syncTags(): row scan failed\n")
      return nil, err
      // log.Fatal(err)
    }
    taskTagsDB[tagName] = tagId
  }
  return taskTagsDB, err
}

// tbd - cache this (on the tm?) since it will be used over and over - but keeping
// the cache up to date as things get saved might be a pain.  If cached, this
// can abstract it - just return the cached map of tags.
func (tm TaskDataMapperPostgreSQL) loadAllTags() (map[string]int, error) {
  allTags := make(map[string]int)
  tagQuery := `SELECT tags.name, tags.id FROM tags`
  tags, err := env.db.Query(tagQuery)
  if err != nil {
    log.Printf("query for the tags failed: %s\n", tagQuery)
    return nil, err
  }
  defer tags.Close()
  for tags.Next() {
    var tagName string
    var dbTagId int
    err := tags.Scan(&tagName, &dbTagId)
    if err != nil {
      log.Printf("tmpg.syncTags(): row scan failed\n")
      return nil, err
    }
    allTags[tagName] = dbTagId
  }
  return allTags, nil
}

func (tm TaskDataMapperPostgreSQL) syncTags(t *Task, newTask bool) error {

  // LOCK NEEDED?

  // collect the list of all tags in the DB (someday just the ones for this user)
  allTags, err := tm.loadAllTags()
  if err != nil {
    return err
  }
  // log.Printf("syncTags(): allTags=%v\n", allTags)

  // collect the list of tags on the DB-version of this task in a modifiable form
  taskTagsDB, err := tm.loadTaskTags(t)
  if err != nil {
    return err
  }
  // log.Printf("syncTags(): taskTagsDB=%v\n", taskTagsDB)

  //   -- note that if "newTask" is true then there is no work to do here - always empty
  // for each tag on this in-memory task
  taskTagsMem := t.GetTags()
  // log.Printf("syncTags(): taskTagsMem=%v\n", taskTagsMem)  
  for _, tagMem := range taskTagsMem {

    // log.Printf("syncTags(): In loop for tag=%v\n", tagMem)

    // if the tag does not already exist in the DB then add the tag and link it to the task
    // note it would be more efficient to add all the tags at once, but the code gets ugly so
    // for now we'll add them one at a time.
    _, inDBAlready := allTags[tagMem]
    if !inDBAlready {
      // log.Printf("syncTags(): Tag <%v> is not in DB yet, adding to DB...\n", tagMem)

        var tagId int
        err := env.db.QueryRow(`INSERT INTO tags (name, system) VALUES ($1, FALSE) RETURNING id`, tagMem).Scan(&tagId)
      if (err != nil) { 
        err = errors.New(fmt.Sprintf("tdmp.syncTags(): Unable to insert tag %s: %s", tagMem, err))
        return err
      }
      
      // log.Printf("syncTags(): Tag <%v> is now in DB as id <%v>, adding to task id <%v>...\n", tagMem, tagId, t.GetId())
        _, err = dbExec(env, `INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)`, t.GetId(), tagId)
      if (err != nil) { 
        err = errors.New(fmt.Sprintf("tdmp.syncTags(): Unable to insert tag linkage %s to %s: %s", tagMem, t.GetName(), err))
        return err
      }

    } else {

      // log.Printf("syncTags(): Tag <%v> is in the DB...\n", tagMem)

      // else if the tag does already exist in the DB then
      // if the tag is not already on the DB-version of the task then
      tagId, setAlready := taskTagsDB[tagMem]
      if !setAlready {
        // log.Printf("syncTags(): Tag <%v> is in the DB but not on the task so assigning...\n", tagMem)

        // link the tag to the task
        tagId = allTags[tagMem]
          _, err := dbExec(env, `INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)`, t.GetId(), tagId)
        if (err != nil) { 
          err = errors.New(fmt.Sprintf("tdmp.syncTags(): Unable to insert tag linkage %s to %s: %s", tagMem, t.GetName(), err))
          return err
        }
      }

      // this tag is handled, remove it here so later we know only remaining ones
      // need to be unlinked from the task in the DB
      delete(taskTagsDB, tagMem)
    }

  }

  // for any "unused" remaining tags in the list of tags on the DB-version of this task
  // log.Printf("syncTags(): taskTagsDB <%v>, len(taskTagsDB) <%v>\n", taskTagsDB, len(taskTagsDB))
  numTagsToDelete := len(taskTagsDB)
  if numTagsToDelete > 0 {
    tagIdsToDelete := make([]int, numTagsToDelete)
    i := 0
    for _, tagId := range taskTagsDB {
      tagIdsToDelete[i] = tagId
      i += 1
    }

    // unlink (remove) the tags' relations to the task
    _, err := dbExec(env, "DELETE FROM task_tags WHERE task_id = $1 AND tag_id = ANY ($2)", t.GetId(), pq.Array(tagIdsToDelete))
    if err != nil {
      err = errors.New(fmt.Sprintf("tdmp.syncTags(): Unable to remove tags from task %s: %s", t.GetName(), err))
      return err
    }

  }

  // UNLOCK NEEDED?

  return nil
}

/*
==================================================================================
 syncLinks()
----------------------------------------------------------------------------------
 This function takes a memory-based task and makes sure the links on that task
 match the database version of the task.  Since we've not implemented "dirty"
 flags on the task to know what has changed (yet), we implement this by making DB
 calls to compare the links state, and then making DB calls to force them to
 match.

 A helper function encapsulates database functions and is used at both load and
 save time to callect the links from a task.  The function returns a map where the 
 key is the link URI and the value is the id of the tag in the database.  This may
 need to change to handle name offsets/lengths (see below).

 TBD: We may need to add some transactional control as to make sure we don't 
 so dumb things like create the same tag more than once if this code is
 executed in parallel.

 TBD: nameOffsets and nameLengths of tags are only partially coded so are not
 supported.  The fields are on the in-memory objects and are in the DB, but the
 mapper does not yet load or save them properly - only the URI is used.
================================================================================*/
func (tm TaskDataMapperPostgreSQL) loadTaskLinks(t *Task) (map[string]int, error) {
  taskLinksDB := make(map[string]int)
  taskQuery := fmt.Sprintf(`SELECT links.uri, links.nameOffset, links.nameLength, links.id FROM task_links AS links WHERE links.task_id = '%s'`, 
                   t.GetId())
  taskLinks, err := env.db.Query(taskQuery)
  if err != nil {
    log.Printf("query for the task links failed: %s (%s)\n", err, taskQuery)
    return nil, err
  }
  defer taskLinks.Close()
  for taskLinks.Next() {
    var linkUri string
    var linkNameOffset int
    var linkNameLen int
    var linkId int
    err := taskLinks.Scan(&linkUri, &linkNameOffset, &linkNameLen, &linkId)
    if err != nil {
      log.Printf("tmpg.syncLinks(): row scan failed\n")
      return nil, err
    }
    // TBD: use the nameOffset and nameLength
    taskLinksDB[linkUri] = linkId
  }
  return taskLinksDB, err
}

func (tm TaskDataMapperPostgreSQL) syncLinks(t *Task, newTask bool) error {

  // collect the list of tags on the DB-version of this task in a modifiable form
  taskLinksDB, err := tm.loadTaskLinks(t)
  if err != nil {
    return err
  }

  // for each link on this in-memory task
  taskLinksMem := t.GetTaskLinks()
  for _, linkMem := range taskLinksMem {
    // if the link does not already exist in the DB then add the link and link it to the task
    // note it would be more efficient to add all the links at once, but the code gets ugly so
    // for now we'll add them one at a time.
    _, inDBAlready := taskLinksDB[linkMem.GetURI()]
    if !inDBAlready {
        var linkId int
        err := env.db.QueryRow(`INSERT INTO task_links (uri, nameOffset, nameLength, task_id) VALUES ($1, $2, $3, $4) RETURNING id`, 
                               linkMem.GetURI(), linkMem.NameOffset, linkMem.NameLen, t.GetId()).Scan(&linkId)
      if (err != nil) { 
        err = errors.New(fmt.Sprintf("tdmp.syncLinks(): Unable to insert link %s: %s", linkMem.GetURI(), err))
        return err
      }

    // otherwise it is there, but still check if we need to update the name / len
    // but for now we don't update the name / len so nothing to do but remove from list
    } else {    
      // this link is handled, remove it here so later we know only remaining ones
      // need to be removed
      delete(taskLinksDB, linkMem.GetURI())
    }
  } // for each link on the in-memory task

  // for any "unused" remaining links in the list of links on the DB-version of this task
  // log.Printf("syncLinks(): taskLinksDB <%v>, len(taskLinksDB) <%v>\n", taskLinksDB, len(taskLinksDB))
  numLinksToDelete := len(taskLinksDB)
  if numLinksToDelete > 0 {
    linkIdsToDelete := make([]int, numLinksToDelete)
    i := 0
    for _, linkId := range taskLinksDB {
      linkIdsToDelete[i] = linkId
      i += 1
    }

    // remove the links from the task
    _, err := dbExec(env, "DELETE FROM task_links WHERE id = ANY ($1)", pq.Array(linkIdsToDelete))
    if err != nil {
      err = errors.New(fmt.Sprintf("tdmp.syncLinks(): Unable to remove links from task %s: %s", t.GetName(), err))
      return err
    }
  }

  return nil
}

/*
==================================================================================
 loadTaskUsers()
----------------------------------------------------------------------------------
 Inputs: t *Task - the task on which to decorate the users that have access
         error   - either a DB error or the user in the DB is not loaded in mem

 Given a loaded task, go find the users that have access to the task and give
 those users access to the task by assigning them to the task in memory.
================================================================================*/
func (tm TaskDataMapperPostgreSQL) loadTaskUsers(t *Task) (Users, error) {
  us := make(Users, 0)
  taskQuery := fmt.Sprintf(`SELECT tu.user_id FROM task_users AS tu WHERE tu.task_id = '%s'`, t.GetId())
  taskUsers, err := env.db.Query(taskQuery)
  if err != nil {
    log.Printf("query for the task users failed: %s (%s)\n", err, taskQuery)
    return nil, err
  }
  defer taskUsers.Close()
  for taskUsers.Next() {
    var userId string
    errScan := taskUsers.Scan(&userId)
    if errScan != nil {
      log.Printf("tmpg.loadTaskUsers(): row scan failed\n")
      return nil, errScan
    }

    // we assume all users are already loaded in our global list
    // so we merely have to find it - but ugly - have to get it
    // off a global list :-(
    u := users.FindById(userId)
    if u == nil {
      log.Printf("User %s referenced on task in DB is not loaded in memory - failing.\n", userId)
      return nil, errors.New("User referenced on task in DB is not loaded in memory - failing.")
    }
    us = append(us, u)

  } // for each userid found

  return us, err
}

/*
==================================================================================
 syncUsers()
----------------------------------------------------------------------------------
 Inputs: t *Task - task whose list of users with access must be synced to the DB
 Return: error   - either a DB error or syncrhonization error with DB

 Given an in-memory task, make the DB match it's list of users that have access
 to that task.  This is tricky: we have to get the list of users that have
 access to the task from both the DB and memory, add any not already in the DB
 and delete any that ARE in the DB but are not on the in-memory task.
================================================================================*/
func (tm TaskDataMapperPostgreSQL) syncUsers(t *Task) error {

  log.Printf("syncUsers(): Enterd for task <%s> with %v users.\n", t.GetName(), len(t.GetUsers()))

  // collect the list of users in the DB for this task
  // note this list can be changed in this function - it is our own copy to play with
  usersDB, err := tm.loadTaskUsers(t)
  if err != nil {
    return err
  }

  // for each user on this in-memory task
  us := t.GetUsers()
  for _, u := range us {
    // if the user does not already exist in the DB then add the user and link it to the task
    // note it would be more efficient to add all the users at once, but the code gets ugly so
    // for now we'll add them one at a time.
    inDBAlready := usersDB.FindById(u.GetId())
    if inDBAlready == nil {
      _, err := dbExec(env, `INSERT INTO task_users (user_id, task_id) VALUES ($1, $2)`, u.GetId(), t.GetId())
      if (err != nil) { 
        err = errors.New(fmt.Sprintf("tdmp.syncLinks(): Unable to insert user access %s: %s", u.GetEmail(), err))
        return err
      }

    // otherwise it is there so we don't have to add it, but we want to remove
    // if from the usersDB list, since anything we don't "process" we will 
    // later know needs to be deleted from the DB
    } else {    
      idx := usersDB.IndexOf(u)
      if idx != -1 { // this should never happen!  assert?
        usersDB = append(usersDB[:idx], usersDB[idx+1:]...)
      }
    }
  } // for each user on the in-memory task

  // remove any "unused" remaining users in the list of links on the DB-version of this task
  for _, uDelete := range usersDB {
    _, err := dbExec(env, "DELETE FROM task_users WHERE user_id = $1 AND task_id = %2", uDelete.GetId(), t.GetId())
    if err != nil {
      err = errors.New(fmt.Sprintf("tdmp.syncUsers(): Unable to remove user access from task %s: %s", t.GetName(), err))
      return err
    }
  }

  return nil
}


/*
==================================================================================
 Save()
----------------------------------------------------------------------------------
 Inputs: t            *Task - the in-memory task to save
     saveChildren bool  - whether or not to save the child tasks
     careMyself   bool  - whether or not to save the task "t" itself (this is
                          useful if the app is using a parent task just to
                          group tasks, but doesn't want to save them)

 This function writes the provided in-memory task into the PostgreSQL database.
================================================================================*/
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
      // tm.syncSystemTags(t, false)
      err = tm.syncTags(t, false)
      if (err != nil) {
        return err;
      }

      // update all links - adding or removing to match the in-memory task
      err = tm.syncLinks(t, true)
      if (err != nil) {
        return err;
      }

      // update all users - adding or removing to match the in-memory task
      err = tm.syncUsers(t)
      if (err != nil) {
        return err;
      }


    } else {
        _, err := dbExec(env, `INSERT INTO tasks (id, name, state, target_start_time, actual_start_time, actual_completion_time, estimate_minutes) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, 
                       t.GetId(), t.GetName(), t.GetState(), t.TargetStartTime, t.ActualStartTime, t.ActualCompletionTime, int(t.Estimate.Minutes()))
      if (err != nil) { 
        err = errors.New(fmt.Sprintf("tdmp.Save(): Unable to insert task %s: %s", t.GetName(), err))
        return err
      }

      // update all tags - adding or removing to match the in-memory task
      err = tm.syncTags(t, true)
      if (err != nil) {
        return err;
      }

      // update all links - adding or removing to match the in-memory task
      err = tm.syncLinks(t, true)
      if (err != nil) {
        return err;
      }

      // update all users - adding or removing to match the in-memory task
      err = tm.syncUsers(t)
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
          if err != nil { 
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
                         db_estimate_minutes       sql.NullInt64) error {
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

/*
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
*/

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
func (tm TaskDataMapperPostgreSQL) loadAndSetTags(t *Task) error {
  tagMap, err := tm.loadTaskTags(t)
  if err != nil {
    return err
  }
  for tag, _ := range tagMap {
    t.SetTag(tag)
  }
  return nil
}

func (tm TaskDataMapperPostgreSQL) loadAndSetLinks(t *Task) error {
  linkMap, err := tm.loadTaskLinks(t)
  if err != nil {
    return err
  }
  for link, _ := range linkMap {
    t.AddLink(link, 0, 0)
  }
  return nil
}

func (tm TaskDataMapperPostgreSQL) loadAndSetUsers(t *Task) error {
  us, err := tm.loadTaskUsers(t)
  if err != nil {
    return err
  }
  for _, u := range us {
    t.AddUser(u)
  }
  return nil
}


func (tm TaskDataMapperPostgreSQL) Load(t *Task, loadChildren bool, root bool) error {

  var (
    name string
    state TaskState
    db_target_start_time pq.NullTime
    db_actual_start_time pq.NullTime
    db_actual_completion_time pq.NullTime
    db_estimate_minutes sql.NullInt64
  )

  // if we're told this is "root" that means we should not attempt
  // to load into "this" task, just move on to the kids
  if (!root) {

    // build and execute the query for the task
    taskQuery := `SELECT t.name, t.state, t.target_start_time, t.actual_start_time, t.actual_completion_time, t.estimate_minutes 
                  FROM tasks t
                  WHERE id = '" + t.GetId() + "'" + "
                  GROUP BY t.id`
    err := env.db.QueryRow(taskQuery).Scan(&name, &state, &db_target_start_time, &db_actual_start_time, &db_actual_completion_time, &db_estimate_minutes)
    if err != nil {
      // log.Printf("query for a task failed: %s, err: %s\n", taskQuery, err)
      return err
    }

    // overwrite my in-memory values
    t.SetName(name)
    t.SetState(state)
    tm.setTaskFields(t, db_target_start_time, db_actual_start_time, db_actual_completion_time, db_estimate_minutes)

    // now go get and set the tags
    err = tm.loadAndSetTags(t)
    if err != nil {
      return err
    }

    // now go get and set the links
    err = tm.loadAndSetLinks(t)
    if err != nil {
      return err
    }

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
  )

  var baseQuery string = `SELECT t.id, t.name, t.state, t.target_start_time, t.actual_start_time, t.actual_completion_time, t.estimate_minutes
                          FROM tasks t
                        LEFT JOIN task_parents tp ON tp.child_id = t.id
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
    err := rows.Scan(&dbid, &dbname, &dbstate, &db_target_start_time, &db_actual_start_time, &db_actual_completion_time, &db_estimate_minutes)
    if err != nil {
      log.Printf("tmpg.loadChildren(): row scan failed\n")
      log.Fatal(err)
    }
    // log.Printf("LoadChildren(): read id=%s, name=%s\n", id, name)

    // create the child task
    k := &Task{id:dbid, name:dbname, state:dbstate}
    tm.setTaskFields(k, db_target_start_time, db_actual_start_time, db_actual_completion_time, db_estimate_minutes)

    // load and set the tags
    err = tm.loadAndSetTags(k)
    if err != nil {
      return err
    }

    // now go get and set the links
    err = tm.loadAndSetLinks(k)
    if err != nil {
      return err
    }

    // now go get and set the user assignments
    err = tm.loadAndSetUsers(k)
    if err != nil {
      return err
    }    

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
    err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to delete child references to task %s with id %s: %s", t.GetName(), t.GetId(), err))
    return err
  }

  // delete all tag references from the task_tags table
  _, err = dbExec(env, "DELETE FROM task_tags WHERE task_id = $1", t.GetId())
  if err != nil { // we should make sure this doesn't return an error if no tags are on the task - if it does we should eat that error
    err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove tags from task %s with id %s: %s", t.GetName(), t.GetId(), err))
    return err
  }

  // delete all link references from the task_links table
  _, err = dbExec(env, "DELETE FROM task_links WHERE task_id = $1", t.GetId())
  if err != nil { // we should make sure this doesn't return an error if no links are on the task - if it does we should eat that error
    err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove links from task %s with id %s: %s", t.GetName(), t.GetId(), err))
    return err
  }

  // delete all user references from the task_users table
  _, err = dbExec(env, "DELETE FROM task_users WHERE task_id = $1", t.GetId())
  if err != nil { // we should make sure this doesn't return an error if no users are on the task - if it does we should eat that error
    err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove users from task %s with id %s: %s", t.GetName(), t.GetId(), err))
    return err
  }  

  // delete this task from the tasks table - must do this after deleting from
  // parent table
  _, err = dbExec(env, "DELETE FROM tasks WHERE id = $1", t.GetId())
  if err != nil {
    err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove task <%s> with id %s: %s", t.GetName(), t.GetId(), err))
    return err
  }

  // clean in-memory tm structures
  tm.loaded = false
  tm.parentIds = nil

  return nil
}

/*
=============================================================================
 UserSave()
-----------------------------------------------------------------------------
 Inputs:  User u - user to save to the database
 Returns: error  - DB call could fail - likely cause is bad DB

 Save the current in-memory version of the user to the database.  This is a
 smart "upsert" function which intelligently decides if an UPDATE or INSERT
 is required based on the state of the in-memory object.  Note that it is
 NOT smart enough to know this from the DB - so if this code becomes
 re-entrant or we envision multiple servers changing users, we'd need to
 make this transactional and smarter to check the stored DB for state
 instead if tm.loaded.

 TBD: Implement some kind of dirty flag to minimize unneeded UPDATES.
===========================================================================*/
func (tm *TaskDataMapperPostgreSQL) UserSave(u *User) error {
  log.Printf("UserSave(%v)\n",u)

  if tm.loaded {
    _, err := dbExec(env, `UPDATE users SET name = $1, email = $2, password = $3 
                         WHERE ID = $4`, u.GetName(), u.GetEmail(), u.GetPassword(), u.GetId())
    if err != nil {
      errStr := fmt.Sprintf("tdmp.UserSave(): Unable to update user %s: %s\n", u.GetEmail(), err)
      log.Printf(errStr)
      err = errors.New(errStr)
      return err
    }

  } else {
      _, err := dbExec(env, `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4) RETURNING id`, 
                     u.GetId(), u.GetName(), u.GetEmail(), u.GetPassword())
    if err != nil {
      errStr := fmt.Sprintf("tdmp.UserSave(): Unable to insert task %s: %s\n", u.GetEmail(), err)
      log.Printf(errStr)
      err = errors.New(errStr)
      return err
    }
    tm.loaded = true
  }
  return nil
}

/*
=============================================================================
 UserLoad()
-----------------------------------------------------------------------------
 Inputs:  User u - empty object to fill up (TBD: shouldn't we return a user?)
 Returns: error  - DB call could fail - likely cause is foreign keys exist

 NOTE: Not yet run or tested!  Have only loaded users with UserLoadAll().
       This function should probably be re-written to return a user and
       take an ID as input?

 Load the user specified in the ID of the supplied user.
===========================================================================*/
func (tm *TaskDataMapperPostgreSQL) UserLoad(u *User) error {
  var (
    dbname string
    dbemail string
    dbpassword string
  )

  // build and execute the query for the task
  taskQuery := `SELECT u.name, u.email, u.password
                FROM users u
                WHERE id = '" + u.GetId()`
  err := env.db.QueryRow(taskQuery).Scan(&dbname, &dbemail, &dbpassword)
  if err != nil {
    // log.Printf("query for a user failed: %s, err: %s\n", taskQuery, err)
    return err
  }

  // overwrite my in-memory values
  u.SetName(dbname)
  u.SetEmail(dbemail)
  u.SetHashedPassword(dbpassword)

  // set myself as loaded from the DB
  tm.loaded = true
  return nil
}

/*
=============================================================================
 UserDelete()
-----------------------------------------------------------------------------
 Inputs:  User u - user object to delete
 Returns: error  - DB call could fail - likely cause is foreign keys exist

 NOTE: Not yet run or tested!

 Delete the specified user.  Note that this function assumes that all tasks
 and other foreign keys to this user have already been deleted.  If that is
 not the case then this function will fail and return an error.
===========================================================================*/
func (tm *TaskDataMapperPostgreSQL) UserDelete(u *User) error {

  // if the user has never been saved then no work to here
  if !tm.loaded {
    return nil
  }

  // delete this user
  _, err := dbExec(env, "DELETE FROM users WHERE id = $1", u.GetId())
  if err != nil {
    err = errors.New(fmt.Sprintf("tdmp.Delete(): Unable to remove user %s with id %d: %s", u.GetEmail(), u.GetId(), err))
    return err
  }

  // clean in-memory tm structures
  tm.loaded = false
  return nil
}

/*
=============================================================================
 UserLoadAll()
-----------------------------------------------------------------------------
 Returns: Users  - list of all users known to the system
          error  - DB calls could fail, or problems creating a user

 This function is used to load _all_ users in the system.  It is
 intended to be called just once at system startup to bootstrap the server
 with all known users.
===========================================================================*/
func (tm *TaskDataMapperPostgreSQL) UserLoadAll() (Users, error) {

  // log.Printf("LoadChildren(): for parent task %s\n", parent.name)
  var (
    dbid string
    dbname string
    dbemail string
    dbpassword string
  )

  var sqlSelect string = `SELECT u.id, u.name, u.email, u.password FROM users u`
  rows, err := env.db.Query(sqlSelect)
  if err != nil {
    log.Printf("query for users failed: %s\n", sqlSelect)
    log.Fatal(err) // maybe we should not just die here
  }
  defer rows.Close()

  // allocate the slice to collect the users
  us := make(Users, 0)

  // for each user in our database of users
  for rows.Next() {
    err := rows.Scan(&dbid, &dbname, &dbemail, &dbpassword)
    if err != nil {
      log.Printf("tmpg.UserLoadAll(): row scan failed\n")
      log.Fatal(err)
    }
    log.Printf("tmpg.UserLoadAll(): read id=%s, name=%s\n", dbid, dbemail)

    // create the user and add to the list
    utm := tm.CopyDataMapper()
    u, errid := LoadUser(dbid, dbname, dbemail, dbpassword, utm)
    if errid != success {
      log.Printf("tmpg.UserLoadAll(): user creation failed\n")
      log.Fatal(pimError(errid))
    }
    us = append(us, u)
  }

  err = rows.Err()
  if err != nil {
    log.Fatal(err) // need nicer errors
  } 
  return us, nil
}
