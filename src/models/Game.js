class Game {
  constructor(id, player1, player2 = null) {
    this.id = id;
    this.player1 = player1;
    this.player2 = player2;
    this.board = Array(6).fill().map(() => Array(7).fill(0));
    this.currentPlayer = 1;
    this.status = 'waiting'; // waiting, playing, finished
    this.winner = null;
    this.createdAt = new Date();
    this.lastMoveAt = new Date();
    this.moves = [];
    this.isBot = false;
  }

  addPlayer(player) {
    if (!this.player2) {
      this.player2 = player;
      this.status = 'playing';
      return true;
    }
    return false;
  }

  makeMove(column, playerId) {
    if (this.status !== 'playing') return { success: false, error: 'Game not active' };
    
    const playerNumber = this.getPlayerNumber(playerId);
    if (playerNumber !== this.currentPlayer) {
      return { success: false, error: 'Not your turn' };
    }

    if (column < 0 || column > 6) {
      return { success: false, error: 'Invalid column' };
    }

    // Find the lowest available row
    let row = -1;
    for (let r = 5; r >= 0; r--) {
      if (this.board[r][column] === 0) {
        row = r;
        break;
      }
    }

    if (row === -1) {
      return { success: false, error: 'Column is full' };
    }

    // Make the move
    this.board[row][column] = playerNumber;
    this.moves.push({ player: playerNumber, row, column, timestamp: new Date() });
    this.lastMoveAt = new Date();

    // Check for win
    if (this.checkWin(row, column, playerNumber)) {
      this.status = 'finished';
      this.winner = playerNumber;
      return { success: true, row, gameOver: true, winner: playerNumber };
    }

    // Check for draw
    if (this.isBoardFull()) {
      this.status = 'finished';
      return { success: true, row, gameOver: true, winner: null };
    }

    // Switch turns
    this.currentPlayer = this.currentPlayer === 1 ? 2 : 1;
    return { success: true, row, gameOver: false };
  }

  checkWin(row, col, player) {
    const directions = [
      [0, 1],   // horizontal
      [1, 0],   // vertical
      [1, 1],   // diagonal /
      [1, -1]   // diagonal \
    ];

    for (const [dr, dc] of directions) {
      let count = 1;
      
      // Check positive direction
      for (let i = 1; i < 4; i++) {
        const newRow = row + dr * i;
        const newCol = col + dc * i;
        if (newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 && 
            this.board[newRow][newCol] === player) {
          count++;
        } else break;
      }
      
      // Check negative direction
      for (let i = 1; i < 4; i++) {
        const newRow = row - dr * i;
        const newCol = col - dc * i;
        if (newRow >= 0 && newRow < 6 && newCol >= 0 && newCol < 7 && 
            this.board[newRow][newCol] === player) {
          count++;
        } else break;
      }
      
      if (count >= 4) return true;
    }
    return false;
  }

  isBoardFull() {
    return this.board[0].every(cell => cell !== 0);
  }

  getPlayerNumber(playerId) {
    if (this.player1.id === playerId) return 1;
    if (this.player2 && this.player2.id === playerId) return 2;
    return null;
  }

  getGameState() {
    return {
      id: this.id,
      board: this.board,
      currentPlayer: this.currentPlayer,
      status: this.status,
      winner: this.winner,
      player1: { id: this.player1.id, username: this.player1.username },
      player2: this.player2 ? { id: this.player2.id, username: this.player2.username } : null,
      isBot: this.isBot,
      moves: this.moves.length
    };
  }

  getDuration() {
    const endTime = this.status === 'finished' ? this.lastMoveAt : new Date();
    return Math.floor((endTime - this.createdAt) / 1000);
  }
}

module.exports = Game;