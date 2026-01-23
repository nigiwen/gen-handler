@echo off
REM gen-handler Windows 编译脚本

setlocal enabledelayedexpansion

REM 版本号（从 git tag 获取，或手动指定）
set VERSION=%1
if "%VERSION%"=="" (
    for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
    if "!VERSION!"=="" set VERSION=dev
)

set APP_NAME=gen-handler
set BUILD_DIR=dist

echo 🚀 开始编译 %APP_NAME% v%VERSION%
echo.

REM 清理旧的构建文件
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
mkdir %BUILD_DIR%

REM 编译 Windows amd64
echo 📦 编译 windows/amd64...
set GOOS=windows
set GOARCH=amd64
set OUTPUT_PATH=%BUILD_DIR%\windows_amd64\%APP_NAME%.exe

go build -ldflags "-X main.version=%VERSION%" -o "%OUTPUT_PATH%" .

REM 创建压缩包
cd %BUILD_DIR%\windows_amd64
powershell -Command "Compress-Archive -Path %APP_NAME%.exe -DestinationPath ..\%APP_NAME%_%VERSION%_windows_amd64.zip -Force"
cd ..\..

echo ✅ windows/amd64 编译完成
echo.
echo 🎉 编译完成！
echo.
echo 📦 构建产物: %BUILD_DIR%\gen-handler_%VERSION%_windows_amd64.zip
echo.
