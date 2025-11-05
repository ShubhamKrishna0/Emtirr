const { v4: uuidv4 } = require('uuid');
const Game = require('../models/Game');
const Bot = require('./Bot');

class GameManager {
  constructor(io, dbService, analyticsService) {
    this.io = io;
    this.dbService = dbService;
    this.analyticsService = analyticsService;
    this.games = new Map();
    this.waitingPlayers = new Map();
    this.playerSockets = new Map();
    this.disconnectedPlayers = new Map();
    this.bot = new Bot();
    
    // Cleanup disconnected players every 30 seconds
    setInterval(() => this.cleanupDisconnectedPlayers(), 30000);
  }

  handlePlayerJoin(socket, data) {
    const { username } = data;
    const validation = require('../middleware/security').validateUsername(username);
    
    if (!validation.valid) {
      socket.emit('error', { message: validation.error });
      return;
    }

    const player = {
      id: socket.id,
      username: validation.username
    };
    
    player.socket = socket;

    this.playerSockets.set(socket.id, {
      id: socket.id,
      username: player.username
    });

    // Check if player was disconnected from an active game
    const reconnectGame = this.findReconnectableGame(username);
    if (reconnectGame) {
      this.handlePlayerRejoin(socket, { gameId: reconnectGame.id, username });
      return;
    }

    // Try to match with waiting player
    const waitingPlayer = this.findWaitingPlayer(username);
    if (waitingPlayer) {
      this.createGame(waitingPlayer, player);
    } else {
      this.addToWaitingQueue(player);
    }
  }

  findWaitingPlayer(currentUsername) {
    for (const [socketId, player] of this.waitingPlayers) {
      if (player.username !== currentUsername) {
        this.waitingPlayers.delete(socketId);
        return player;
      }
    }
    return null;
  }

  addToWaitingQueue(player) {
    this.waitingPlayers.set(player.id, player);
    player.socket.emit('waiting_for_opponent');

    // Start bot game after 10 seconds if no opponent found
    setTimeout(() => {
      if (this.waitingPlayers.has(player.id)) {
        console.log(`Starting bot game for ${player.username}`);
        this.waitingPlayers.delete(player.id);
        this.createBotGame(player);
      }
    }, 10000);
  }

  createGame(player1, player2) {
    const gameId = uuidv4();
    const game = new Game(gameId, player1, player2);
    this.games.set(gameId, game);

    // Join both players to the game room
    player1.socket.join(gameId);
    player2.socket.join(gameId);

    // Notify players
    this.io.to(gameId).emit('game_started', {
      gameId,
      gameState: game.getGameState(),
      yourPlayer: 1
    });

    player1.socket.emit('your_turn', { player: 1 });
    player2.socket.emit('your_turn', { player: 2 });

    // Analytics
    this.analyticsService.trackEvent('game_started', {
      gameId,
      player1: player1.username,
      player2: player2.username,
      gameType: 'pvp'
    });
  }

  createBotGame(player) {
    console.log(`Creating bot game for ${player.username}`);
    const gameId = uuidv4();
    const game = new Game(gameId, player, this.bot);
    game.isBot = true;
    game.status = 'playing';
    this.games.set(gameId, game);

    player.socket.join(gameId);

    const gameState = game.getGameState();
    delete gameState.player1.socket;
    delete gameState.player2.socket;
    
    player.socket.emit('game_started', {
      gameId,
      gameState,
      yourPlayer: 1
    });

    // Analytics
    this.analyticsService.trackEvent('game_started', {
      gameId,
      player1: player.username,
      player2: 'AI Bot',
      gameType: 'bot'
    });
  }

  handlePlayerMove(socket, data) {
    const { gameId, column } = data;
    const validation = require('../middleware/security').validateMove(column);
    
    if (!validation.valid) {
      socket.emit('error', { message: validation.error });
      return;
    }
    
    const game = this.games.get(gameId);
    if (!game) {
      socket.emit('error', { message: 'Game not found' });
      return;
    }

    const result = game.makeMove(validation.column, socket.id);
    
    if (!result.success) {
      socket.emit('error', { message: result.error });
      return;
    }

    // Broadcast move to all players in the game
    this.io.to(gameId).emit('move_made', {
      column,
      row: result.row,
      player: game.getPlayerNumber(socket.id),
      gameState: game.getGameState()
    });

    // Analytics
    this.analyticsService.trackEvent('move_made', {
      gameId,
      player: socket.id,
      column,
      row: result.row
    });

    if (result.gameOver) {
      this.handleGameEnd(game);
    } else if (game.isBot && game.currentPlayer === 2) {
      // Bot's turn
      setTimeout(() => this.makeBotMove(game), 1000);
    }
  }

  makeBotMove(game) {
    if (game.status !== 'playing' || game.currentPlayer !== 2) return;

    // Get bot's move (immediate threat/opportunity or strategic move)
    let column = this.bot.getImmediateMove(game);
    if (column === null) {
      column = this.bot.getBestMove(game);
    }

    const result = game.makeMove(column, this.bot.id);
    
    if (result.success) {
      this.io.to(game.id).emit('move_made', {
        column,
        row: result.row,
        player: 2,
        gameState: game.getGameState()
      });

      // Analytics
      this.analyticsService.trackEvent('bot_move', {
        gameId: game.id,
        column,
        row: result.row
      });

      if (result.gameOver) {
        this.handleGameEnd(game);
      }
    }
  }

  async handleGameEnd(game) {
    const duration = game.getDuration();
    
    // Save game to database
    try {
      await this.dbService.saveGame({
        id: game.id,
        player1: game.player1.username,
        player2: game.player2.username,
        winner: game.winner,
        duration,
        moves: game.moves.length,
        isBot: game.isBot,
        createdAt: game.createdAt
      });

      // Update player stats
      if (game.winner) {
        const winnerUsername = game.winner === 1 ? game.player1.username : game.player2.username;
        await this.dbService.updatePlayerStats(winnerUsername, true);
        
        if (!game.isBot) {
          const loserUsername = game.winner === 1 ? game.player2.username : game.player1.username;
          await this.dbService.updatePlayerStats(loserUsername, false);
        }
      }
    } catch (error) {
      console.error('Failed to save game:', error);
    }

    // Notify players
    this.io.to(game.id).emit('game_ended', {
      winner: game.winner,
      gameState: game.getGameState(),
      duration
    });

    // Analytics
    this.analyticsService.trackEvent('game_ended', {
      gameId: game.id,
      winner: game.winner,
      duration,
      moves: game.moves.length,
      gameType: game.isBot ? 'bot' : 'pvp'
    });

    // Cleanup
    setTimeout(() => {
      this.games.delete(game.id);
    }, 30000);
  }

  handlePlayerRejoin(socket, data) {
    const { gameId, username } = data;
    let game = null;

    if (gameId) {
      game = this.games.get(gameId);
    } else {
      game = this.findReconnectableGame(username);
    }

    if (!game) {
      socket.emit('error', { message: 'No reconnectable game found' });
      return;
    }

    // Update player socket
    if (game.player1.username === username) {
      game.player1.id = socket.id;
      game.player1.socket = socket;
    } else if (game.player2.username === username) {
      game.player2.id = socket.id;
      game.player2.socket = socket;
    }

    this.playerSockets.set(socket.id, { id: socket.id, username, socket });
    this.disconnectedPlayers.delete(username);

    socket.join(game.id);
    socket.emit('game_rejoined', {
      gameId: game.id,
      gameState: game.getGameState(),
      yourPlayer: game.getPlayerNumber(socket.id)
    });

    // Analytics
    this.analyticsService.trackEvent('player_rejoined', {
      gameId: game.id,
      player: username
    });
  }

  handlePlayerDisconnect(socket) {
    const player = this.playerSockets.get(socket.id);
    if (!player) return;

    // Remove from waiting queue
    this.waitingPlayers.delete(socket.id);
    this.playerSockets.delete(socket.id);

    // Find active game
    const game = this.findPlayerGame(socket.id);
    if (game && game.status === 'playing') {
      // Mark player as disconnected with 30-second grace period
      this.disconnectedPlayers.set(player.username, {
        gameId: game.id,
        disconnectedAt: new Date()
      });

      // Notify other player
      socket.to(game.id).emit('player_disconnected', {
        player: player.username,
        reconnectTime: 30
      });

      // Analytics
      this.analyticsService.trackEvent('player_disconnected', {
        gameId: game.id,
        player: player.username
      });
    }
  }

  findPlayerGame(socketId) {
    for (const game of this.games.values()) {
      if ((game.player1.id === socketId) || 
          (game.player2 && game.player2.id === socketId)) {
        return game;
      }
    }
    return null;
  }

  findReconnectableGame(username) {
    const disconnectedInfo = this.disconnectedPlayers.get(username);
    if (!disconnectedInfo) return null;

    const game = this.games.get(disconnectedInfo.gameId);
    if (!game || game.status !== 'playing') return null;

    return game;
  }

  cleanupDisconnectedPlayers() {
    const now = new Date();
    
    for (const [username, info] of this.disconnectedPlayers) {
      const timeDiff = (now - info.disconnectedAt) / 1000;
      
      if (timeDiff > 30) {
        const game = this.games.get(info.gameId);
        if (game && game.status === 'playing') {
          // Forfeit the game
          const disconnectedPlayer = game.player1.username === username ? 1 : 2;
          game.winner = disconnectedPlayer === 1 ? 2 : 1;
          game.status = 'finished';
          
          this.handleGameEnd(game);
        }
        
        this.disconnectedPlayers.delete(username);
      }
    }
  }
}

module.exports = GameManager;