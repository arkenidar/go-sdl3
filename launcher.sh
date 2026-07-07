#!/bin/bash

# Run this from anywhere to start the Launcher on Linux.
# Build with: go build -o bin/launcher ./cmd/launcher
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR" || exit 1
env LD_LIBRARY_PATH="$DIR/lib" "$DIR/bin/launcher"
