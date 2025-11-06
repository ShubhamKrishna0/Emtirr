const { Kafka } = require('kafkajs');
const RedisKafka = require('./RedisKafka');
const DatabaseService = require('./DatabaseService');

class KafkaConsumer {
  constructor() {
    this.kafka = null;
    this.consumer = null;
    this.redisKafka = null;
    this.dbService = new DatabaseService();
    this.isRunning = false;
    this.useRedis = false;
    this.metrics = {
      totalGames: 0,
      totalMoves: 0,
      averageGameDuration: 0,
      botWins: 0,
      humanWins: 0,
      gamesPerHour: {},
      dailyStats: {}
    };
  }

  async initialize() {
    try {
      const brokerUrl = process.env.KAFKA_BROKER || process.env.REDIS_URL;
      if (!brokerUrl) {
        console.log('No broker configured');
        return false;
      }
      
      // Try Redis first
      if (brokerUrl.includes('redis://') || process.env.REDIS_URL) {
        this.redisKafka = new RedisKafka();
        const initialized = await this.redisKafka.initialize();
        if (initialized) {
          this.useRedis = true;
          console.log('Redis Analytics Consumer initialized');
          return true;
        }
      }

      // Fallback to Kafka
      this.kafka = new Kafka({
        clientId: 'analytics-consumer',
        brokers: [brokerUrl],
        connectionTimeout: 5000,
        requestTimeout: 10000,
        retry: {
          initialRetryTime: 100,
          retries: 3
        }
      });

      this.consumer = this.kafka.consumer({ 
        groupId: 'analytics-group',
        sessionTimeout: 30000,
        heartbeatInterval: 3000
      });

      await this.consumer.connect();
      await this.consumer.subscribe({ topic: 'game-events' });
      
      console.log('Kafka Analytics Consumer initialized');
      return true;
    } catch (error) {
      console.error('Consumer initialization failed:', error.message);
      return false;
    }
  }

  async start() {
    try {
      this.isRunning = true;
      
      if (this.useRedis) {
        await this.redisKafka.subscribe('game-events', async (event) => {
          await this.processEvent(event);
        });
      } else if (this.consumer) {
        await this.consumer.run({
          eachMessage: async ({ message }) => {
            try {
              const event = JSON.parse(message.value.toString());
              await this.processEvent(event);
            } catch (error) {
              console.error('Error processing message:', error.message);
            }
          },
        });
      }

      console.log('Analytics Consumer started');
    } catch (error) {
      console.error('Failed to start consumer:', error.message);
      this.isRunning = false;
    }
  }

  async processEvent(event) {
    const { eventType, data, timestamp } = event;
    
    switch (eventType) {
      case 'game_started':
        this.trackGameStart(data, timestamp);
        break;
      case 'game_ended':
        await this.trackGameEnd(data, timestamp);
        break;
      case 'move_made':
        this.trackMove(data, timestamp);
        break;
      default:
        console.log(`Processing ${eventType} event`);
    }
  }

  trackGameStart(data, timestamp) {
    const hour = new Date(timestamp).getHours();
    this.metrics.gamesPerHour[hour] = (this.metrics.gamesPerHour[hour] || 0) + 1;
    
    console.log(`Game started: ${data.gameId} at ${timestamp}`);
  }

  async trackGameEnd(data, timestamp) {
    this.metrics.totalGames++;
    
    if (data.winner === 'bot') {
      this.metrics.botWins++;
    } else if (data.winner !== 'draw') {
      this.metrics.humanWins++;
    }

    if (data.duration) {
      this.metrics.averageGameDuration = 
        (this.metrics.averageGameDuration * (this.metrics.totalGames - 1) + data.duration) / this.metrics.totalGames;
    }

    // Store in database if available
    try {
      await this.dbService.saveGameAnalytics({
        gameId: data.gameId,
        winner: data.winner,
        duration: data.duration,
        moves: data.moves,
        timestamp
      });
    } catch (error) {
      console.log('Database unavailable for analytics storage');
    }

    console.log(`Game ended: ${data.gameId}, Winner: ${data.winner}, Duration: ${data.duration}ms`);
  }

  trackMove(data, timestamp) {
    this.metrics.totalMoves++;
    console.log(`Move tracked: Game ${data.gameId}, Column ${data.column}`);
  }

  getMetrics() {
    return {
      ...this.metrics,
      botWinRate: this.metrics.totalGames > 0 ? 
        (this.metrics.botWins / this.metrics.totalGames * 100).toFixed(1) : 0,
      humanWinRate: this.metrics.totalGames > 0 ? 
        (this.metrics.humanWins / this.metrics.totalGames * 100).toFixed(1) : 0,
      averageGameDurationSeconds: Math.round(this.metrics.averageGameDuration / 1000)
    };
  }

  async stop() {
    try {
      this.isRunning = false;
      if (this.useRedis && this.redisKafka) {
        await this.redisKafka.close();
      } else if (this.consumer) {
        await this.consumer.disconnect();
      }
      console.log('Analytics Consumer stopped');
    } catch (error) {
      console.error('Error stopping consumer:', error.message);
    }
  }
}

module.exports = KafkaConsumer;