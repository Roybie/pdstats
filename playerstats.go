package main

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlayerPlacements map[string]uint

type HeadToHeadGame struct {
	Id      uint               `json:"id"`
	Players []HeadToHeadPlayer `json:"players"`
}

type HeadToHeadPlayer struct {
	Id       uint `json:"id"`
	Position uint `json:"position"`
}

type PlayerStats struct {
	db *sql.DB
}

func (ps *PlayerStats) GetPlacements(c *gin.Context) {
	var (
		pos        uint
		numplayers uint
	)

	idStr := c.Param("id")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, "Invalid ID")
		return
	}

	playerPlacement := make(PlayerPlacements)

	query := "select position, numplayers from games_users where player = " + idStr

	rows, err := ps.db.Query(query)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(500, `{"error":"sql error"}`)
		return
	}
	for rows.Next() {
		err = rows.Scan(&pos, &numplayers)
		if err != nil {
			if err == sql.ErrNoRows {
			} else {
				rows.Close()
				fmt.Println(err.Error())
				c.JSON(500, `{"error":"sql error"}`)
				return
			}
		}
		if pos == numplayers-1 {
			playerPlacement["last"]++
		} else {
			playerPlacement[strconv.Itoa(int(pos)+1)]++
		}
	}
	defer rows.Close()

	c.JSON(200, playerPlacement)
}

func (ps *PlayerStats) GetHeadToHead(c *gin.Context) {
	var (
		game     uint
		player   uint
		position uint
	)
	id1 := c.Param("id1")
	id2 := c.Param("id2")
	headStats := make([]HeadToHeadGame, 0)
	checkGame := make(map[uint]HeadToHeadGame)
	if _, err := strconv.Atoi(id1); err != nil {
		c.JSON(400, "Invalid ID")
		return
	}
	if _, err := strconv.Atoi(id2); err != nil {
		c.JSON(400, "Invalid ID")
		return
	}

	query := "select ag.game, player, position from games_users as ag join (select game from games_users where player in (" + id1 + "," + id2 + ") group by game having count(*) > 1) as gt on ag.game = gt.game where player in(" + id1 + "," + id2 + ");"

	rows, err := ps.db.Query(query)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(500, `{"error":"sql error"}`)
		return
	}
	for rows.Next() {
		err = rows.Scan(&game, &player, &position)
		if err != nil {
			if err == sql.ErrNoRows {
			} else {
				rows.Close()
				fmt.Println(err.Error())
				c.JSON(500, `{"error":"sql error"}`)
				return
			}
		}
		if g, ok := checkGame[game]; ok {
			p := HeadToHeadPlayer{Id: player, Position: position}
			g.Players = append(g.Players, p)
			headStats = append(headStats, g)
		} else {
			p := HeadToHeadPlayer{Id: player, Position: position}
			g := HeadToHeadGame{Id: game}
			g.Players = append(g.Players, p)
			checkGame[game] = g
		}
	}
	defer rows.Close()

	c.JSON(200, headStats)
}
