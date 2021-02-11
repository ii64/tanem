@ECHO off
go build -i -v -ldflags "-X main.GIT_COMMIT=local" tanem.go