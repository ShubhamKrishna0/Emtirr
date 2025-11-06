const fs = require('fs');
const path = require('path');

class Logger {
  constructor() {
    this.logDir = path.join(__dirname, '../../logs');
    this.ensureLogDirectory();
  }

  ensureLogDirectory() {
    if (!fs.existsSync(this.logDir)) {
      fs.mkdirSync(this.logDir, { recursive: true });
    }
  }

  formatMessage(level, message, data = null) {
    const timestamp = new Date().toISOString();
    const logEntry = {
      timestamp,
      level,
      message,
      ...(data && { data })
    };
    
    return JSON.stringify(logEntry);
  }

  writeToFile(filename, message) {
    const logFile = path.join(this.logDir, filename);
    const logLine = message + '\n';
    
    fs.appendFileSync(logFile, logLine);
  }

  info(message, data = null) {
    const formattedMessage = this.formatMessage('INFO', message, data);
    console.log(`ℹ️  ${message}`, data || '');
    
    if (process.env.NODE_ENV === 'production') {
      this.writeToFile('app.log', formattedMessage);
    }
  }

  error(message, error = null) {
    const errorData = error ? {
      message: error.message,
      stack: error.stack,
      ...(error.code && { code: error.code })
    } : null;
    
    const formattedMessage = this.formatMessage('ERROR', message, errorData);
    console.error(`❌ ${message}`, error || '');
    
    if (process.env.NODE_ENV === 'production') {
      this.writeToFile('error.log', formattedMessage);
    }
  }

  warn(message, data = null) {
    const formattedMessage = this.formatMessage('WARN', message, data);
    console.warn(`⚠️  ${message}`, data || '');
    
    if (process.env.NODE_ENV === 'production') {
      this.writeToFile('app.log', formattedMessage);
    }
  }

  debug(message, data = null) {
    if (process.env.NODE_ENV === 'development') {
      const formattedMessage = this.formatMessage('DEBUG', message, data);
      console.log(`🐛 ${message}`, data || '');
    }
  }

  game(message, gameData = null) {
    const formattedMessage = this.formatMessage('GAME', message, gameData);
    console.log(`🎮 ${message}`, gameData || '');
    
    if (process.env.NODE_ENV === 'production') {
      this.writeToFile('game.log', formattedMessage);
    }
  }

  analytics(event, data = null) {
    const formattedMessage = this.formatMessage('ANALYTICS', event, data);
    console.log(`📊 ${event}`, data || '');
    
    if (process.env.NODE_ENV === 'production') {
      this.writeToFile('analytics.log', formattedMessage);
    }
  }
}

// Create singleton instance
const logger = new Logger();

module.exports = logger;