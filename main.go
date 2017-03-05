package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
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

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	leaderboard := &LeaderBoard{db: db}
	playerstats := &PlayerStats{db: db}

	statgroup := router.Group("/stats")
	{
		statgroup.GET("/leaderboard/*length", leaderboard.GetLeader)
		statgroup.POST("/leaderboard", leaderboard.PostResult)
		statgroup.GET("/placements/:id", playerstats.GetPlacements)
		statgroup.GET("/headtohead/:id1/:id2", playerstats.GetHeadToHead)
	}

	router.Run(":8080")
}
