# Create a build directory
mkdir build

# Build for Linux
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o build\sodoff-linux

# Build for macOS
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o build\sodoff-macos

# Build for Windows
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o build\sodoff-windows.exe

Write-Output "Builds completed successfully!"

