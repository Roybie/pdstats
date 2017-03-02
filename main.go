package main

import (
    "database/sql"
    "fmt"

    //"github.com/roybie/pdstats/leaderboard"
    _ "github.com/go-sql-driver/mysql"
    "github.com/gin-gonic/gin"
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

    router := gin.Default()

    leaderboard := &LeaderBoard{db: db}

    router.GET("/leaderboard/*length", leaderboard.GetLeader)
    router.POST("/leaderboard", leaderboard.PostResult)

    router.Run(":8080")
}
