@rem env PATH="$PATH:$(pwd)/exe" go run app.go

@set PATH=%PATH%;%CD%
@cd ..
@%CD%\exe\app.exe

@rem @pause