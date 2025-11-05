# 🚀 Deployment Guide

This guide covers different deployment options for the 4-in-a-Row game.

## 🏠 Local Development

### Prerequisites
- Node.js 14+
- PostgreSQL 12+
- (Optional) Apache Kafka 2.8+

### Quick Start
```bash
# Windows
start.bat

# Linux/Mac
chmod +x setup.js
node setup.js
npm start
```

## ☁️ Cloud Deployment Options

### 1. Heroku Deployment

#### Setup
```bash
# Install Heroku CLI
npm install -g heroku

# Login and create app
heroku login
heroku create your-app-name

# Add PostgreSQL addon
heroku addons:create heroku-postgresql:hobby-dev

# Set environment variables
heroku config:set NODE_ENV=production
heroku config:set KAFKA_BROKER=your-kafka-url

# Deploy
git push heroku main
```

#### Heroku Configuration
Create `Procfile`:
```
web: node server.js
```

### 2. AWS Deployment

#### Using AWS Elastic Beanstalk
```bash
# Install EB CLI
pip install awsebcli

# Initialize and deploy
eb init
eb create production
eb deploy
```

#### Using AWS ECS with Docker
```bash
# Build and push to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin your-account.dkr.ecr.us-east-1.amazonaws.com
docker build -t 4-in-a-row .
docker tag 4-in-a-row:latest your-account.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row:latest
docker push your-account.dkr.ecr.us-east-1.amazonaws.com/4-in-a-row:latest
```

### 3. Digital Ocean App Platform

#### app.yaml
```yaml
name: 4-in-a-row
services:
- name: web
  source_dir: /
  github:
    repo: your-username/4-in-a-row
    branch: main
  run_command: npm start
  environment_slug: node-js
  instance_count: 1
  instance_size_slug: basic-xxs
  envs:
  - key: NODE_ENV
    value: production
databases:
- engine: PG
  name: gamedb
  num_nodes: 1
  size: db-s-dev-database
  version: "13"
```

### 4. Railway Deployment

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway init
railway up
```

## 🐳 Docker Deployment

### Local Docker
```bash
# Build and run
docker build -t 4-in-a-row .
docker run -p 3001:3001 4-in-a-row
```

### Docker Compose (Recommended)
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Production Docker Compose
```yaml
version: '3.8'
services:
  app:
    image: your-registry/4-in-a-row:latest
    ports:
      - "80:3001"
    environment:
      - NODE_ENV=production
      - DB_HOST=postgres
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: four_in_a_row
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    
  nginx:
    image: nginx:alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  postgres_data:
```

## 🔧 Environment Configuration

### Production Environment Variables
```env
NODE_ENV=production
PORT=3001

# Database
DB_HOST=your-db-host
DB_PORT=5432
DB_NAME=four_in_a_row
DB_USER=your-db-user
DB_PASSWORD=your-secure-password

# Kafka (Optional)
KAFKA_BROKER=your-kafka-broker:9092

# Security
SESSION_SECRET=your-session-secret
CORS_ORIGIN=https://yourdomain.com
```

## 🔒 Security Considerations

### SSL/TLS Setup
```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    ssl_certificate /etc/ssl/cert.pem;
    ssl_certificate_key /etc/ssl/key.pem;
    
    location / {
        proxy_pass http://app:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Rate Limiting
```javascript
const rateLimit = require('express-rate-limit');

const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100 // limit each IP to 100 requests per windowMs
});

app.use('/api/', limiter);
```

## 📊 Monitoring & Analytics

### Health Checks
```javascript
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    memory: process.memoryUsage()
  });
});
```

### Logging in Production
```javascript
const winston = require('winston');

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  transports: [
    new winston.transports.File({ filename: 'error.log', level: 'error' }),
    new winston.transports.File({ filename: 'combined.log' })
  ]
});
```

## 🔄 CI/CD Pipeline

### GitHub Actions
```yaml
name: Deploy to Production

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Setup Node.js
      uses: actions/setup-node@v2
      with:
        node-version: '18'
        
    - name: Install dependencies
      run: |
        npm ci
        cd client && npm ci
        
    - name: Run tests
      run: npm test
      
    - name: Build application
      run: |
        cd client && npm run build
        
    - name: Deploy to Heroku
      uses: akhileshns/heroku-deploy@v3.12.12
      with:
        heroku_api_key: ${{secrets.HEROKU_API_KEY}}
        heroku_app_name: "your-app-name"
        heroku_email: "your-email@example.com"
```

## 📈 Performance Optimization

### Database Optimization
```sql
-- Add indexes for better performance
CREATE INDEX idx_games_created_at ON games(created_at);
CREATE INDEX idx_players_games_won ON players(games_won DESC);
CREATE INDEX idx_analytics_events_type ON analytics_events(event_type);
```

### Caching Strategy
```javascript
const redis = require('redis');
const client = redis.createClient(process.env.REDIS_URL);

// Cache leaderboard for 5 minutes
app.get('/api/leaderboard', async (req, res) => {
  const cached = await client.get('leaderboard');
  if (cached) {
    return res.json(JSON.parse(cached));
  }
  
  const leaderboard = await dbService.getLeaderboard();
  await client.setex('leaderboard', 300, JSON.stringify(leaderboard));
  res.json(leaderboard);
});
```

## 🚨 Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```bash
   # Find and kill process using port 3001
   netstat -ano | findstr :3001
   taskkill /PID <PID> /F
   ```

2. **Database Connection Issues**
   ```bash
   # Check PostgreSQL status
   pg_isready -h localhost -p 5432
   
   # Test connection
   psql -h localhost -U postgres -d four_in_a_row
   ```

3. **Memory Issues**
   ```javascript
   // Monitor memory usage
   setInterval(() => {
     const used = process.memoryUsage();
     console.log('Memory usage:', Math.round(used.rss / 1024 / 1024 * 100) / 100, 'MB');
   }, 30000);
   ```

## 📞 Support

For deployment issues:
1. Check the logs: `docker-compose logs -f`
2. Verify environment variables
3. Test database connectivity
4. Check firewall settings
5. Review application health endpoint

Remember to update DNS records and SSL certificates for production domains!