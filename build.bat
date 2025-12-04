@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   AnyProxyAi - Build Script (Windows)
echo ========================================
echo.

REM 检查依赖
where go >nul 2>nul || (echo [ERROR] Go not found & pause & exit /b 1)
where node >nul 2>nul || (echo [ERROR] Node.js not found & pause & exit /b 1)
where wails >nul 2>nul || (
    echo [INFO] Installing Wails CLI...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
)

REM 清空 build 目录
echo [0/4] Cleaning build directory...
if exist "build\bin\anyproxyai-windows*" del /q "build\bin\anyproxyai-windows*"
if not exist "build\bin" mkdir "build\bin"

echo [1/4] Installing frontend dependencies...
cd frontend
call npm install
cd ..

echo [2/4] Building frontend...
cd frontend
call npm run build
cd ..

echo [3/4] Downloading Go dependencies...
go mod tidy
go mod download

echo.
echo [4/4] Building Windows GUI...
echo.

echo [Windows amd64] Building...
wails build -platform windows/amd64 -o anyproxyai-windows-amd64.exe
if %errorlevel% equ 0 (echo [OK] Windows amd64) else (echo [FAIL] Windows amd64)

echo [Windows arm64] Building...
wails build -platform windows/arm64 -o anyproxyai-windows-arm64.exe
if %errorlevel% equ 0 (echo [OK] Windows arm64) else (echo [FAIL] Windows arm64)

echo.
echo ========================================
echo   Build Complete!
echo ========================================
echo.
echo Output: build\bin\
dir /b build\bin\anyproxyai-windows*.exe 2>nul
echo.
echo ========================================
echo Notes:
echo   - Windows builds: Full GUI with system tray
echo   - Linux/macOS builds: Use GitHub Actions or build on native platform
echo   - See .github/workflows/build.yml for CI/CD
echo ========================================
pause
