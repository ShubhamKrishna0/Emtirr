@echo off
echo Installing Go dependencies...
go mod tidy

echo Building Go application...
go build -o four-in-a-row.exe main.go

echo Build complete! Run with: four-in-a-row.exe
pause