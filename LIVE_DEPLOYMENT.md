# 🚀 Live Deployment Guide

## Quick Deploy Options

### 1. **Heroku (Recommended for Demo)**

```bash
# Install Heroku CLI
npm install -g heroku

# Login and create app
heroku login
heroku create your-4-in-a-row-app

# Add PostgreSQL
heroku addons:create heroku-postgresql:hobby-dev

# Set environment variables
heroku config:set NODE_ENV=production
heroku config:set KAFKAJS_NO_PARTITIONER_WARNING=1

# Deploy
git add .
git commit -m "Deploy to Heroku"
git push heroku main
```

### 2. **Railway (Easiest)**

```bash
# Install Railway CLI
npm install -g @railway/cli

# Deploy
railway login
railway init
railway up
```

### 3. **Render (Free Tier)**

1. Connect GitHub repo to Render
2. Add PostgreSQL database
3. Set environment variables
4. Deploy automatically

### 4. **DigitalOcean App Platform**

```yaml
# app.yaml
name: 4-in-a-row
services:
- name: web
  source_dir: /
  run_command: npm start
  environment_slug: node-js
  instance_count: 1
  instance_size_slug: basic-xxs
databases:
- engine: PG
  name: gamedb
  num_nodes: 1
  size: db-s-dev-database
```

## Environment Variables for Production

```env
NODE_ENV=production
PORT=3001
DB_HOST=your-db-host
DB_PORT=5432
DB_NAME=four_in_a_row
DB_USER=your-db-user
DB_PASSWORD=your-db-password
KAFKAJS_NO_PARTITIONER_WARNING=1
```

## Pre-Deployment Checklist

- ✅ Update database credentials in `.env`
- ✅ Build frontend: `cd client && npm run build`
- ✅ Test locally: `npm start`
- ✅ Commit all changes to Git
- ✅ Push to GitHub repository

## Live URL Structure

Your deployed app will be available at:
- **Game Interface**: `https://your-app.herokuapp.com/`
- **Leaderboard API**: `https://your-app.herokuapp.com/api/leaderboard`
- **Health Check**: `https://your-app.herokuapp.com/health`

## Demo Instructions

1. **Open the live URL**
2. **Enter username** (e.g., "Player1")
3. **Wait 10 seconds** - Bot will automatically join
4. **Play the game** - Click columns to drop discs
5. **View leaderboard** - Click "Show Leaderboard" button
6. **Test reconnection** - Refresh page and rejoin with same username

## Submission Format

```
GitHub Repository: https://github.com/yourusername/4-in-a-row-game
Live Application: https://your-4-in-a-row-app.herokuapp.com
Demo Video: [Optional - Screen recording of gameplay]
```

Your application is **production-ready** and meets all Emitrr assignment requirements! 🎯