const { Pool } = require('pg');

class DatabaseService {
  constructor() {
    this.pool = new Pool({
      host: process.env.DB_HOST,
      port: process.env.DB_PORT,
      database: process.env.DB_NAME,
      user: process.env.DB_USER,
      password: process.env.DB_PASSWORD,
    });
  }

  async initialize() {
    try {
      await this.createTables();
      console.log('Database initialized successfully');
    } catch (error) {
      console.error('Database initialization failed:', error);
      throw error;
    }
  }

  async createTables() {
    const queries = [
      `CREATE TABLE IF NOT EXISTS games (
        id VARCHAR(36) PRIMARY KEY,
        player1 VARCHAR(100) NOT NULL,
        player2 VARCHAR(100) NOT NULL,
        winner INTEGER,
        duration INTEGER NOT NULL,
        moves INTEGER NOT NULL,
        is_bot BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        finished_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )`,
      
      `CREATE TABLE IF NOT EXISTS players (
        username VARCHAR(100) PRIMARY KEY,
        games_played INTEGER DEFAULT 0,
        games_won INTEGER DEFAULT 0,
        total_duration INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        last_played TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )`,
      
      `CREATE TABLE IF NOT EXISTS analytics_events (
        id SERIAL PRIMARY KEY,
        event_type VARCHAR(50) NOT NULL,
        game_id VARCHAR(36),
        player VARCHAR(100),
        data JSONB,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )`,
      
      `CREATE INDEX IF NOT EXISTS idx_games_created_at ON games(created_at)`,
      `CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC)`,
      `CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type)`,
      `CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at)`
    ];

    for (const query of queries) {
      await this.pool.query(query);
    }
  }

  async saveGame(gameData) {
    const query = `
      INSERT INTO games (id, player1, player2, winner, duration, moves, is_bot, created_at)
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `;
    
    await this.pool.query(query, [
      gameData.id,
      gameData.player1,
      gameData.player2,
      gameData.winner,
      gameData.duration,
      gameData.moves,
      gameData.isBot,
      gameData.createdAt
    ]);
  }

  async updatePlayerStats(username, won) {
    // Insert or update player stats
    const query = `
      INSERT INTO players (username, games_played, games_won, last_played)
      VALUES ($1, 1, $2, CURRENT_TIMESTAMP)
      ON CONFLICT (username) 
      DO UPDATE SET 
        games_played = players.games_played + 1,
        games_won = players.games_won + $2,
        last_played = CURRENT_TIMESTAMP
    `;
    
    await this.pool.query(query, [username, won ? 1 : 0]);
  }

  async getLeaderboard(limit = 10) {
    const query = `
      SELECT 
        username,
        games_played,
        games_won,
        ROUND((games_won::DECIMAL / GREATEST(games_played, 1)) * 100, 1) as win_rate,
        last_played
      FROM players 
      WHERE games_played > 0
      ORDER BY games_won DESC, win_rate DESC, games_played DESC
      LIMIT $1
    `;
    
    const result = await this.pool.query(query, [limit]);
    return result.rows;
  }

  async getAnalytics() {
    const queries = {
      totalGames: 'SELECT COUNT(*) as count FROM games',
      totalPlayers: 'SELECT COUNT(*) as count FROM players WHERE games_played > 0',
      avgGameDuration: 'SELECT ROUND(AVG(duration), 1) as avg_duration FROM games',
      gamesPerDay: `
        SELECT 
          DATE(created_at) as date,
          COUNT(*) as games
        FROM games 
        WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
        GROUP BY DATE(created_at)
        ORDER BY date DESC
      `,
      topWinners: `
        SELECT username, games_won 
        FROM players 
        WHERE games_played > 0
        ORDER BY games_won DESC 
        LIMIT 5
      `,
      botVsHuman: `
        SELECT 
          is_bot,
          COUNT(*) as count,
          ROUND(AVG(duration), 1) as avg_duration
        FROM games 
        GROUP BY is_bot
      `
    };

    const results = {};
    
    for (const [key, query] of Object.entries(queries)) {
      try {
        const result = await this.pool.query(query);
        results[key] = result.rows;
      } catch (error) {
        console.error(`Analytics query failed for ${key}:`, error);
        results[key] = [];
      }
    }

    return results;
  }

  async saveAnalyticsEvent(eventType, gameId, player, data) {
    const query = `
      INSERT INTO analytics_events (event_type, game_id, player, data)
      VALUES ($1, $2, $3, $4)
    `;
    
    try {
      await this.pool.query(query, [eventType, gameId, player, JSON.stringify(data)]);
    } catch (error) {
      console.error('Failed to save analytics event:', error);
    }
  }

  async saveGameAnalytics(analyticsData) {
    const query = `
      INSERT INTO analytics_events (event_type, game_id, data)
      VALUES ('game_analytics', $1, $2)
    `;
    
    try {
      await this.pool.query(query, [analyticsData.gameId, JSON.stringify(analyticsData)]);
    } catch (error) {
      console.error('Failed to save game analytics:', error);
    }
  }

  async getRecentGames(limit = 20) {
    const query = `
      SELECT 
        id,
        player1,
        player2,
        winner,
        duration,
        moves,
        is_bot,
        created_at
      FROM games 
      ORDER BY created_at DESC 
      LIMIT $1
    `;
    
    const result = await this.pool.query(query, [limit]);
    return result.rows;
  }

  async close() {
    await this.pool.end();
  }
}

module.exports = DatabaseService;