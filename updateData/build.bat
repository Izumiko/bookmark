@echo off
rem set GOOS=linux
rem set GOARCH=amd64
go build -ldflags "-s -w"

