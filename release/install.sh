#!/bin/bash

PROJECT_NAME="sodoff"
VERSION="1.0.0"

# Determine the OS
OS=$(uname -s)
ARCH=$(uname -m)

# Set the download URL based on the OS
if [ "$OS" = "Linux" ]; then
    URL="https://github.com/GitPaulo/$PROJECT_NAME/releases/download/v$VERSION/${PROJECT_NAME}-linux"
elif [ "$OS" = "Darwin" ]; then
    URL="https://github.com/GitPaulo/$PROJECT_NAME/releases/download/v$VERSION/${PROJECT_NAME}-macos"
elif [ "$OS" = "MINGW32_NT" ] || [ "$OS" = "MINGW64_NT" ]; then
    URL="https://github.com/GitPaulo/$PROJECT_NAME/releases/download/v$VERSION/${PROJECT_NAME}-windows.exe"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

# Download the binary
echo "Downloading $PROJECT_NAME from $URL..."
curl -L $URL -o /usr/local/bin/$PROJECT_NAME

# Make it executable
chmod +x /usr/local/bin/$PROJECT_NAME

echo "$PROJECT_NAME installed successfully!"
