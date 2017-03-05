package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
)

type Player struct {
	Id   int
	Name string
}

type Game struct {
	Id      int
	Players []Player
}

type LBPlayer struct {
	Id     uint    `json:"id"`
	Name   string  `json:"name"`
	Played uint    `json:"played"`
	Wins   uint    `json:"wins"`
	Losses uint    `json:"losses"`
	Score  float32 `json:"score"`
}

func main() {
	var (
		player Player
		game   Game
		games  []Game
	)
	var players = make(map[int]string)

	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/pusoy_dos")
	if err != nil {
		fmt.Println(err.Error())
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}

	rows, err := db.Query("select id, name from user;")
	if err != nil {
		fmt.Println(err.Error())
	}

	for rows.Next() {
		err = rows.Scan(&player.Id, &player.Name)
		players[player.Id] = player.Name
		if err != nil {
			if err == sql.ErrNoRows {
			} else {
				rows.Close()
				fmt.Println(err.Error())
				return
			}
		}
	}
	rows.Close()

	fmt.Println(players)

	//now get game results
	rows, err = db.Query("select game, current_player, winners from round inner join game on round.game = game.id where game.complete = 1;")

	if err != nil {
		fmt.Println(err.Error())
	}

	for rows.Next() {
		var (
			winStr      string
			winners     []string
			loser       int
			gamePlayers []Player
		)
		err = rows.Scan(&game.Id, &loser, &winStr)
		if err != nil {
			if err == sql.ErrNoRows {
			} else {
				rows.Close()
				fmt.Println(err.Error())
				return
			}
		}
		winners = strings.Split(winStr[1:len(winStr)-1], ",")
		winners = append(winners, strconv.Itoa(loser))
		for _, v := range winners {
			pid, _ := strconv.Atoi(v)
			gamePlayers = append(gamePlayers, Player{Id: pid, Name: players[pid]})
		}
		game.Players = gamePlayers
		games = append(games, game)
	}
	rows.Close()

	db.Close()
	fmt.Println(games)
	//now insert into leaderboard table

	db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/pdstats")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE IF EXISTS leaderboard;")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = db.Exec("CREATE TABLE leaderboard (id int UNSIGNED, name varchar(50) NOT NULL, played int UNSIGNED DEFAULT 0, wins int UNSIGNED DEFAULT 0, losses int UNSIGNED DEFAULT 0, score double DEFAULT 0, PRIMARY KEY (id));")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	_, err = db.Exec("DROP TABLE IF EXISTS games_users;")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = db.Exec("CREATE TABLE games_users (game int UNSIGNED, player int UNSIGNED, position int UNSIGNED, numplayers int UNSIGNED, PRIMARY KEY (game, player));")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, g := range games {
		updateLeader(db, g)
	}
}

func updateLeader(db *sql.DB, game Game) {

	getstmt, err := db.Prepare("select ifnull(played, 0), ifnull(wins, 0), ifnull(losses, 0), ifnull(score, 0.0) from leaderboard where id = ?;")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer getstmt.Close()

	//TODO: combine all players into one insert/update
	updatestmt, err := db.Prepare("insert into leaderboard (id, name, played, wins, losses, score) values (?,?,?,?,?,?) on duplicate key update played=values(played),wins=values(wins),losses=values(losses),score=values(score);")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer updatestmt.Close()

	gamestmt, err := db.Prepare("insert into games_users (game, player, position, numplayers) values (?,?,?,?);")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer gamestmt.Close()

	numplayers := len(game.Players)

	for _, p := range game.Players {
		if len(p.Name) == 0 {
			return
		}
	}
	for pos, v := range game.Players {
		var (
			player            LBPlayer
			interpolatedscore float32
		)

		err := getstmt.QueryRow(v.Id).Scan(&player.Played, &player.Wins, &player.Losses, &player.Score)
		if err != nil {
			if err == sql.ErrNoRows {
			} else {
				fmt.Println(err.Error())
			}
		}
		if numplayers <= 1 {
			interpolatedscore = 1
		} else {
			interpolatedscore = (1 - (float32(pos) / float32(numplayers-1)))
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
		_, err = updatestmt.Exec(v.Id, v.Name, player.Played, player.Wins, player.Losses, player.Score)
		if err != nil {
			fmt.Println(err.Error())
		}
		_, err = gamestmt.Exec(game.Id, v.Id, pos, numplayers)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
