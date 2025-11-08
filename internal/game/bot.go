package game

import (
	"emitrr-4-in-a-row/internal/models"
	"math"
	"math/rand"
	"sort"
	"time"
)

type Bot struct {
	ID       string
	Username string
	IsBot    bool
	transTable map[uint64]TransEntry
}

type TransEntry struct {
	Score float64
	Depth int
	Flag  int // 0=exact, 1=lower, 2=upper
}

func NewBot() *Bot {
	return &Bot{
		ID:         "bot",
		Username:   "AI Bot",
		IsBot:      true,
		transTable: make(map[uint64]TransEntry),
	}
}

func (b *Bot) GetBestMove(game *models.Game) int {
	validMoves := b.getValidMoves(game.Board)
	if len(validMoves) == 0 {
		return -1
	}

	// Immediate tactical moves (win/block)
	if move := b.GetImmediateMove(game); move != nil {
		return *move
	}

	// Advanced threat analysis
	if move := b.analyzeThreats(game.Board); move != -1 {
		return move
	}

	// Iterative deepening with time limit
	depth := b.getOptimalDepth(game.Board)
	result := b.iterativeDeepening(game.Board, depth, 2*time.Second)
	
	for _, col := range validMoves {
		if result.Column == col {
			return result.Column
		}
	}
	
	return b.selectStrategicMove(game.Board, validMoves)
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
	
	if emptySpaces > 35 {
		return 8
	} else if emptySpaces > 20 {
		return 10
	} else if emptySpaces > 10 {
		return 12
	} else {
		return 15 // Endgame - solve completely
	}
}

func (b *Bot) iterativeDeepening(board [][]int, maxDepth int, timeLimit time.Duration) MinimaxResult {
	start := time.Now()
	var bestResult MinimaxResult
	
	for depth := 1; depth <= maxDepth; depth++ {
		if time.Since(start) > timeLimit {
			break
		}
		
		result := b.minimax(board, depth, math.Inf(-1), math.Inf(1), true)
		bestResult = result
		
		// If we found a winning move, return immediately
		if result.Score >= 10000 {
			break
		}
	}
	
	return bestResult
}

func (b *Bot) analyzeThreats(board [][]int) int {
	// Look for fork opportunities (multiple threats)
	bestScore := -1.0
	bestMove := -1
	
	for col := 0; col < 7; col++ {
		if board[0][col] != 0 {
			continue
		}
		
		testBoard := b.makeMove(board, col, 2)
		threats := b.countAdvancedThreats(testBoard, 2)
		defensiveValue := b.evaluateDefensivePosition(testBoard, col)
		
		score := float64(threats)*100 + defensiveValue
		
		if score > bestScore {
			bestScore = score
			bestMove = col
		}
	}
	
	if bestScore > 150 { // Threshold for strong tactical move
		return bestMove
	}
	return -1
}

func (b *Bot) countAdvancedThreats(board [][]int, player int) int {
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

func (b *Bot) evaluateDefensivePosition(board [][]int, col int) float64 {
	score := 0.0
	
	// Check if this move blocks opponent threats
	testBoard := b.makeMove(board, col, 1) // Simulate opponent move
	if b.checkWinInBoard(testBoard, col, 1) {
		score += 200 // High value for blocking
	}
	
	// Check for trap setups (moves that create unavoidable threats)
	for nextCol := 0; nextCol < 7; nextCol++ {
		if board[0][nextCol] == 0 && nextCol != col {
			nextBoard := b.makeMove(board, nextCol, 2)
			if b.countAdvancedThreats(nextBoard, 2) >= 2 {
				score += 50
			}
		}
	}
	
	return score
}

func (b *Bot) selectStrategicMove(board [][]int, validMoves []int) int {
	// Prioritize center columns with some randomness
	centerPreference := []int{3, 2, 4, 1, 5, 0, 6}
	
	for _, col := range centerPreference {
		for _, valid := range validMoves {
			if col == valid {
				// Add slight randomness to avoid predictability
				if rand.Float64() < 0.8 {
					return col
				}
			}
		}
	}
	
	return validMoves[rand.Intn(len(validMoves))]
}

type MinimaxResult struct {
	Score  float64
	Column int
}

func (b *Bot) minimax(board [][]int, depth int, alpha, beta float64, isMaximizing bool) MinimaxResult {
	// Transposition table lookup
	hash := b.hashBoard(board)
	if entry, exists := b.transTable[hash]; exists && entry.Depth >= depth {
		if entry.Flag == 0 || (entry.Flag == 1 && entry.Score >= beta) || (entry.Flag == 2 && entry.Score <= alpha) {
			return MinimaxResult{Score: entry.Score, Column: -1}
		}
	}

	score := b.evaluateBoard(board)
	if depth == 0 || math.Abs(score) >= 10000 || b.isBoardFull(board) {
		return MinimaxResult{Score: score, Column: -1}
	}

	validMoves := b.getValidMoves(board)
	if len(validMoves) == 0 {
		return MinimaxResult{Score: score, Column: -1}
	}

	// Move ordering for better pruning
	orderedMoves := b.orderMoves(board, validMoves, isMaximizing)
	bestColumn := orderedMoves[0]
	originalAlpha := alpha

	if isMaximizing {
		maxScore := math.Inf(-1)
		for _, col := range orderedMoves {
			newBoard := b.makeMove(board, col, 2)
			result := b.minimax(newBoard, depth-1, alpha, beta, false)

			if result.Score > maxScore {
				maxScore = result.Score
				bestColumn = col
			}

			alpha = math.Max(alpha, result.Score)
			if beta <= alpha {
				break // Beta cutoff
			}
		}

		// Store in transposition table
		flag := 0
		if maxScore <= originalAlpha {
			flag = 2 // Upper bound
		} else if maxScore >= beta {
			flag = 1 // Lower bound
		}
		b.transTable[hash] = TransEntry{Score: maxScore, Depth: depth, Flag: flag}

		return MinimaxResult{Score: maxScore, Column: bestColumn}
	} else {
		minScore := math.Inf(1)
		for _, col := range orderedMoves {
			newBoard := b.makeMove(board, col, 1)
			result := b.minimax(newBoard, depth-1, alpha, beta, true)

			if result.Score < minScore {
				minScore = result.Score
				bestColumn = col
			}

			beta = math.Min(beta, result.Score)
			if beta <= alpha {
				break // Alpha cutoff
			}
		}

		// Store in transposition table
		flag := 0
		if minScore <= originalAlpha {
			flag = 2
		} else if minScore >= beta {
			flag = 1
		}
		b.transTable[hash] = TransEntry{Score: minScore, Depth: depth, Flag: flag}

		return MinimaxResult{Score: minScore, Column: bestColumn}
	}
}

func (b *Bot) hashBoard(board [][]int) uint64 {
	var hash uint64 = 0
	for i := 0; i < 6; i++ {
		for j := 0; j < 7; j++ {
			hash = hash*3 + uint64(board[i][j])
		}
	}
	return hash
}

func (b *Bot) orderMoves(board [][]int, moves []int, isMaximizing bool) []int {
	type moveScore struct {
		col   int
		score float64
	}
	
	scores := make([]moveScore, len(moves))
	for i, col := range moves {
		player := 2
		if !isMaximizing {
			player = 1
		}
		testBoard := b.makeMove(board, col, player)
		score := b.evaluateBoard(testBoard)
		
		// Prioritize center columns
		if col == 3 {
			score += 10
		} else if col == 2 || col == 4 {
			score += 5
		}
		
		scores[i] = moveScore{col: col, score: score}
	}
	
	// Sort by score (descending for maximizing, ascending for minimizing)
	sort.Slice(scores, func(i, j int) bool {
		if isMaximizing {
			return scores[i].score > scores[j].score
		}
		return scores[i].score < scores[j].score
	})
	
	orderedMoves := make([]int, len(moves))
	for i, ms := range scores {
		orderedMoves[i] = ms.col
	}
	return orderedMoves
}

func (b *Bot) evaluateBoard(board [][]int) float64 {
	score := 0.0

	// Check for immediate wins/losses
	if winner := b.checkBoardWinner(board); winner != 0 {
		if winner == 2 {
			return 100000
		}
		return -100000
	}

	// Advanced positional evaluation
	score += b.evaluatePositionalAdvantage(board)
	score += b.evaluateConnections(board, 2) - b.evaluateConnections(board, 1)
	score += b.evaluateThreatPotential(board)
	score += b.evaluateControlledColumns(board)

	return score
}

func (b *Bot) evaluatePositionalAdvantage(board [][]int) float64 {
	score := 0.0
	
	// Center control is crucial
	for row := 0; row < 6; row++ {
		if board[row][3] == 2 {
			score += 8.0 * float64(6-row) // Higher pieces worth more
		} else if board[row][3] == 1 {
			score -= 8.0 * float64(6-row)
		}
	}
	
	// Adjacent center columns
	for row := 0; row < 6; row++ {
		for _, col := range []int{2, 4} {
			if board[row][col] == 2 {
				score += 5.0 * float64(6-row)
			} else if board[row][col] == 1 {
				score -= 5.0 * float64(6-row)
			}
		}
	}
	
	// Penalize edge columns
	for row := 0; row < 6; row++ {
		for _, col := range []int{0, 6} {
			if board[row][col] == 2 {
				score -= 3.0
			} else if board[row][col] == 1 {
				score += 1.0 // Slightly good to force opponent to edges
			}
		}
	}
	
	return score
}

func (b *Bot) evaluateConnections(board [][]int, player int) float64 {
	score := 0.0
	
	// Horizontal connections
	for row := 0; row < 6; row++ {
		for col := 0; col < 4; col++ {
			window := []int{board[row][col], board[row][col+1], board[row][col+2], board[row][col+3]}
			score += b.scoreAdvancedWindow(window, player, "horizontal")
		}
	}
	
	// Vertical connections
	for col := 0; col < 7; col++ {
		for row := 0; row < 3; row++ {
			window := []int{board[row][col], board[row+1][col], board[row+2][col], board[row+3][col]}
			score += b.scoreAdvancedWindow(window, player, "vertical")
		}
	}
	
	// Diagonal connections
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			window1 := []int{board[row][col], board[row+1][col+1], board[row+2][col+2], board[row+3][col+3]}
			score += b.scoreAdvancedWindow(window1, player, "diagonal")
		}
		for col := 3; col < 7; col++ {
			window2 := []int{board[row][col], board[row+1][col-1], board[row+2][col-2], board[row+3][col-3]}
			score += b.scoreAdvancedWindow(window2, player, "diagonal")
		}
	}
	
	return score
}

func (b *Bot) scoreAdvancedWindow(window []int, player int, direction string) float64 {
	score := 0.0
	opponent := 3 - player
	
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
	
	// Can't form 4 in a row if opponent has pieces
	if opponentCount > 0 {
		return 0
	}
	
	// Scoring based on potential
	switch playerCount {
	case 4:
		score = 10000
	case 3:
		score = 500
		if direction == "vertical" {
			score *= 1.5 // Vertical threats are stronger
		}
	case 2:
		score = 50
		if direction == "horizontal" && emptyCount == 2 {
			score *= 1.2 // Open-ended horizontals are valuable
		}
	case 1:
		score = 5
	}
	
	return score
}

func (b *Bot) evaluateThreatPotential(board [][]int) float64 {
	score := 0.0
	
	// Count potential threats for both players
	botThreats := b.countPotentialThreats(board, 2)
	oppThreats := b.countPotentialThreats(board, 1)
	
	score += float64(botThreats)*20 - float64(oppThreats)*25
	
	return score
}

func (b *Bot) countPotentialThreats(board [][]int, player int) int {
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

func (b *Bot) evaluateControlledColumns(board [][]int) float64 {
	score := 0.0
	
	for col := 0; col < 7; col++ {
		botControl := 0
		oppControl := 0
		
		for row := 5; row >= 0; row-- {
			if board[row][col] == 2 {
				botControl++
			} else if board[row][col] == 1 {
				oppControl++
			} else {
				break // Empty space, stop counting
			}
		}
		
		if botControl > oppControl {
			score += float64(botControl-oppControl) * 3
		} else if oppControl > botControl {
			score -= float64(oppControl-botControl) * 3
		}
	}
	
	return score
}

func (b *Bot) checkBoardWinner(board [][]int) int {
	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			if board[row][col] != 0 {
				if b.checkWinFromPosition(board, row, col, board[row][col]) {
					return board[row][col]
				}
			}
		}
	}
	return 0
}

func (b *Bot) checkWinFromPosition(board [][]int, row, col, player int) bool {
	directions := [][]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	
	for _, dir := range directions {
		count := 1
		dr, dc := dir[0], dir[1]
		
		// Check positive direction
		for i := 1; i < 4; i++ {
			nr, nc := row+dr*i, col+dc*i
			if nr >= 0 && nr < 6 && nc >= 0 && nc < 7 && board[nr][nc] == player {
				count++
			} else {
				break
			}
		}
		
		// Check negative direction
		for i := 1; i < 4; i++ {
			nr, nc := row-dr*i, col-dc*i
			if nr >= 0 && nr < 6 && nc >= 0 && nc < 7 && board[nr][nc] == player {
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

	// 1. Immediate win
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, 2)
			if b.checkWinInBoard(testBoard, col, 2) {
				return &col
			}
		}
	}

	// 2. Block opponent win
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, 1)
			if b.checkWinInBoard(testBoard, col, 1) {
				return &col
			}
		}
	}

	// 3. Create multiple threats
	for col := 0; col < 7; col++ {
		if board[0][col] == 0 {
			testBoard := b.makeMove(board, col, 2)
			threats := b.countThreats(testBoard, 2)
			if threats >= 2 {
				return &col
			}
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

	return b.checkWinFromPosition(board, row, col, player)
}