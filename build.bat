@echo off
setlocal

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=dev"

set "APP_NAME=gen-handler"
set "BUILD_DIR=dist"
set "PLATFORM_DIR=%BUILD_DIR%\windows_amd64"
set "OUTPUT_PATH=%PLATFORM_DIR%\%APP_NAME%.exe"
set "ARCHIVE_PATH=%BUILD_DIR%\%APP_NAME%_%VERSION%_windows_amd64.zip"

echo [INFO] Building %APP_NAME% %VERSION%

if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%PLATFORM_DIR%"
if errorlevel 1 exit /b 1

go build -ldflags "-X main.version=%VERSION%" -o "%OUTPUT_PATH%" .
if errorlevel 1 exit /b 1

powershell -NoProfile -Command "Compress-Archive -Path '%OUTPUT_PATH%' -DestinationPath '%ARCHIVE_PATH%' -Force"
if errorlevel 1 exit /b 1

echo [INFO] Created %ARCHIVE_PATH%
exit /b 0
