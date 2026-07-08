#!/bin/bash

# Builds the shared libgui.{so,dylib,dll} (+ header) used by every FFI
# example under examples/ (lua-use, python-use) and by python/pygosdl.
# Run this after changing any cffi/*.go file; the library is gitignored
# and regenerated here, not committed.
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR/.." || exit 1

case "$(go env GOOS)" in
    windows) ext=dll ;;
    darwin) ext=dylib ;;
    *) ext=so ;;
esac

go build -buildmode=c-shared -o "cffi/libgui.$ext" ./cffi
