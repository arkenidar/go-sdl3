@rem env PATH="$PATH:$(pwd)/exe" go run ./cmd/counter-demo

@set PATH=%PATH%;%CD%
@cd ..
@%CD%\exe\app.exe

@rem @pause