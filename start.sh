#!/bin/bash

# This script sets up the environment and runs the Go SDL3 application
# Build with: go build -o bin/app ./cmd/counter-demo
env LD_LIBRARY_PATH=$(pwd)/lib bin/app
