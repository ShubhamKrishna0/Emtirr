const Game = require('../src/models/Game');
const Bot = require('../src/game/Bot');

describe('Game Logic Tests', () => {
  let game;
  let player1;
  let player2;

  beforeEach(() => {
    player1 = { id: 'player1', username: 'Alice' };
    player2 = { id: 'player2', username: 'Bob' };
    game = new Game('test-game', player1, player2);
  });

  describe('Game Initialization', () => {
    test('should create game with correct initial state', () => {
      expect(game.id).toBe('test-game');
      expect(game.player1).toBe(player1);
      expect(game.player2).toBe(player2);
      expect(game.currentPlayer).toBe(1);
      expect(game.status).toBe('waiting');
      expect(game.board).toHaveLength(6);
      expect(game.board[0]).toHaveLength(7);
    });

    test('should start game when second player joins', () => {
      const singlePlayerGame = new Game('test', player1);
      expect(singlePlayerGame.status).toBe('waiting');
      
      const result = singlePlayerGame.addPlayer(player2);
      expect(result).toBe(true);
      expect(singlePlayerGame.status).toBe('playing');
    });
  });

  describe('Move Validation', () => {
    beforeEach(() => {
      game.status = 'playing';
    });

    test('should accept valid moves', () => {
      const result = game.makeMove(3, 'player1');
      expect(result.success).toBe(true);
      expect(result.row).toBe(5);
      expect(game.board[5][3]).toBe(1);
    });

    test('should reject moves from wrong player', () => {
      const result = game.makeMove(3, 'player2');
      expect(result.success).toBe(false);
      expect(result.error).toBe('Not your turn');
    });

    test('should reject invalid column numbers', () => {
      const result1 = game.makeMove(-1, 'player1');
      const result2 = game.makeMove(7, 'player1');
      
      expect(result1.success).toBe(false);
      expect(result2.success).toBe(false);
      expect(result1.error).toBe('Invalid column');
      expect(result2.error).toBe('Invalid column');
    });

    test('should reject moves in full columns', () => {
      // Fill column 0
      for (let i = 0; i < 6; i++) {
        game.makeMove(0, i % 2 === 0 ? 'player1' : 'player2');
      }
      
      const result = game.makeMove(0, 'player1');
      expect(result.success).toBe(false);
      expect(result.error).toBe('Column is full');
    });
  });

  describe('Win Detection', () => {
    beforeEach(() => {
      game.status = 'playing';
    });

    test('should detect horizontal win', () => {
      // Create horizontal win for player 1
      game.board[5] = [1, 1, 1, 0, 0, 0, 0];
      const result = game.makeMove(3, 'player1');
      
      expect(result.success).toBe(true);
      expect(result.gameOver).toBe(true);
      expect(result.winner).toBe(1);
      expect(game.status).toBe('finished');
    });

    test('should detect vertical win', () => {
      // Create vertical win setup
      game.board[5][0] = 1;
      game.board[4][0] = 1;
      game.board[3][0] = 1;
      
      const result = game.makeMove(0, 'player1');
      
      expect(result.success).toBe(true);
      expect(result.gameOver).toBe(true);
      expect(result.winner).toBe(1);
    });

    test('should detect diagonal win', () => {
      // Create diagonal win setup
      game.board[5][0] = 1;
      game.board[4][1] = 1;
      game.board[3][2] = 1;
      
      const result = game.makeMove(3, 'player1');
      
      expect(result.success).toBe(true);
      expect(result.gameOver).toBe(true);
      expect(result.winner).toBe(1);
    });
  });
});

describe('Bot AI Tests', () => {
  let bot;
  let game;

  beforeEach(() => {
    bot = new Bot();
    const player1 = { id: 'human', username: 'Human' };
    game = new Game('test-game', player1, bot);
    game.status = 'playing';
  });

  describe('Bot Decision Making', () => {
    test('should make winning move when available', () => {
      // Set up winning opportunity for bot (player 2)
      game.board[5] = [2, 2, 2, 0, 0, 0, 0];
      
      const move = bot.getImmediateMove(game);
      expect(move).toBe(3);
    });

    test('should block opponent winning move', () => {
      // Set up winning opportunity for human (player 1)
      game.board[5] = [1, 1, 1, 0, 0, 0, 0];
      
      const move = bot.getImmediateMove(game);
      expect(move).toBe(3);
    });

    test('should prefer center columns', () => {
      // Empty board - bot should prefer center
      const move = bot.getBestMove(game);
      expect([2, 3, 4]).toContain(move);
    });
  });
});

// Mock test functions for demonstration
function describe(name, fn) {
  console.log(`\n📝 ${name}`);
  fn();
}

function test(name, fn) {
  try {
    fn();
    console.log(`  ✅ ${name}`);
  } catch (error) {
    console.log(`  ❌ ${name}: ${error.message}`);
  }
}

function beforeEach(fn) {
  // Setup function
}

function expect(actual) {
  return {
    toBe: (expected) => {
      if (actual !== expected) {
        throw new Error(`Expected ${expected}, got ${actual}`);
      }
    },
    toHaveLength: (expected) => {
      if (actual.length !== expected) {
        throw new Error(`Expected length ${expected}, got ${actual.length}`);
      }
    },
    toContain: (expected) => {
      if (!actual.includes(expected)) {
        throw new Error(`Expected ${actual} to contain ${expected}`);
      }
    }
  };
}

module.exports = { describe, test, beforeEach, expect };