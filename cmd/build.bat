@echo off
setlocal

REM =========================
REM Package for Windows (reads metadata from FyneApp.toml)
cd /d "%~dp0.."
if errorlevel 1 (
    echo.
    echo Failed to switch to project root
    exit /b 1
)

echo Packaging app with fyne ...

fyne package -os windows -release

@REM go build -v -x ^
@REM   -ldflags "-s -w -H=windowsgui" ^
@REM   -o "%OUT%"
IF ERRORLEVEL 1 (
    echo.
    echo Package failed
    exit /b 1
)

echo.
echo Package success
echo.

endlocal
