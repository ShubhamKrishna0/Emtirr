# Kafka Analytics Deployment Guide

## Overview
The Kafka analytics system is designed to be **completely optional** and deployment-safe. The application will work perfectly without Kafka.

## Deployment Options

### 1. Without Kafka (Default)
```bash
# Standard deployment - no analytics
docker-compose up -d app postgres
```

### 2. With Kafka Analytics
```bash
# Full deployment with analytics
docker-compose --profile analytics up -d
```

### 3. Local Development
```bash
# Terminal 1: Start main app
npm start

# Terminal 2: Start analytics consumer (optional)
npm run analytics
```

## Environment Variables

### Required (App will work without these)
```env
KAFKA_BROKER=localhost:9092  # Optional - app works without this
```

### Optional SSL Configuration
```env
KAFKA_SSL_CA_PATH=/path/to/ca.pem
KAFKA_SSL_KEY_PATH=/path/to/service.key  
KAFKA_SSL_CERT_PATH=/path/to/service.cert
```

## Analytics Features

### Tracked Events
- `game_started` - New game initialization
- `game_ended` - Game completion with winner/duration
- `move_made` - Player moves
- `player_disconnected` - Connection issues
- `player_rejoined` - Reconnections

### Metrics Calculated
- **Game Duration**: Average game length
- **Win Rates**: Bot vs Human performance
- **Player Activity**: Games per hour/day
- **Popular Moves**: Column preferences
- **Connection Stability**: Disconnect/reconnect rates

### Data Storage
- **Kafka Topics**: Real-time event streaming
- **PostgreSQL**: Processed analytics storage
- **Console Logs**: Development debugging

## Production Deployment

### Heroku (Without Kafka)
```bash
heroku create your-app
heroku addons:create heroku-postgresql:hobby-dev
git push heroku main
```

### AWS/GCP (With Kafka)
```bash
# Use managed Kafka service
export KAFKA_BROKER=your-managed-kafka-endpoint:9092
docker-compose --profile analytics up -d
```

### Docker Swarm
```bash
# Deploy without analytics
docker stack deploy -c docker-compose.yml game-stack

# Deploy with analytics
docker stack deploy -c docker-compose.yml --compose-file docker-compose.analytics.yml game-stack
```

## Troubleshooting

### Common Issues
1. **Kafka Connection Failed**: App continues normally, analytics disabled
2. **Consumer Crashes**: Only affects analytics, game continues
3. **Topic Not Found**: Auto-created on first message

### Health Checks
```bash
# Check if analytics is working
curl http://localhost:3001/api/analytics

# View analytics logs
docker logs 4-in-a-row-analytics
```

### Manual Topic Creation
```bash
# If auto-creation is disabled
kafka-topics.sh --create --topic game-events --bootstrap-server localhost:9092
```

## Monitoring

### Key Metrics to Monitor
- Message throughput
- Consumer lag
- Error rates
- Database storage growth

### Sample Analytics Output
```json
{
  "totalGames": 150,
  "botWinRate": "45.2",
  "humanWinRate": "54.8", 
  "averageGameDurationSeconds": 180,
  "totalMoves": 2340
}
```

## Scaling

### High Volume Scenarios
- Increase Kafka partitions
- Add more consumer instances
- Use batch processing for database writes
- Implement data retention policies

The system is designed to gracefully handle failures and scale as needed.