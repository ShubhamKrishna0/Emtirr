#!/usr/bin/env node

require('dotenv').config();
const KafkaConsumer = require('./src/services/KafkaConsumer');

const consumer = new KafkaConsumer();

async function startAnalyticsService() {
  console.log('Starting Analytics Consumer Service...');
  
  const initialized = await consumer.initialize();
  if (!initialized) {
    console.log('Kafka not available, exiting analytics service');
    process.exit(0);
  }

  await consumer.start();

  // Graceful shutdown
  process.on('SIGINT', async () => {
    console.log('Shutting down analytics consumer...');
    await consumer.stop();
    process.exit(0);
  });

  process.on('SIGTERM', async () => {
    console.log('Shutting down analytics consumer...');
    await consumer.stop();
    process.exit(0);
  });

  // Log metrics every 5 minutes
  setInterval(() => {
    const metrics = consumer.getMetrics();
    console.log('Analytics Metrics:', JSON.stringify(metrics, null, 2));
  }, 5 * 60 * 1000);
}

startAnalyticsService().catch(error => {
  console.error('Analytics service failed:', error);
  process.exit(1);
});