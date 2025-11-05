class Bot {
  constructor() {
    this.id = 'bot';
    this.username = 'AI Bot';
    this.isBot = true;
  }

  // Minimax algorithm with alpha-beta pruning for competitive AI
  getBestMove(game) {
    const depth = 6; // Look ahead 6 moves
    const result = this.minimax(game.board, depth, -Infinity, Infinity, true);
    return result.column;
  }

  minimax(board, depth, alpha, beta, isMaximizing) {
    const score = this.evaluateBoard(board);
    
    // Terminal conditions
    if (depth === 0 || Math.abs(score) >= 1000 || this.isBoardFull(board)) {
      return { score, column: -1 };
    }

    const validMoves = this.getValidMoves(board);
    let bestColumn = validMoves[0];

    if (isMaximizing) {
      let maxScore = -Infinity;
      
      for (const col of validMoves) {
        const newBoard = this.makeMove(board, col, 2); // Bot is player 2
        const result = this.minimax(newBoard, depth - 1, alpha, beta, false);
        
        if (result.score > maxScore) {
          maxScore = result.score;
          bestColumn = col;
        }
        
        alpha = Math.max(alpha, result.score);
        if (beta <= alpha) break; // Alpha-beta pruning
      }
      
      return { score: maxScore, column: bestColumn };
    } else {
      let minScore = Infinity;
      
      for (const col of validMoves) {
        const newBoard = this.makeMove(board, col, 1); // Human is player 1
        const result = this.minimax(newBoard, depth - 1, alpha, beta, true);
        
        if (result.score < minScore) {
          minScore = result.score;
          bestColumn = col;
        }
        
        beta = Math.min(beta, result.score);
        if (beta <= alpha) break; // Alpha-beta pruning
      }
      
      return { score: minScore, column: bestColumn };
    }
  }

  evaluateBoard(board) {
    let score = 0;
    
    // Center column preference
    const centerCol = 3;
    for (let row = 0; row < 6; row++) {
      if (board[row][centerCol] === 2) score += 3;
      if (board[row][centerCol] === 1) score -= 3;
    }
    
    // Evaluate all possible 4-in-a-row windows
    score += this.evaluateWindows(board, 2) - this.evaluateWindows(board, 1);
    
    return score;
  }

  evaluateWindows(board, player) {
    let score = 0;
    
    // Horizontal windows
    for (let row = 0; row < 6; row++) {
      for (let col = 0; col < 4; col++) {
        const window = [board[row][col], board[row][col+1], board[row][col+2], board[row][col+3]];
        score += this.scoreWindow(window, player);
      }
    }
    
    // Vertical windows
    for (let col = 0; col < 7; col++) {
      for (let row = 0; row < 3; row++) {
        const window = [board[row][col], board[row+1][col], board[row+2][col], board[row+3][col]];
        score += this.scoreWindow(window, player);
      }
    }
    
    // Diagonal windows (positive slope)
    for (let row = 0; row < 3; row++) {
      for (let col = 0; col < 4; col++) {
        const window = [board[row][col], board[row+1][col+1], board[row+2][col+2], board[row+3][col+3]];
        score += this.scoreWindow(window, player);
      }
    }
    
    // Diagonal windows (negative slope)
    for (let row = 0; row < 3; row++) {
      for (let col = 3; col < 7; col++) {
        const window = [board[row][col], board[row+1][col-1], board[row+2][col-2], board[row+3][col-3]];
        score += this.scoreWindow(window, player);
      }
    }
    
    return score;
  }

  scoreWindow(window, player) {
    let score = 0;
    const opponent = player === 1 ? 2 : 1;
    
    const playerCount = window.filter(cell => cell === player).length;
    const opponentCount = window.filter(cell => cell === opponent).length;
    const emptyCount = window.filter(cell => cell === 0).length;
    
    if (playerCount === 4) {
      score += 1000; // Winning move
    } else if (playerCount === 3 && emptyCount === 1) {
      score += 100; // Three in a row with space
    } else if (playerCount === 2 && emptyCount === 2) {
      score += 10; // Two in a row with spaces
    }
    
    if (opponentCount === 3 && emptyCount === 1) {
      score -= 80; // Block opponent's winning move
    }
    
    return score;
  }

  getValidMoves(board) {
    const validMoves = [];
    for (let col = 0; col < 7; col++) {
      if (board[0][col] === 0) {
        validMoves.push(col);
      }
    }
    return validMoves;
  }

  makeMove(board, column, player) {
    const newBoard = board.map(row => [...row]);
    
    for (let row = 5; row >= 0; row--) {
      if (newBoard[row][column] === 0) {
        newBoard[row][column] = player;
        break;
      }
    }
    
    return newBoard;
  }

  isBoardFull(board) {
    return board[0].every(cell => cell !== 0);
  }

  // Quick move for immediate threats/opportunities
  getImmediateMove(game) {
    const board = game.board;
    
    // Check for winning move
    for (let col = 0; col < 7; col++) {
      if (board[0][col] === 0) {
        const testBoard = this.makeMove(board, col, 2);
        if (this.checkWinInBoard(testBoard, col, 2)) {
          return col;
        }
      }
    }
    
    // Check for blocking move
    for (let col = 0; col < 7; col++) {
      if (board[0][col] === 0) {
        const testBoard = this.makeMove(board, col, 1);
        if (this.checkWinInBoard(testBoard, col, 1)) {
          return col;
        }
      }
    }
    
    return null;
  }

  checkWinInBoard(board, col, player) {
    let row = -1;
    for (let r = 5; r >= 0; r--) {
      if (board[r][col] === player && (r === 5 || board[r+1][col] !== 0)) {
        row = r;
        break;
      }
    }
    
    if (row === -1) return false;
    
    const directions = [[0, 1], [1, 0], [1, 1], [1, -1]];
    
    for (const [dr, dc] of directions) {
      let count = 1;
      
      for (let i = 1; i < 4; i++) {
        const newRow = row + dr * i;
        const newCol = col + dc * i;
        if (newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 && 
            board[newRow][newCol] === player) {
          count++;
        } else break;
      }
      
      for (let i = 1; i < 4; i++) {
        const newRow = row - dr * i;
        const newCol = col - dc * i;
        if (newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 && 
            board[newRow][newCol] === player) {
          count++;
        } else break;
      }
      
      if (count >= 4) return true;
    }
    return false;
  }
}

module.exports = Bot;