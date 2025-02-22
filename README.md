brackeysGameJam  
https://itch.io/jam/brackeys-13  
theme: 
Nothing can go wrong...

build & run:

for linux, just build.
- remember to enable "Executable as Program" option for the executable.

for windows, do ..  
export PATH="/home/gwk/go/go1.24.0/bin:$PATH"
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags "-s -w"
