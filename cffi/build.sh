#!/bin/bash

# Builds the shared libgui.so (+ libgui.h) used by every FFI example under
# examples/ (lua-use, python-use). Run this after changing any cffi/*.go
# file; the .so is gitignored and regenerated here, not committed.
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR/.." || exit 1
go build -buildmode=c-shared -o cffi/libgui.so ./cffi
