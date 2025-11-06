const redis = require('redis');

class RedisKafka {
  constructor() {
    this.client = null;
    this.subscriber = null;
    this.isInitialized = false;
  }

  async initialize() {
    try {
      const redisUrl = process.env.KAFKA_BROKER || process.env.REDIS_URL;
      if (!redisUrl) {
        console.log('No Redis URL configured, analytics disabled');
        return false;
      }

      this.client = redis.createClient({ url: redisUrl });
      this.subscriber = redis.createClient({ url: redisUrl });

      await this.client.connect();
      await this.subscriber.connect();

      this.isInitialized = true;
      console.log('Redis Kafka initialized');
      return true;
    } catch (error) {
      console.error('Redis Kafka initialization failed:', error.message);
      return false;
    }
  }

  async publish(topic, message) {
    if (!this.isInitialized) return;
    
    try {
      await this.client.lPush(topic, JSON.stringify(message));
    } catch (error) {
      console.error('Failed to publish message:', error.message);
    }
  }

  async subscribe(topic, callback) {
    if (!this.isInitialized) return;

    try {
      while (true) {
        const message = await this.subscriber.brPop({ key: topic, timeout: 1 });
        if (message) {
          const data = JSON.parse(message.element);
          await callback(data);
        }
      }
    } catch (error) {
      console.error('Subscription error:', error.message);
    }
  }

  async close() {
    try {
      if (this.client) await this.client.disconnect();
      if (this.subscriber) await this.subscriber.disconnect();
    } catch (error) {
      console.error('Error closing Redis connections:', error.message);
    }
  }
}

module.exports = RedisKafka;