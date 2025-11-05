#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

console.log('🚀 Setting up 4-in-a-Row Game...\n');

// Check if required tools are installed
function checkRequirements() {
  console.log('📋 Checking requirements...');
  
  const requirements = [
    { name: 'Node.js', command: 'node --version', min: '14.0.0' },
    { name: 'npm', command: 'npm --version', min: '6.0.0' },
    { name: 'PostgreSQL', command: 'psql --version', min: '12.0.0' }
  ];

  for (const req of requirements) {
    try {
      const version = execSync(req.command, { encoding: 'utf8' }).trim();
      console.log(`✅ ${req.name}: ${version}`);
    } catch (error) {
      console.log(`❌ ${req.name}: Not found or not in PATH`);
      console.log(`   Please install ${req.name} version ${req.min} or higher`);
    }
  }
  console.log('');
}

// Install dependencies
function installDependencies() {
  console.log('📦 Installing dependencies...');
  
  try {
    console.log('Installing backend dependencies...');
    execSync('npm install', { stdio: 'inherit' });
    
    console.log('Installing frontend dependencies...');
    execSync('cd client && npm install', { stdio: 'inherit', shell: true });
    
    console.log('✅ Dependencies installed successfully\n');
  } catch (error) {
    console.error('❌ Failed to install dependencies:', error.message);
    process.exit(1);
  }
}

// Create environment file if it doesn't exist
function createEnvFile() {
  console.log('⚙️  Setting up environment...');
  
  const envPath = path.join(__dirname, '.env');
  
  if (!fs.existsSync(envPath)) {
    const envContent = `PORT=3001
DB_HOST=localhost
DB_PORT=5432
DB_NAME=four_in_a_row
DB_USER=postgres
DB_PASSWORD=password
KAFKA_BROKER=localhost:9092
NODE_ENV=development`;

    fs.writeFileSync(envPath, envContent);
    console.log('✅ Created .env file with default values');
    console.log('   Please update the database credentials in .env file');
  } else {
    console.log('✅ .env file already exists');
  }
  console.log('');
}

// Setup database
function setupDatabase() {
  console.log('🗄️  Database setup...');
  console.log('Please ensure PostgreSQL is running and create the database:');
  console.log('   createdb four_in_a_row');
  console.log('   OR');
  console.log('   psql -U postgres -c "CREATE DATABASE four_in_a_row;"');
  console.log('');
}

// Setup Kafka
function setupKafka() {
  console.log('📡 Kafka setup...');
  console.log('To run with Kafka analytics:');
  console.log('1. Download and start Kafka:');
  console.log('   https://kafka.apache.org/downloads');
  console.log('2. Windows - Start Zookeeper:');
  console.log('   cd C:\\kafka');
  console.log('   .\\bin\\windows\\zookeeper-server-start.bat .\\config\\zookeeper.properties');
  console.log('3. Windows - Start Kafka:');
  console.log('   cd C:\\kafka');
  console.log('   .\\bin\\windows\\kafka-server-start.bat .\\config\\server.properties');
  console.log('4. Create topic:');
  console.log('   .\\bin\\windows\\kafka-topics.bat --create --topic game-events --bootstrap-server localhost:9092');
  console.log('');
  console.log('Note: The application will work without Kafka, but analytics will be limited.');
  console.log('');
}

// Build frontend
function buildFrontend() {
  console.log('🏗️  Building frontend...');
  
  try {
    execSync('cd client && npm run build', { stdio: 'inherit', shell: true });
    console.log('✅ Frontend built successfully\n');
  } catch (error) {
    console.error('❌ Failed to build frontend:', error.message);
    process.exit(1);
  }
}

// Final instructions
function showFinalInstructions() {
  console.log('🎉 Setup completed!\n');
  console.log('Next steps:');
  console.log('1. Update database credentials in .env file');
  console.log('2. Ensure PostgreSQL is running and database is created');
  console.log('3. (Optional) Set up Kafka for analytics');
  console.log('4. Start the application:');
  console.log('   npm start');
  console.log('');
  console.log('The application will be available at: http://localhost:3001');
  console.log('');
  console.log('For development with auto-reload:');
  console.log('   npm run dev');
  console.log('');
  console.log('Happy coding! 🚀');
}

// Main setup function
function main() {
  try {
    checkRequirements();
    installDependencies();
    createEnvFile();
    buildFrontend();
    setupDatabase();
    setupKafka();
    showFinalInstructions();
  } catch (error) {
    console.error('❌ Setup failed:', error.message);
    process.exit(1);
  }
}

// Run setup
main();