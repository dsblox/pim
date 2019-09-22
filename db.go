package main

import (
    "database/sql"
    _ "github.com/lib/pq"
    "os"
    "bufio"
    "fmt"
    "errors"
    "path/filepath"
)


func NewDB(dataSourceName string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dataSourceName)
    if err != nil {
        return nil, err
    }
    if err = db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}




func dbExec(env *Env, sqlStr string, args ...interface{}) (sql.Result, error) {
    result, err := env.db.Exec(sqlStr, args...)
    if err != nil {
        return nil, err
    }
    return result, err
}

// returns id of inserted row if no error - NOTE THAT THE CALLER MUST USE
// THE RETURNING CLAUSE IN THEIR SQL STRING - TBD: Use prepare statements
// to make this automatic
func dbInsert(env *Env, sqlStr string, args ...interface{}) (int, error) {
    var id int
    err := env.db.QueryRow(sqlStr, args...).Scan(&id)
    if err == sql.ErrNoRows {
        err = nil // ErrNoRows is expected from a single row insert
    }
    return id, err
}

func dbCreate(env *Env, dbName string) (sql.Result, error) {
    return dbExec(env, "CREATE DATABASE $1", dbName) 
}

func dbUse(env *Env, dbName string) (sql.Result, error) {
    return dbExec(env, "USE DATABASE $1", dbName)
}

/*
===============================================================================
 dbMigrateFilename()
-------------------------------------------------------------------------------
 Inputs:  version  int    - number from 0-9999 for this migration version
          modifier string - UP, DOWN or ORIGIN identifying the file with SQL
 Returns:          string - absolute filename or "" if a error occurs
                   error  - not nil if error building the file name 
                            or if the version file does not exist

 Return the absolute path to the filename associated with the version number
 provided and the type of file requested (UP, DOWN, CLEAN).
=============================================================================*/
func dbMigrateFilename(version int, modifier string) (string, error) {
    if version < 0 || version > 9999 {
        return "", errors.New("invalid version: must be 0-9999")
    }
    if modifier != "UP" && modifier != "DOWN" && modifier != "CLEAN" {
        return "", errors.New("invalid modifier: must be UP, DOWN or CLEAN")        
    }

    // build the filename
    relativePath := "db/migrations/"
    filename := fmt.Sprintf("%04d-", version) + modifier + ".sql"
    absPath, _ := filepath.Abs(relativePath + filename)

    // make sure it exists
    _, err := os.Stat(absPath)
    if err != nil {
        if os.IsNotExist(err) {
            return "", err
        }
    }
    return absPath, nil
}

// open and execute all SQL comments in the specified file
// we stop on any error, and return any SQL error from
// that execution.  all SQL commands in the file must end
// with a semi-colon.
func dbExecuteSQLFile(env *Env, filename string) error {

    // open the file specified
    file, err := os.Open(filename)
    if err != nil {
        fmt.Printf("Unable to open SQL file: %s, error: %s\n", filename, err)
        return err
    }
    defer file.Close()
    reader := bufio.NewReader(file)

    // for each command in the file
    var sqlErr error = nil;
    var sqlCmd string = "XXXX";
    for ; sqlErr == nil && len(sqlCmd) > 3;  {

        // read the command
        sqlCmd, sqlErr = reader.ReadString(';')

        // run the command
        fmt.Printf("SQL Command: %s\n", sqlCmd)
        _, sqlErr = dbExec(env, sqlCmd);

        // on error just return it and let client deal with it
        if sqlErr != nil {
            return sqlErr;
        }
    }
    
    return nil    
}

/*
===============================================================================
 dbMigrateOrigin()
-------------------------------------------------------------------------------
 This function looks for the CLEAN db migraiton file for the current version
 which contains the commands needed to initialize a completely empty database
 for the application.  Note that each migration version provides a full CLEAN
 file which includes all needed tables for the application.
=============================================================================*/
func dbMigrateOrigin(env *Env) error {

    // build / find the file name
    filename, err := dbMigrateFilename(env.migrationVersion, "CLEAN")
    if filename == "" {
        fmt.Printf("Unable to find database origin file: %s\n", err)
        return err
    }

    err = dbExecuteSQLFile(env, filename)

    // record the new version in the migration table
    if err == nil {
        dbInsert(env, "INSERT INTO migrations (version_applied, file_applied) VALUES ($1, $2)", env.migrationVersion, filename)
    }

    return err
}

/*
===============================================================================
 dbMigrateDBVersion()
-------------------------------------------------------------------------------
 Look up the version from the database.  If there is no migration table in
 the database then this simply returns -1.
=============================================================================*/
func dbMigrateDBVersion(env *Env) int {
    var version int = -1
    env.db.QueryRow("SELECT MAX(version_applied) FROM migrations").Scan(&version)
    return version;
}

/*
===============================================================================
 dbMigrateUp()
-------------------------------------------------------------------------------
 This function checks the migration table of the database, and applies any
 needed migrations it finds in the db/migrations directory, in order, to
 upgrade as needed.
=============================================================================*/
func dbMigrateUp(env *Env) error {

    // target is what the code expects, current is what the DB says it is now
    targetVersion := env.migrationVersion
    currentVersion := dbMigrateDBVersion(env)
    fmt.Printf("Code requires DB version %d and database is version %d\n", targetVersion, currentVersion)

    // loop through and execute all migration files found from current+1 to target
    var err error = nil
    for nextVersion := currentVersion + 1; 
        nextVersion <= targetVersion && err == nil; // stop on any SQL error
        nextVersion++ {
        filename, err := dbMigrateFilename(nextVersion, "UP")
        if err != nil {
            fmt.Printf("Warning: skipping no UP file found for version %d, error: %s\n", nextVersion, err)
            err = nil 
        }
        fmt.Printf("Migrating DB applying <%s>\n", filename)
        err = dbExecuteSQLFile(env, filename)

        // record the new version in the migration table
        if err == nil {
            dbInsert(env, "INSERT INTO migrations (version_applied, file_applied) VALUES ($1, $2)", nextVersion, filename)
        }
    }

    return err
}


