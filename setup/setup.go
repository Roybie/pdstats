package main

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/pdstats")
    if err != nil {
        fmt.Println(err.Error())
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        fmt.Println(err.Error())
    }

    stmt, err := db.Prepare("CREATE TABLE leaderboard (id int UNSIGNED, name varchar(50) NOT NULL, played int UNSIGNED DEFAULT 0, wins int UNSIGNED DEFAULT 0, losses int UNSIGNED DEFAULT 0, score double DEFAULT 0, PRIMARY KEY (id));")
    if err != nil {
        fmt.Println(err.Error())
    }

    _, err = stmt.Exec()
    if err != nil {
        fmt.Println(err.Error())
    } else {
        fmt.Println("Leaderboard table successfully created...")
    }
}
