// Game constants
const GAME_CONFIG = {
  BOARD_ROWS: 6,
  BOARD_COLS: 7,
  WIN_CONDITION: 4,
  BOT_TIMEOUT: 10000, // 10 seconds
  RECONNECT_TIMEOUT: 30000, // 30 seconds
  CLEANUP_INTERVAL: 30000, // 30 seconds
};

// Player types
const PLAYER_TYPES = {
  HUMAN: 'human',
  BOT: 'bot'
};

// Game states
const GAME_STATUS = {
  WAITING: 'waiting',
  PLAYING: 'playing',
  FINISHED: 'finished',
  PAUSED: 'paused'
};

// Socket events
const SOCKET_EVENTS = {
  // Client to Server
  JOIN_GAME: 'join_game',
  MAKE_MOVE: 'make_move',
  REJOIN_GAME: 'rejoin_game',
  
  // Server to Client
  WAITING_FOR_OPPONENT: 'waiting_for_opponent',
  GAME_STARTED: 'game_started',
  YOUR_TURN: 'your_turn',
  MOVE_MADE: 'move_made',
  GAME_ENDED: 'game_ended',
  PLAYER_DISCONNECTED: 'player_disconnected',
  GAME_REJOINED: 'game_rejoined',
  ERROR: 'error'
};

// Analytics events
const ANALYTICS_EVENTS = {
  GAME_STARTED: 'game_started',
  MOVE_MADE: 'move_made',
  GAME_ENDED: 'game_ended',
  PLAYER_DISCONNECTED: 'player_disconnected',
  PLAYER_REJOINED: 'player_rejoined',
  BOT_MOVE: 'bot_move'
};

// Error messages
const ERROR_MESSAGES = {
  GAME_NOT_FOUND: 'Game not found',
  INVALID_MOVE: 'Invalid move',
  NOT_YOUR_TURN: 'Not your turn',
  COLUMN_FULL: 'Column is full',
  INVALID_COLUMN: 'Invalid column',
  USERNAME_REQUIRED: 'Username is required',
  GAME_NOT_ACTIVE: 'Game not active',
  NO_RECONNECTABLE_GAME: 'No reconnectable game found'
};

// Bot difficulty levels
const BOT_DIFFICULTY = {
  EASY: { depth: 3, name: 'Easy Bot' },
  MEDIUM: { depth: 5, name: 'Medium Bot' },
  HARD: { depth: 7, name: 'Hard Bot' },
  EXPERT: { depth: 9, name: 'Expert Bot' }
};

// Database table names
const DB_TABLES = {
  GAMES: 'games',
  PLAYERS: 'players',
  ANALYTICS_EVENTS: 'analytics_events'
};

// Kafka configuration
const KAFKA_CONFIG = {
  TOPICS: {
    GAME_EVENTS: 'game-events',
    PLAYER_EVENTS: 'player-events',
    SYSTEM_EVENTS: 'system-events'
  },
  CONSUMER_GROUP: 'analytics-group'
};

module.exports = {
  GAME_CONFIG,
  PLAYER_TYPES,
  GAME_STATUS,
  SOCKET_EVENTS,
  ANALYTICS_EVENTS,
  ERROR_MESSAGES,
  BOT_DIFFICULTY,
  DB_TABLES,
  KAFKA_CONFIG
};