const { Kafka } = require('kafkajs');
const RedisKafka = require('./RedisKafka');

class AnalyticsService {
  constructor() {
    this.kafka = null;
    this.producer = null;
    this.redisKafka = null;
    this.isInitialized = false;
    this.useRedis = false;
  }

  async initialize() {
    try {
      const brokerUrl = process.env.KAFKA_BROKER;
      if (!brokerUrl) {
        console.log('No broker configured, analytics disabled');
        return;
      }

      // Try Redis first (for Render)
      if (brokerUrl.includes('redis://') || process.env.REDIS_URL) {
        this.redisKafka = new RedisKafka();
        this.isInitialized = await this.redisKafka.initialize();
        this.useRedis = this.isInitialized;
        if (this.isInitialized) {
          console.log('Redis Analytics initialized');
          return;
        }
      }

      // Fallback to Kafka
      this.kafka = new Kafka({
        clientId: '4-in-a-row-producer',
        brokers: [brokerUrl],
        connectionTimeout: 5000,
        requestTimeout: 10000,
        retry: {
          initialRetryTime: 100,
          retries: 2
        }
      });

      this.producer = this.kafka.producer();
      await this.producer.connect();
      
      this.isInitialized = true;
      console.log('Kafka Analytics Producer initialized');
    } catch (error) {
      console.error('Analytics initialization failed:', error.message);
      this.isInitialized = false;
    }
  }



  async trackEvent(eventType, data) {
    const event = {
      eventType,
      timestamp: new Date().toISOString(),
      data
    };

    if (this.isInitialized) {
      try {
        if (this.useRedis) {
          await this.redisKafka.publish('game-events', event);
        } else if (this.producer) {
          await this.producer.send({
            topic: 'game-events',
            messages: [{
              key: data.gameId || 'system',
              value: JSON.stringify(event)
            }]
          });
        }
      } catch (error) {
        console.error('Failed to send analytics event:', error);
      }
    }

    console.log(`Analytics [${eventType}]:`, {
      gameId: data.gameId,
      player: data.player || data.winner,
      timestamp: event.timestamp
    });
  }



  async close() {
    try {
      if (this.useRedis && this.redisKafka) {
        await this.redisKafka.close();
      } else if (this.producer) {
        await this.producer.disconnect();
      }
      console.log('Analytics service closed');
    } catch (error) {
      console.error('Error closing analytics service:', error);
    }
  }
}

module.exports = AnalyticsService;