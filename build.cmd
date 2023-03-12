SET GOOS=windows
SET GOARCH=amd64
go build -ldflags="-s -w" -o bin/RediDB-64.exe main.go

SET GOOS=windows
SET GOARCH=386
go build -o bin/RediDB-32.exe main.go

SET GOOS=linux
SET GOARCH=amd64
go build -ldflags="-s -w" -o bin/RediDB-64 main.go

SET GOOS=linux
SET GOARCH=386
go build -ldflags="-s -w" -o bin/RediDB-32 main.go