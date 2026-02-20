@echo off
setlocal

REM =========================
REM Build for Windows 64-bit
set APP_NAME=Futu
set OUT=%APP_NAME%.exe

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

echo Building %APP_NAME% ...

go build -v -x ^
  -ldflags "-s -w -H=windowsgui" ^
  -o "%OUT%"

IF ERRORLEVEL 1 (
    echo.
    echo Build failed
    exit /b 1
)

echo.
echo Build success: %OUT%
echo.

endlocal
