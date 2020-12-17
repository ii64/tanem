@ECHO off
go build -i -v -ldflags "-X github.com/ii64/tanem/cmd.GIT_COMMIT=local" tanem.go