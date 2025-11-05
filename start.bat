@echo off
echo 🚀 Starting 4-in-a-Row Game Server...
echo.

REM Check if Node.js is installed
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    pause
    exit /b 1
)

REM Check if dependencies are installed
if not exist "node_modules" (
    echo 📦 Installing dependencies...
    npm install
    if %errorlevel% neq 0 (
        echo ❌ Failed to install dependencies
        pause
        exit /b 1
    )
)

REM Check if frontend is built
if not exist "client\build" (
    echo 🏗️ Building frontend...
    cd client
    npm run build
    if %errorlevel% neq 0 (
        echo ❌ Failed to build frontend
        pause
        exit /b 1
    )
    cd ..
)

REM Start the server
echo ✅ Starting server on http://localhost:3001
echo.
echo 📝 Note: Make sure PostgreSQL is running and database is created
echo 📝 Kafka is optional but recommended for full analytics
echo.
node server.js

pause