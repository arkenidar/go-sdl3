#!/bin/bash

# This script sets up the environment and runs the Go SDL3 application
env LD_LIBRARY_PATH=$(pwd)/lib bin/app
