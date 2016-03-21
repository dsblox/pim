package main

import (
    "database/sql"
    _ "github.com/lib/pq"
    "log"
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
        log.Fatal(err)
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

