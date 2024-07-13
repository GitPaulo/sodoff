#!/bin/bash

# Set the project name
PROJECT_NAME="sodoff"

# Create a build directory
mkdir -p build

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o build/${PROJECT_NAME}-linux

# Build for macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o build/${PROJECT_NAME}-macos

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o build/${PROJECT_NAME}-windows.exe

echo "Build completed!"
