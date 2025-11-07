package models

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"isBot"`
}

type Move struct {
	Player    int       `json:"player"`
	Row       int       `json:"row"`
	Column    int       `json:"column"`
	Timestamp time.Time `json:"timestamp"`
}

type Game struct {
	ID            string    `json:"id"`
	Player1       *Player   `json:"player1"`
	Player2       *Player   `json:"player2"`
	Board         [][]int   `json:"board"`
	CurrentPlayer int       `json:"currentPlayer"`
	Status        string    `json:"status"` // waiting, playing, finished
	Winner        *int      `json:"winner"`
	CreatedAt     time.Time `json:"createdAt"`
	LastMoveAt    time.Time `json:"lastMoveAt"`
	Moves         []Move    `json:"moves"`
	IsBot         bool      `json:"isBot"`
}

func NewGame(player1 *Player, player2 *Player) *Game {
	board := make([][]int, 6)
	for i := range board {
		board[i] = make([]int, 7)
	}

	return &Game{
		ID:            uuid.New().String(),
		Player1:       player1,
		Player2:       player2,
		Board:         board,
		CurrentPlayer: 1,
		Status:        "waiting",
		CreatedAt:     time.Now(),
		LastMoveAt:    time.Now(),
		Moves:         make([]Move, 0),
		IsBot:         player2 != nil && player2.IsBot,
	}
}

func (g *Game) AddPlayer(player *Player) bool {
	if g.Player2 == nil {
		g.Player2 = player
		g.Status = "playing"
		return true
	}
	return false
}

func (g *Game) MakeMove(column int, playerNumber int) (int, bool, *int, error) {
	if g.Status != "playing" {
		return -1, false, nil, &GameError{"Game not active"}
	}

	if playerNumber != g.CurrentPlayer {
		return -1, false, nil, &GameError{"Not your turn"}
	}

	if column < 0 || column > 6 {
		return -1, false, nil, &GameError{"Invalid column"}
	}

	// Find lowest available row
	row := -1
	for r := 5; r >= 0; r-- {
		if g.Board[r][column] == 0 {
			row = r
			break
		}
	}

	if row == -1 {
		return -1, false, nil, &GameError{"Column is full"}
	}

	// Make the move
	g.Board[row][column] = playerNumber
	g.Moves = append(g.Moves, Move{
		Player:    playerNumber,
		Row:       row,
		Column:    column,
		Timestamp: time.Now(),
	})
	g.LastMoveAt = time.Now()

	// Check for win
	if g.CheckWin(row, column, playerNumber) {
		g.Status = "finished"
		g.Winner = &playerNumber
		return row, true, &playerNumber, nil
	}

	// Check for draw
	if g.IsBoardFull() {
		g.Status = "finished"
		return row, true, nil, nil
	}

	// Switch turns
	if g.CurrentPlayer == 1 {
		g.CurrentPlayer = 2
	} else {
		g.CurrentPlayer = 1
	}

	return row, false, nil, nil
}

func (g *Game) CheckWin(row, col, player int) bool {
	directions := [][]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, dir := range directions {
		count := 1
		dr, dc := dir[0], dir[1]

		// Check positive direction
		for i := 1; i < 4; i++ {
			newRow, newCol := row+dr*i, col+dc*i
			if newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 &&
				g.Board[newRow][newCol] == player {
				count++
			} else {
				break
			}
		}

		// Check negative direction
		for i := 1; i < 4; i++ {
			newRow, newCol := row-dr*i, col-dc*i
			if newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 &&
				g.Board[newRow][newCol] == player {
				count++
			} else {
				break
			}
		}

		if count >= 4 {
			return true
		}
	}
	return false
}

func (g *Game) IsBoardFull() bool {
	for _, cell := range g.Board[0] {
		if cell == 0 {
			return false
		}
	}
	return true
}

func (g *Game) GetPlayerNumber(playerID string) int {
	if g.Player1.ID == playerID {
		return 1
	}
	if g.Player2 != nil && g.Player2.ID == playerID {
		return 2
	}
	return 0
}

func (g *Game) GetPlayerNumberByUsername(username string) int {
	if g.Player1.Username == username {
		return 1
	}
	if g.Player2 != nil && g.Player2.Username == username {
		return 2
	}
	return 0
}

func (g *Game) GetDuration() int {
	endTime := g.LastMoveAt
	if g.Status == "finished" {
		endTime = g.LastMoveAt
	} else {
		endTime = time.Now()
	}
	return int(endTime.Sub(g.CreatedAt).Seconds())
}

type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}