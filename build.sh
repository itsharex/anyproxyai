#!/bin/bash

echo "========================================"
echo "  OpenAI Router - Cross Platform Build"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Go installation
if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERROR]${NC} Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check Node.js installation
if ! command -v node &> /dev/null; then
    echo -e "${RED}[ERROR]${NC} Node.js is not installed or not in PATH"
    echo "Please install Node.js from https://nodejs.org/"
    exit 1
fi

# Check Wails installation
if ! command -v wails &> /dev/null; then
    echo -e "${YELLOW}[WARN]${NC} Wails is not installed"
    echo "Installing Wails CLI..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if [ $? -ne 0 ]; then
        echo -e "${RED}[ERROR]${NC} Failed to install Wails"
        exit 1
    fi
fi

echo "[1/5] Installing frontend dependencies..."
cd frontend
npm install
if [ $? -ne 0 ]; then
    echo -e "${RED}[ERROR]${NC} Failed to install frontend dependencies"
    exit 1
fi
cd ..

echo ""
echo "[2/5] Building frontend..."
cd frontend
npm run build
if [ $? -ne 0 ]; then
    echo -e "${RED}[ERROR]${NC} Failed to build frontend"
    exit 1
fi
cd ..

echo ""
echo "[3/5] Downloading Go dependencies..."
go mod tidy
go mod download
if [ $? -ne 0 ]; then
    echo -e "${RED}[ERROR]${NC} Failed to download Go dependencies"
    exit 1
fi

echo ""
echo "[4/5] Building Wails desktop application..."
wails build
if [ $? -ne 0 ]; then
    echo -e "${RED}[ERROR]${NC} Failed to build Wails application"
    exit 1
fi

echo ""
echo "[5/5] Building cross-platform binaries..."

# Create build directory
mkdir -p build/bin

# Windows
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/bin/openai-router-windows-amd64.exe .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}[OK]${NC} Windows amd64 build complete"
else
    echo -e "${YELLOW}[WARN]${NC} Failed to build for Windows amd64"
fi

# Linux
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/bin/openai-router-linux-amd64 .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}[OK]${NC} Linux amd64 build complete"
else
    echo -e "${YELLOW}[WARN]${NC} Failed to build for Linux amd64"
fi

# macOS (Intel)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/bin/openai-router-darwin-amd64 .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}[OK]${NC} macOS amd64 build complete"
else
    echo -e "${YELLOW}[WARN]${NC} Failed to build for macOS amd64"
fi

# macOS (Apple Silicon)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/bin/openai-router-darwin-arm64 .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}[OK]${NC} macOS arm64 build complete"
else
    echo -e "${YELLOW}[WARN]${NC} Failed to build for macOS arm64"
fi

# Make binaries executable
chmod +x build/bin/openai-router-* 2>/dev/null

echo ""
echo "========================================"
echo "  Build Complete!"
echo "========================================"
echo ""
echo "Desktop Application:"
echo "  build/bin/OpenAI Router (Wails GUI)"
echo ""
echo "CLI Binaries:"
echo "  build/bin/openai-router-windows-amd64.exe"
echo "  build/bin/openai-router-linux-amd64"
echo "  build/bin/openai-router-darwin-amd64"
echo "  build/bin/openai-router-darwin-arm64"
echo ""
echo "To start the application:"
echo "  Linux/macOS: ./start.sh"
echo "  Windows: start.bat"
echo "========================================"
