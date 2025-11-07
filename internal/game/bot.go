package game

import (
	"emitrr-4-in-a-row/internal/models"
	"math"
)

type Bot struct {
	ID       string
	Username string
	IsBot    bool
}

func NewBot() *Bot {
	return &Bot{
		ID:       "bot",
		Username: "AI Bot",
		IsBot:    true,
	}
}

func (b *Bot) GetBestMove(game *models.Game) int {
	// Adaptive depth based on game state
	depth := b.getOptimalDepth(game.Board)
	result := b.minimax(game.Board, depth, math.Inf(-1), math.Inf(1), true)
	return result.Column
}

func (b *Bot) getOptimalDepth(board [][]int) int {
	emptySpaces := 0
	for i := 0; i < 6; i++ {
		for j := 0; j < 7; j++ {
			if board[i][j] == 0 {
				emptySpaces++
			}
		}
	}
	
	// Deeper search when fewer pieces on board
	if emptySpaces > 35 {
		return 7 // Early game - deeper search
	} else if emptySpaces > 20 {
		return 8 // Mid game - deepest search
	} else {
		return 9 // End game - maximum depth
	}
}

type MinimaxResult struct {
	Score  float64
	Column int
}

func (b *Bot) minimax(board [][]int, depth int, alpha, beta float64, isMaximizing bool) MinimaxResult {
	score := b.evaluateBoard(board)

	if depth == 0 || math.Abs(score) >= 1000 || b.isBoardFull(board) {
		return MinimaxResult{Score: score, Column: -1}
	}

	validMoves := b.getValidMoves(board)
	bestColumn := validMoves[0]

	if isMaximizing {
		maxScore := math.Inf(-1)

		for _, col := range validMoves {
			newBoard := b.makeMove(board, col, 2)
			result := b.minimax(newBoard, depth-1, alpha, beta, false)

			if result.Score > maxScore {
				maxScore = result.Score
				bestColumn = col
			}

			alpha = math.Max(alpha, result.Score)
			if beta <= alpha {
				break
			}
		}

		return MinimaxResult{Score: maxScore, Column: bestColumn}
	} else {
		minScore := math.Inf(1)

		for _, col := range validMoves {
			newBoard := b.makeMove(board, col, 1)
			result := b.minimax(newBoard, depth-1, alpha, beta, true)

			if result.Score < minScore {
				minScore = result.Score
				bestColumn = col
			}

			beta = math.Min(beta, result.Score)
			if beta <= alpha {
				break
			}
		}

		return MinimaxResult{Score: minScore, Column: bestColumn}
	}
}

func (b *Bot) evaluateBoard(board [][]int) float64 {
	score := 0.0

	// Strong center column preference (most important)
	centerCol := 3
	for row := 0; row < 6; row++ {
		if board[row][centerCol] == 2 {
			score += 6 // Doubled importance
		}
		if board[row][centerCol] == 1 {
			score -= 6
		}
	}

	// Adjacent center columns also valuable
	for row := 0; row < 6; row++ {
		if board[row][2] == 2 || board[row][4] == 2 {
			score += 4
		}
		if board[row][2] == 1 || board[row][4] == 1 {
			score -= 4
		}
	}

	// Evaluate all windows with enhanced scoring
	score += b.evaluateWindows(board, 2) - b.evaluateWindows(board, 1)
	
	// Penalize edge columns heavily
	for row := 0; row < 6; row++ {
		if board[row][0] == 2 || board[row][6] == 2 {
			score -= 2
		}
	}

	return score
}

func (b *Bot) evaluateWindows(board [][]int, player int) float64 {
	score := 0.0

	// Horizontal windows
	for row := 0; row < 6; row++ {
		for col := 0; col < 4; col++ {
			window := []int{board[row][col], board[row][col+1], board[row][col+2], board[row][col+3]}
			score += b.scoreWindow(window, player)
		}
	}

	// Vertical windows
	for col := 0; col < 7; col++ {
		for row := 0; row < 3; row++ {
			window := []int{board[row][col], board[row+1][col], board[row+2][col], board[row+3][col]}
			score += b.scoreWindow(window, player)
		}
	}

	// Diagonal windows (positive slope)
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			window := []int{board[row][col], board[row+1][col+1], board[row+2][col+2], board[row+3][col+3]}
			score += b.scoreWindow(window, player)
		}
	}

	// Diagonal windows (negative slope)
	for row := 0; row < 3; row++ {
		for col := 3; col < 7; col++ {
			window := []int{board[row][col], board[row+1][col-1], board[row+2][col-2], board[row+3][col-3]}
			score += b.scoreWindow(window, player)
		}
	}

	return score
}

func (b *Bot) scoreWindow(window []int, player int) float64 {
	score := 0.0
	opponent := 1
	if player == 1 {
		opponent = 2
	}

	playerCount := 0
	opponentCount := 0
	emptyCount := 0

	for _, cell := range window {
		if cell == player {
			playerCount++
		} else if cell == opponent {
			opponentCount++
		} else {
			emptyCount++
		}
	}

	// Winning positions
	if playerCount == 4 {
		score += 10000 // Massive win bonus
	} else if playerCount == 3 && emptyCount == 1 {
		score += 500 // Strong threat
	} else if playerCount == 2 && emptyCount == 2 {
		score += 50 // Good position
	} else if playerCount == 1 && emptyCount == 3 {
		score += 5 // Potential
	}

	// Defensive positions - CRITICAL
	if opponentCount == 4 {
		score -= 10000 // Prevent loss
	} else if opponentCount == 3 && emptyCount == 1 {
		score -= 1000 // MUST block
	} else if opponentCount == 2 && emptyCount == 2 {
		score -= 100 // Block potential threat
	}

	return score
}

func (b *Bot) getValidMoves(board [][]int) []int {
	var validMoves []int
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			validMoves = append(validMoves, col)
		}
	}
	return validMoves
}

func (b *Bot) makeMove(board [][]int, column, player int) [][]int {
	newBoard := make([][]int, 6)
	for i := range board {
		newBoard[i] = make([]int, 7)
		copy(newBoard[i], board[i])
	}

	for row := 5; row >= 0; row-- {
		if newBoard[row][column] == 0 {
			newBoard[row][column] = player
			break
		}
	}

	return newBoard
}

func (b *Bot) isBoardFull(board [][]int) bool {
	for _, cell := range board[0] {
		if cell == 0 {
			return false
		}
	}
	return true
}

func (b *Bot) GetImmediateMove(game *models.Game) *int {
	board := game.Board

	// 1. Check for immediate winning move (HIGHEST PRIORITY)
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, 2)
			if b.checkWinInBoard(testBoard, col, 2) {
				return &col
			}
		}
	}

	// 2. Check for blocking opponent's winning move (CRITICAL)
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, 1)
			if b.checkWinInBoard(testBoard, col, 1) {
				return &col
			}
		}
	}

	// 3. Check for creating double threats (ADVANCED)
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, 2)
			threats := b.countThreats(testBoard, 2)
			if threats >= 2 {
				return &col // Create multiple winning opportunities
			}
		}
	}

	// 4. Prefer center columns for strategic advantage
	centerCols := []int{3, 2, 4, 1, 5, 0, 6}
	for _, col := range centerCols {
		if board[0][col] == 0 {
			return &col
		}
	}

	return nil
}

func (b *Bot) countThreats(board [][]int, player int) int {
	threats := 0
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, player)
			if b.checkWinInBoard(testBoard, col, player) {
				threats++
			}
		}
	}
	return threats
}

func (b *Bot) checkWinInBoard(board [][]int, col, player int) bool {
	row := -1
	for r := 5; r >= 0; r-- {
		if board[r][col] == player && (r == 5 || board[r+1][col] != 0) {
			row = r
			break
		}
	}

	if row == -1 {
		return false
	}
// det
	directions := [][]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, dir := range directions {
		count := 1
		dr, dc := dir[0], dir[1]

		for i := 1; i < 4; i++ {
			newRow, newCol := row+dr*i, col+dc*i
			if newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 &&
				board[newRow][newCol] == player {
				count++
			} else {
				break
			}
		}

		for i := 1; i < 4; i++ {
			newRow, newCol := row-dr*i, col-dc*i
			if newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 &&
				board[newRow][newCol] == player {
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