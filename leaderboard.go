package main

import (
    "database/sql"
    "fmt"
    "strconv"

    "github.com/gin-gonic/gin"
)

type Player struct {
    Id      uint    `json:"id"`
    Name    string  `json:"name"`
    Played  uint    `json:"played"`
    Wins    uint    `json:"wins"`
    Losses  uint    `json:"losses"`
    Score   float32 `json:"score"`
}

type GamePlayer struct {
    Id      *uint   `json:"id" binding:"exists"`
    Name    string  `json:"name" binding:"required"`
    Pos     *uint    `json:"pos" binding:"exists"`
}

type Game struct {
    Id          *uint           `json:"id" binding:"exists"`
    Players     []GamePlayer    `json:"players" binding:"required"`
}

type LeaderBoard struct {
    db *sql.DB
}

func (lb *LeaderBoard) GetLeader(c *gin.Context) {
    var (
        player  Player
        players []Player
        length  int
    )
    lenStr := c.Param("length")[1:]
    lenInt, _ := strconv.Atoi(lenStr)
    length = int(lenInt)

    query := "select id, name, played, wins, losses, score from leaderboard order by score desc, wins desc"
    if length > 0 {
        query = query + " limit " + strconv.Itoa(length)
    }

    rows, err := lb.db.Query(query)
    if err != nil {
        fmt.Println(err.Error())
        c.JSON(500, `{"error":"sql error"}`)
        return
    }
    for rows.Next() {
        err = rows.Scan(&player.Id, &player.Name, &player.Played, &player.Wins, &player.Losses, &player.Score)
        player.Score = 200 * ((player.Score + 1) / 3) - 100
        players = append(players, player)
        if err != nil {
            if err == sql.ErrNoRows {
            } else {
                rows.Close()
                fmt.Println(err.Error())
                c.JSON(500, `{"error":"sql error"}`)
                return
            }
        }
    }
    defer rows.Close()

    c.JSON(200, players)
}

func (lb *LeaderBoard) PostResult(c *gin.Context) {
    var (
        game Game
        numplayers int
    )

    err := c.Bind(&game)
    if err != nil {
        fmt.Println(err.Error())
        c.JSON(400, `{"error":"Invalid Data"}`)
        return
    }

    getstmt, err := lb.db.Prepare("select ifnull(played, 0), ifnull(wins, 0), ifnull(losses, 0), ifnull(score, 0.0) from leaderboard where id = ?;")
    if err != nil {
        fmt.Println(err.Error())
        c.JSON(500, `{"error":"sql error"}`)
        return
    }
    defer getstmt.Close()

    //TODO: combine all players into one insert/update
    updatestmt, err := lb.db.Prepare("insert into leaderboard (id, name, played, wins, losses, score) values (?,?,?,?,?,?) on duplicate key update played=values(played),wins=values(wins),losses=values(losses),score=values(score);")
    if err != nil {
        fmt.Println(err.Error())
        c.JSON(500, `{"error":"sql error"}`)
        return
    }
    defer updatestmt.Close()


    numplayers = len(game.Players)

    for _, v := range game.Players {
        var (
            player Player
            interpolatedscore float32
        )

        if v.Id != nil  && v.Pos != nil {
            err := getstmt.QueryRow(v.Id).Scan(&player.Played, &player.Wins, &player.Losses, &player.Score)
            if err != nil {
                if err == sql.ErrNoRows {
                } else {
                    fmt.Println(err.Error())
                    c.JSON(500, `{"error":"error retrieving db data"}`)
                    return
                }
            }
            if numplayers <= 1 {
                interpolatedscore = 1
            } else {
                interpolatedscore = (1 - (float32(*v.Pos) / float32(numplayers - 1)))
            }
            player.Score = player.Score * float32(player.Played)
            player.Played = player.Played + 1
            //score
            player.Score = player.Score + interpolatedscore
            //winner
            if interpolatedscore == 1 {
                player.Wins = player.Wins + 1
                //winning bonus
                player.Score = player.Score + 1
            }
            //loser
            if interpolatedscore == 0 {
                player.Losses = player.Losses + 1
                //losing penalty
                player.Score = player.Score - 1
            }

            //recalculate average score
            player.Score = player.Score / float32(player.Played)

            //now update database
            _, err2 := updatestmt.Exec(v.Id, v.Name, player.Played, player.Wins, player.Losses, player.Score)
            if err2 != nil {
                fmt.Println(err2.Error())
                c.JSON(500, `{"error":"error updating db"}`)
                return
            }
        }
    }

    c.JSON(200, `{"success":"Updated"}`)
}
