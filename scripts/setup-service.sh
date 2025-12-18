#!/bin/bash

set -e

echo "Setting up Arabella API Service..."

PROJECT_DIR="/var/www/arabella/backend"
SERVICE_FILE="arabella-api.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_FILE}"

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then 
    echo "This script needs to be run with sudo privileges"
    echo "Please run: sudo bash $0"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing Go..."
    
    # Install Go 1.22
    GO_VERSION="1.22.0"
    ARCH="amd64"
    
    cd /tmp
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz
    tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz
    rm go${GO_VERSION}.linux-${ARCH}.tar.gz
    
    # Add Go to PATH for current session
    export PATH=$PATH:/usr/local/go/bin
    
    # Add Go to PATH permanently
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi
    
    echo "Go installed successfully"
fi

# Add Go to PATH if not already there
export PATH=$PATH:/usr/local/go/bin

# Navigate to project directory
cd "$PROJECT_DIR"

# Install dependencies
echo "Installing Go dependencies..."
go mod download
go mod tidy

# Build the application
echo "Building the application..."
mkdir -p bin
go build -ldflags "-X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" -o bin/api ./cmd/api

if [ ! -f "bin/api" ]; then
    echo "Error: Build failed. Binary not found."
    exit 1
fi

echo "Build successful!"

# Make binary executable
chmod +x bin/api

# Copy service file
echo "Installing systemd service..."
cp "$SERVICE_FILE" "$SERVICE_PATH"

# Set proper permissions
chmod 644 "$SERVICE_PATH"

# Reload systemd
systemctl daemon-reload

# Enable service to start on boot
systemctl enable "$SERVICE_FILE"

echo ""
echo "Service installed successfully!"
echo ""
echo "To start the service, run:"
echo "  sudo systemctl start arabella-api"
echo ""
echo "To check status:"
echo "  sudo systemctl status arabella-api"
echo ""
echo "To view logs:"
echo "  sudo journalctl -u arabella-api -f"
echo ""


