# Linux
go build -ldflags="-s -w" -o build/

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/