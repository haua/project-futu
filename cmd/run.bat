@echo off
setlocal

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

echo Running Futu (dev mode)...

go run -tags software .

endlocal
