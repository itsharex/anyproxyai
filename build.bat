@echo off
echo ========================================
echo   OpenAI Router - Cross Platform Build
echo ========================================
echo.

REM 检查是否安装了 Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM 检查是否安装了 Node.js
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    pause
    exit /b 1
)

REM 检查是否安装了 Wails
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARN] Wails is not installed
    echo Installing Wails CLI...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install Wails
        pause
        exit /b 1
    )
)

echo [1/5] Installing frontend dependencies...
cd frontend
call npm install
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install frontend dependencies
    pause
    exit /b 1
)
cd ..

echo.
echo [2/5] Building frontend...
cd frontend
call npm run build
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build frontend
    pause
    exit /b 1
)
cd ..

echo.
echo [3/5] Downloading Go dependencies...
go mod tidy
go mod download
if %errorlevel% neq 0 (
    echo [ERROR] Failed to download Go dependencies
    pause
    exit /b 1
)

echo.
echo [4/5] Building Wails desktop application for Windows...
wails build
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build Wails application
    pause
    exit /b 1
)

echo.
echo ========================================
echo   Build Complete!
echo ========================================
echo.
echo Wails GUI application (with embedded API server):
echo   build\bin\openai-router.exe
echo.
echo To start: run start.bat or double-click openai-router.exe
echo.
echo Note: For Linux/macOS, run 'wails build' on those platforms
echo ========================================
pause
