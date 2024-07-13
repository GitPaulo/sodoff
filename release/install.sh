#!/bin/bash

PROJECT_NAME="sodoff"
VERSION="1.0.0"

# Determine the OS
OS=$(uname -s)
ARCH=$(uname -m)

# Set the install directory
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# Set the download URL based on the OS
if [ "$OS" = "Linux" ]; then
    URL="https://github.com/GitPaulo/$PROJECT_NAME/releases/download/$VERSION/${PROJECT_NAME}-linux"
elif [ "$OS" = "Darwin" ]; then
    URL="https://github.com/GitPaulo/$PROJECT_NAME/releases/download/$VERSION/${PROJECT_NAME}-macos"
elif [[ "$OS" = "MINGW32_NT" ]] || [[ "$OS" = "MINGW64_NT" ]]; then
    URL="https://github.com/GitPaulo/$PROJECT_NAME/releases/download/$VERSION/${PROJECT_NAME}-windows.exe"
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

# Set the target path for the download
if [ "$OS" = "Linux" ] || [ "$OS" = "Darwin" ]; then
    TARGET="$INSTALL_DIR/$PROJECT_NAME"
elif [[ "$OS" = "MINGW32_NT" ]] || [[ "$OS" = "MINGW64_NT" ]]; then
    TARGET="$INSTALL_DIR/$PROJECT_NAME.exe"
fi

# Download the binary
echo "Downloading $PROJECT_NAME from $URL..."
curl -L "$URL" -o "$TARGET"

# Make it executable
chmod +x "$TARGET"

# Add the install directory to PATH if it's not already there
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "export PATH=\$PATH:$INSTALL_DIR" >> "$HOME/.bashrc"
    echo "Added $INSTALL_DIR to PATH. Please restart your terminal or run 'source ~/.bashrc' to apply the changes."
fi

echo "$PROJECT_NAME installed successfully to $TARGET!"
