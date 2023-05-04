#!/bin/sh
export GOOS=windows
export GOARCH=amd64
go build -ldflags="-s -w" -o bin/RediDB-64.exe main.go

export GOOS=windows
export GOARCH=386
go build -o bin/RediDB-32.exe main.go

export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w" -o bin/RediDB-64 main.go

export GOOS=linux
export GOARCH=386
go build -ldflags="-s -w" -o bin/RediDB-32 main.go

