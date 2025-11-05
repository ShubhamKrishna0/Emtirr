const { Kafka } = require('kafkajs');

class AnalyticsService {
  constructor() {
    this.kafka = null;
    this.producer = null;
    this.consumer = null;
    this.isInitialized = false;
  }

  async initialize() {
    try {
      this.kafka = new Kafka({
        clientId: '4-in-a-row-game',
        brokers: [process.env.KAFKA_BROKER || 'localhost:9092'],
        retry: {
          initialRetryTime: 100,
          retries: 3
        }
      });

      this.producer = this.kafka.producer();
      await this.producer.connect();

      // Start consumer for analytics processing
      this.startConsumer();
      
      this.isInitialized = true;
      console.log('Kafka Analytics Service initialized');
    } catch (error) {
      console.error('Kafka initialization failed:', error);
      this.isInitialized = false;
    }
  }

  async startConsumer() {
    try {
      this.consumer = this.kafka.consumer({ groupId: 'analytics-group' });
      await this.consumer.connect();
      await this.consumer.subscribe({ topic: 'game-events' });

      await this.consumer.run({
        eachMessage: async ({ topic, partition, message }) => {
          try {
            const event = JSON.parse(message.value.toString());
            await this.processAnalyticsEvent(event);
          } catch (error) {
            console.error('Error processing analytics event:', error);
          }
        },
      });

      console.log('Kafka consumer started for analytics');
    } catch (error) {
      console.error('Failed to start Kafka consumer:', error);
    }
  }

  async trackEvent(eventType, data) {
    const event = {
      eventType,
      timestamp: new Date().toISOString(),
      data
    };

    // Send to Kafka if available
    if (this.isInitialized && this.producer) {
      try {
        await this.producer.send({
          topic: 'game-events',
          messages: [{
            key: data.gameId || 'system',
            value: JSON.stringify(event)
          }]
        });
      } catch (error) {
        console.error('Failed to send event to Kafka:', error);
      }
    }

    // Also log locally for development
    console.log(`Analytics Event [${eventType}]:`, data);
  }

  async processAnalyticsEvent(event) {
    const { eventType, data, timestamp } = event;

    // Process different types of events
    switch (eventType) {
      case 'game_started':
        await this.processGameStarted(data, timestamp);
        break;
      case 'move_made':
        await this.processMoveEvent(data, timestamp);
        break;
      case 'game_ended':
        await this.processGameEnded(data, timestamp);
        break;
      case 'player_disconnected':
        await this.processPlayerDisconnected(data, timestamp);
        break;
      case 'player_rejoined':
        await this.processPlayerRejoined(data, timestamp);
        break;
      case 'bot_move':
        await this.processBotMove(data, timestamp);
        break;
      default:
        console.log(`Unknown event type: ${eventType}`);
    }
  }

  async processGameStarted(data, timestamp) {
    console.log(`Game Started Analytics:`, {
      gameId: data.gameId,
      gameType: data.gameType,
      players: [data.player1, data.player2],
      timestamp
    });

    // Here you could:
    // - Update real-time dashboards
    // - Send notifications
    // - Update player activity metrics
    // - Track peak gaming hours
  }

  async processMoveEvent(data, timestamp) {
    // Track move patterns, response times, etc.
    console.log(`Move Analytics:`, {
      gameId: data.gameId,
      column: data.column,
      timestamp
    });

    // Analytics you could implement:
    // - Average move time
    // - Popular column choices
    // - Move patterns analysis
  }

  async processGameEnded(data, timestamp) {
    console.log(`Game Ended Analytics:`, {
      gameId: data.gameId,
      winner: data.winner,
      duration: data.duration,
      moves: data.moves,
      gameType: data.gameType,
      timestamp
    });

    // Analytics calculations:
    // - Game duration statistics
    // - Win rate by game type
    // - Player performance metrics
    // - Bot effectiveness analysis
  }

  async processPlayerDisconnected(data, timestamp) {
    console.log(`Player Disconnection Analytics:`, {
      gameId: data.gameId,
      player: data.player,
      timestamp
    });

    // Track:
    // - Disconnection rates
    // - Common disconnection points in games
    // - Network stability metrics
  }

  async processPlayerRejoined(data, timestamp) {
    console.log(`Player Rejoin Analytics:`, {
      gameId: data.gameId,
      player: data.player,
      timestamp
    });

    // Track:
    // - Successful reconnection rates
    // - Time to reconnect
    // - User engagement metrics
  }

  async processBotMove(data, timestamp) {
    console.log(`Bot Move Analytics:`, {
      gameId: data.gameId,
      column: data.column,
      timestamp
    });

    // Track:
    // - Bot decision patterns
    // - Bot performance metrics
    // - Human vs Bot game dynamics
  }

  // Utility method to get real-time analytics
  async getRealtimeMetrics() {
    // This could query a real-time analytics store
    // For now, return mock data structure
    return {
      activeGames: 0,
      playersOnline: 0,
      gamesPlayedToday: 0,
      averageGameDuration: 0,
      botWinRate: 0,
      humanWinRate: 0
    };
  }

  // Method to generate analytics reports
  async generateReport(timeframe = '24h') {
    console.log(`Generating analytics report for ${timeframe}`);
    
    // This would typically:
    // 1. Query processed analytics data
    // 2. Generate insights and trends
    // 3. Create visualizations data
    // 4. Return formatted report
    
    return {
      timeframe,
      totalGames: 0,
      uniquePlayers: 0,
      averageDuration: 0,
      popularPlayTimes: [],
      winRateByType: {},
      playerRetention: 0
    };
  }

  async close() {
    try {
      if (this.producer) {
        await this.producer.disconnect();
      }
      if (this.consumer) {
        await this.consumer.disconnect();
      }
      console.log('Analytics service closed');
    } catch (error) {
      console.error('Error closing analytics service:', error);
    }
  }
}

module.exports = AnalyticsService;